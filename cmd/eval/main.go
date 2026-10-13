// Command evidencecheck-eval runs the Plan 0016 Phase 0 offline eval:
// it sends a folder of staged packing photos through a Checker, scores
// the observations against a human-graded labels file, and prints the
// D8 numbers plus measured tokens, latency and (if prices are given)
// cost.
//
// It never touches production data. The photo set is staged by the
// team with dummy waybills (Plan 0016 §4 Phase 0).
//
// Layout:
//
//	<dir>/labels.csv              see labelHeader in eval.go
//	<dir>/<record_id>/01.jpg ...  photos; file-name order stands in
//	                              for order_index
//
// Usage:
//
//	go run ./cmd/evidencecheck-eval -dir ./eval-set \
//	    -project my-gcp-project -location eu -model <model-id>
//
//	GEMINI_API_KEY=... go run ./cmd/evidencecheck-eval -dir ./eval-set \
//	    -backend gemini-api -model <model-id> -delay 5s
//
// -media-resolution low|medium|high sets how many tokens each photo is
// encoded into, the main cost lever; the report header records it.
//
// Vertex (the default backend) uses Application Default Credentials.
// The Gemini Developer API backend (Plan 0016 D11, staged photos only)
// reads its key from $GEMINI_API_KEY, never from a flag. A key from a
// project with no billing account cannot incur charges.
//
// -dry-run replaces the model with one that echoes the human labels, to
// check the set and the pipeline without spending anything. -base-url
// points either backend at a mock server and never sends it a
// credential. A quota or permission refusal stops the run: the rest of
// the records are not sent, and the partial report says so.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jeejak/parcel-proof/evidencecheck"
	"github.com/jeejak/parcel-proof/gemini"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "evidencecheck-eval:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("evidencecheck-eval", flag.ContinueOnError)
	dir := fs.String("dir", "", "eval set directory (labels.csv + one subdirectory per record)")
	project := fs.String("project", os.Getenv("GEMINI_PROJECT_ID"), "GCP project id (default $GEMINI_PROJECT_ID)")
	location := fs.String("location", os.Getenv("GEMINI_LOCATION"), "Vertex location (default $GEMINI_LOCATION)")
	model := fs.String("model", os.Getenv("GEMINI_MODEL"), "model id (default $GEMINI_MODEL)")
	dryRun := fs.Bool("dry-run", false, "echo the human labels instead of calling the model")
	limit := fs.Int("limit", 0, "check at most this many records (0 = all)")
	outPath := fs.String("out", "", "write per-record results as JSON lines to this file")
	priceIn := fs.Float64("usd-per-mtok-in", 0, "input price, USD per million tokens (for the cost line)")
	priceOut := fs.Float64("usd-per-mtok-out", 0, "output price, USD per million tokens")
	timeout := fs.Duration("timeout", 60*time.Second, "per-record timeout")
	baseURL := fs.String("base-url", "", "send requests to this host instead of Google, without any credential (testing against a mock server only)")
	backend := fs.String("backend", gemini.APIVertex, "vertex, gemini-api (Plan 0016 D11: staged photos only; key from $GEMINI_API_KEY), or basic (no model: dark and blurry from pixel statistics, no cost)")
	basicDark := fs.Float64("basic-dark", 0, "basic: a photo with mean brightness (0-255) below this is dark")
	basicBlur := fs.Float64("basic-blur", 0, "basic: a photo with Laplacian variance below this is blurry")
	basicBarcodes := fs.Bool("basic-barcodes", false, "basic: also read one Code 128 shipping barcode as label_resi_read")
	statsOut := fs.String("stats-out", "", "basic: write per-photo brightness and sharpness as CSV here (for calibration)")
	delay := fs.Duration("delay", 0, "wait this long between records, to stay under a rate limit")
	mediaRes := fs.String("media-resolution", "", "low, medium or high, set on every photo; empty sends nothing and the model default applies")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *dir == "" {
		return errors.New("-dir is required")
	}
	if err := gemini.CheckMediaResolution(*mediaRes); err != nil {
		return err
	}

	f, err := os.Open(filepath.Join(*dir, "labels.csv"))
	if err != nil {
		return err
	}
	labels, err := ReadLabels(f)
	f.Close()
	if err != nil {
		return err
	}
	if *limit > 0 && *limit < len(labels) {
		labels = labels[:*limit]
	}

	ctx := context.Background()
	var checker evidencecheck.Checker
	isBasic := *backend == "basic"
	var statsW io.Writer
	if *statsOut != "" {
		sf, err := os.Create(*statsOut)
		if err != nil {
			return err
		}
		defer sf.Close()
		statsW = sf
		fmt.Fprintln(statsW, "record_id,photo,luma,sharpness,expected_image_usable")
	}
	if !*dryRun && !isBasic {
		cfg := gemini.Config{API: *backend, Project: *project, Location: *location, Model: *model, MediaResolution: *mediaRes}
		switch *backend {
		case gemini.APIVertex:
		case gemini.APIGemini:
			cfg.APIKey = os.Getenv("GEMINI_API_KEY")
			if cfg.APIKey == "" && *baseURL == "" {
				return errors.New("-backend gemini-api reads its key from $GEMINI_API_KEY, which is empty")
			}
		default:
			return fmt.Errorf("-backend %q: want %s or %s", *backend, gemini.APIVertex, gemini.APIGemini)
		}
		if *baseURL != "" {
			// A mock server must never receive a credential of any kind.
			cfg.BaseURL = *baseURL
			cfg.HTTPClient = &http.Client{}
			cfg.APIKey = ""
		}
		c, err := gemini.New(ctx, cfg)
		if err != nil {
			return err
		}
		checker = c
	}

	out := io.Writer(io.Discard)
	if *outPath != "" {
		of, err := os.Create(*outPath)
		if err != nil {
			return err
		}
		defer of.Close()
		out = of
	}

	outcomes := make([]Outcome, 0, len(labels))
	var stopped error
	for i, l := range labels {
		if i > 0 && *delay > 0 && !*dryRun {
			time.Sleep(*delay)
		}
		c := checker
		if isBasic {
			l := l
			bc := evidencecheck.BasicChecker{DarkBelow: *basicDark, BlurBelow: *basicBlur, ReadBarcodes: *basicBarcodes}
			if statsW != nil {
				bc.OnStats = func(i int, st evidencecheck.PhotoStats) {
					fmt.Fprintf(statsW, "%s,%d,%.2f,%.2f,%s\n", l.RecordID, i, st.Luma, st.Sharpness, l.Expected.ImageUsable)
				}
			}
			c = bc
		}
		if *dryRun {
			c = echoChecker(l)
		}
		o := checkOne(ctx, c, *dir, l, *timeout)
		outcomes = append(outcomes, o)
		status := o.Result.Verdict
		switch {
		case o.SchemaErr:
			status = "schema-fail"
		case o.Err != nil:
			status = "error: " + o.Err.Error()
		}
		fmt.Fprintf(os.Stderr, "[%d/%d] %s %s\n", i+1, len(labels), l.RecordID, status)
		// -out is the durable record of a run that costs money. Dropping
		// a write here would lose rows from a measurement that cannot be
		// reproduced without paying for it again, so stop instead: the
		// records already written stay valid and the console summary
		// still prints for what ran.
		if err := json.NewEncoder(out).Encode(resultLine(o)); err != nil {
			return fmt.Errorf("writing -out after %d of %d records: %w", i+1, len(labels), err)
		}
		// The API refused (quota, rate limit, or permissions). Sending the
		// rest would only collect more refusals, so stop and report.
		if errors.Is(o.Err, gemini.ErrQuota) || errors.Is(o.Err, gemini.ErrAuth) {
			stopped = fmt.Errorf("stopped after record %d of %d: %w", i+1, len(labels), o.Err)
			break
		}
	}

	fmt.Fprintf(stdout, "Prompt version: %s\n", evidencecheck.PromptVersion)
	switch {
	case isBasic:
		fmt.Fprintf(stdout, "Model: basic check, no model (dark below %.1f, blurry below %.1f, barcodes %v)\n", *basicDark, *basicBlur, *basicBarcodes)
	case *dryRun:
		fmt.Fprintf(stdout, "Model: dry run (labels echoed)\n")
	case *baseURL != "":
		fmt.Fprintf(stdout, "Model: %s (%s) via MOCK SERVER %s (not a measurement)\n", *model, *backend, *baseURL)
	case *backend == gemini.APIGemini:
		fmt.Fprintf(stdout, "Model: %s via Gemini Developer API (D11: staged photos only; free-tier inputs may be used by Google)\n", *model)
	default:
		fmt.Fprintf(stdout, "Model: %s @ %s (Vertex)\n", *model, *location)
	}
	if *mediaRes == "" {
		fmt.Fprintf(stdout, "Media resolution: default (none sent)\n")
	} else {
		fmt.Fprintf(stdout, "Media resolution: %s\n", *mediaRes)
	}
	if stopped != nil {
		fmt.Fprintf(stdout, "STOPPED after record %d of %d: the API refused (%v). Partial results below.\n", len(outcomes), len(labels), stopped)
	}
	fmt.Fprintln(stdout)
	Score(outcomes).Print(stdout, *priceIn, *priceOut)
	return stopped
}

// checkOne loads, selects and downscales a record's photos, then runs
// the checker and the server-side decision.
func checkOne(ctx context.Context, c evidencecheck.Checker, dir string, l Label, timeout time.Duration) Outcome {
	o := Outcome{Label: l}
	paths, err := recordPhotos(filepath.Join(dir, l.RecordID))
	if err != nil {
		o.Err = err
		return o
	}
	selected := evidencecheck.SelectPhotos(paths)
	photos := make([][]byte, 0, len(selected))
	for _, p := range selected {
		raw, err := os.ReadFile(p) //nolint:gosec // G304: operator-chosen eval set; this CLI exists to read that directory
		if err != nil {
			o.Err = err
			return o
		}
		small, err := evidencecheck.Downscale(bytes.NewReader(raw), evidencecheck.MaxEdge)
		if err != nil {
			o.Err = fmt.Errorf("%s: %w", filepath.Base(p), err)
			return o
		}
		photos = append(photos, small)
	}

	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	obs, usage, err := c.Check(cctx, evidencecheck.Request{Photos: photos, Koli: l.Koli, Courier: l.Courier})
	o.Usage = usage
	if errors.Is(err, evidencecheck.ErrSchema) {
		o.SchemaErr = true
		o.Err = err
		return o
	}
	if err != nil {
		o.Err = err
		return o
	}
	o.Obs = obs
	o.Result = evidencecheck.Decide(evidencecheck.Input{Obs: obs, Resi: l.Resi, Koli: l.Koli})
	return o
}

// recordPhotos lists a record's JPEG/PNG files in name order.
func recordPhotos(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		switch strings.ToLower(filepath.Ext(e.Name())) {
		case ".jpg", ".jpeg", ".png":
			out = append(out, filepath.Join(dir, e.Name()))
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no photos in %s", dir)
	}
	sort.Strings(out)
	return out, nil
}

// echoChecker returns one record's own human labels as the observation,
// so a dry run must score 100% agreement; anything else is a bug in the
// set or the pipeline. It is built per record rather than handing out
// labels in order: a record that fails before Check (a missing folder,
// an unreadable photo) would otherwise shift every later record's
// labels by one and show up as a wave of fake disagreements.
func echoChecker(l Label) evidencecheck.Checker {
	return evidencecheck.FakeChecker{Fn: func(ctx context.Context, req evidencecheck.Request) (evidencecheck.Observation, evidencecheck.Usage, error) {
		return l.Expected, evidencecheck.Usage{Model: "dry-run"}, nil
	}}
}

type resultJSON struct {
	RecordID     string                     `json:"record_id"`
	HumanVerdict string                     `json:"human_verdict"`
	Verdict      string                     `json:"verdict,omitempty"`
	Reasons      []string                   `json:"reasons,omitempty"`
	Observation  *evidencecheck.Observation `json:"observation,omitempty"`
	SchemaFail   bool                       `json:"schema_fail,omitempty"`
	Error        string                     `json:"error,omitempty"`
	Model        string                     `json:"model,omitempty"`
	InputTokens  int                        `json:"input_tokens"`
	OutputTokens int                        `json:"output_tokens"`
	// ThoughtTokens is the reasoning part of OutputTokens, already
	// included in it. Kept per row because it is the share that moves
	// cost between models (Plan 0016 Q4).
	ThoughtTokens int    `json:"thought_tokens"`
	LatencyMS     int64  `json:"latency_ms"`
	PromptVersion string `json:"prompt_version"`
}

func resultLine(o Outcome) resultJSON {
	r := resultJSON{
		RecordID:      o.Label.RecordID,
		HumanVerdict:  o.Label.HumanVerdict,
		SchemaFail:    o.SchemaErr,
		Model:         o.Usage.Model,
		InputTokens:   o.Usage.InputTokens,
		OutputTokens:  o.Usage.OutputTokens,
		ThoughtTokens: o.Usage.ThoughtTokens,
		LatencyMS:     o.Usage.Latency.Milliseconds(),
		PromptVersion: evidencecheck.PromptVersion,
	}
	if o.Err != nil {
		r.Error = o.Err.Error()
		return r
	}
	obs := o.Obs
	r.Observation = &obs
	r.Verdict = o.Result.Verdict
	r.Reasons = o.Result.Reasons
	return r
}
