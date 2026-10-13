package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jeejak/parcel-proof/evidencecheck"
)

const header = "record_id,resi,koli,courier,image_usable,label_visible,label_resi_read,package_sealed,item_visible,koli_seen,human_verdict\n"

func TestReadLabels(t *testing.T) {
	ok := header + "r1,SPX1,1,SPX,ok,yes,SPX1,yes,yes,1,kuat\nr2,JNT2,2,J&T,dark,no,,unsure,no,,lemah\n"
	ls, err := ReadLabels(strings.NewReader(ok))
	if err != nil {
		t.Fatal(err)
	}
	if len(ls) != 2 || ls[1].Expected.LabelResiRead != nil || ls[1].Expected.KoliSeen != nil || ls[1].Koli != 2 {
		t.Fatalf("parsed %+v", ls)
	}

	bad := map[string]string{
		"wrong header":  "a,b\n",
		"bad enum":      header + "r1,SPX1,1,SPX,fine,yes,SPX1,yes,yes,1,kuat\n",
		"bad verdict":   header + "r1,SPX1,1,SPX,ok,yes,SPX1,yes,yes,1,good\n",
		"zero koli":     header + "r1,SPX1,0,SPX,ok,yes,SPX1,yes,yes,1,kuat\n",
		"duplicate":     header + "r1,SPX1,1,SPX,ok,yes,SPX1,yes,yes,1,kuat\nr1,SPX1,1,SPX,ok,yes,SPX1,yes,yes,1,kuat\n",
		"bad koli_seen": header + "r1,SPX1,1,SPX,ok,yes,SPX1,yes,yes,x,kuat\n",
	}
	for name, in := range bad {
		t.Run(name, func(t *testing.T) {
			if _, err := ReadLabels(strings.NewReader(in)); err == nil {
				t.Fatal("accepted")
			}
		})
	}
}

func sp(s string) *string { return &s }

func TestScore(t *testing.T) {
	obs := func(usable string) evidencecheck.Observation {
		return evidencecheck.Observation{ImageUsable: usable, LabelVisible: "yes", LabelResiRead: sp("SPX1"), PackageSealed: "yes", ItemVisible: "yes"}
	}
	lab := func(id, human string) Label {
		return Label{RecordID: id, Resi: "SPX1", Koli: 1, Expected: obs("ok"), HumanVerdict: human}
	}
	mk := func(l Label, o evidencecheck.Observation) Outcome {
		return Outcome{Label: l, Obs: o, Result: evidencecheck.Decide(evidencecheck.Input{Obs: o, Resi: l.Resi, Koli: l.Koli}),
			Usage: evidencecheck.Usage{InputTokens: 100, OutputTokens: 10, Latency: time.Second}}
	}
	unsure := obs("ok")
	unsure.PackageSealed = "unsure"

	outs := []Outcome{
		mk(lab("a", "kuat"), obs("ok")),
		mk(lab("b", "lemah"), obs("dark")),  // correct lemah
		mk(lab("c", "kuat"), obs("blurry")), // wrong lemah
		mk(lab("d", "lemah"), obs("dark")),  // correct lemah
		mk(lab("e", "kuat"), unsure),        // unsure, still kuat
		{Label: lab("f", "kuat"), SchemaErr: true, Usage: evidencecheck.Usage{InputTokens: 100}},
		{Label: lab("g", "kuat"), Err: fmt.Errorf("network")},
	}
	r := Score(outs)

	if r.Records != 7 || r.Checked != 5 || r.SchemaFailures != 1 || r.OtherErrors != 1 {
		t.Fatalf("counts %+v", r)
	}
	if p, ok := r.LemahPrecision(); !ok || r.ModelLemah != 3 || r.ModelLemahCorrect != 2 || p < 0.66 || p > 0.67 {
		t.Fatalf("lemah precision %v (%d/%d)", p, r.ModelLemahCorrect, r.ModelLemah)
	}
	if r.UnsureOrSchema != 2 {
		t.Fatalf("unsure+schema = %d, want 2", r.UnsureOrSchema)
	}
	// The rate divides by ANSWERED records (6: five checked plus one
	// schema failure), not by all 7. Record "g" failed with a network
	// error, so the model never judged it; counting it would pad the
	// denominator and drag a D8 gate number down. 2/6 = 33.3%, not
	// 2/7 = 28.6%.
	if r.Answered() != 6 {
		t.Fatalf("answered = %d, want 6", r.Answered())
	}
	if got := r.UnsureRate(); got < 0.333 || got > 0.334 {
		t.Fatalf("unsure rate = %.4f, want 2/6; dividing by all records would give %.4f", got, 2.0/7.0)
	}
	if r.HumanLemah != 2 {
		t.Fatalf("human lemah = %d", r.HumanLemah)
	}
	if r.Fields["image_usable"]["ok→dark"] != 2 || r.Fields["label_resi_read"]["match→match"] != 5 {
		t.Fatalf("fields %+v", r.Fields)
	}
	if r.InputTokens != 600 || r.Percentile(0.5) != time.Second {
		t.Fatalf("usage %d %s", r.InputTokens, r.Percentile(0.5))
	}

	var buf bytes.Buffer
	r.Print(&buf, 0, 0)
	s := buf.String()
	for _, want := range []string{"lemah precision:", "FAIL", "cost: not computed"} {
		if !strings.Contains(s, want) {
			t.Errorf("report missing %q:\n%s", want, s)
		}
	}
	buf.Reset()
	r.Print(&buf, 1, 2)
	if !strings.Contains(buf.String(), "cost: $") {
		t.Error("cost line missing when prices are given")
	}
}

func TestRun_DryRunAgreesWithLabels(t *testing.T) {
	dir := t.TempDir()
	csv := header +
		"r1,SPX1,1,SPX,ok,yes,SPX1,yes,yes,1,kuat\n" +
		"r2,JNT2,1,J&T,dark,no,,no,no,,lemah\n"
	if err := os.WriteFile(filepath.Join(dir, "labels.csv"), []byte(csv), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"r1", "r2"} {
		if err := os.Mkdir(filepath.Join(dir, id), 0o755); err != nil {
			t.Fatal(err)
		}
		for i := 1; i <= 6; i++ {
			var b bytes.Buffer
			_ = jpeg.Encode(&b, image.NewRGBA(image.Rect(0, 0, 1600, 1200)), nil)
			if err := os.WriteFile(filepath.Join(dir, id, fmt.Sprintf("%02d.jpg", i)), b.Bytes(), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}

	var out bytes.Buffer
	if err := run([]string{"-dir", dir, "-dry-run", "-out", filepath.Join(dir, "res.jsonl")}, &out); err != nil {
		t.Fatal(err)
	}
	s := out.String()
	for _, want := range []string{
		"checked: 2",
		"image_usable: 2/2 agree",
		"label_resi_read: 2/2 agree",
		"lemah precision:          100.0% (1/1)",
		"kuat→kuat",
		"lemah→lemah",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q in:\n%s", want, s)
		}
	}
	res, err := os.ReadFile(filepath.Join(dir, "res.jsonl"))
	if err != nil || strings.Count(string(res), "\n") != 2 {
		t.Fatalf("results file: %v %q", err, res)
	}
}

// A record that fails before the checker runs (here: its folder is
// missing) must not shift the echoed labels onto the records after it.
// With positional echoing it did, and one missing folder turned every
// later record into a fake disagreement.
func TestRun_DryRunMissingFolderDoesNotShiftLabels(t *testing.T) {
	dir := t.TempDir()
	csv := header +
		"r1,SPX1,1,SPX,ok,yes,SPX1,yes,yes,1,kuat\n" +
		"r2,JNT2,1,J&T,dark,no,,no,no,,lemah\n" +
		"r3,JNE3,2,JNE,ok,yes,JNE3,no,yes,2,perlu_dicek\n"
	if err := os.WriteFile(filepath.Join(dir, "labels.csv"), []byte(csv), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"r1", "r3"} { // r2's folder is missing on purpose
		if err := os.Mkdir(filepath.Join(dir, id), 0o755); err != nil {
			t.Fatal(err)
		}
		var b bytes.Buffer
		_ = jpeg.Encode(&b, image.NewRGBA(image.Rect(0, 0, 64, 48)), nil)
		if err := os.WriteFile(filepath.Join(dir, id, "01.jpg"), b.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	var out bytes.Buffer
	if err := run([]string{"-dir", dir, "-dry-run"}, &out); err != nil {
		t.Fatal(err)
	}
	s := out.String()
	for _, want := range []string{
		"checked: 2",
		"other errors: 1",
		"image_usable: 2/2 agree",
		"package_sealed: 2/2 agree",
		"perlu_dicek→perlu_dicek",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q in:\n%s", want, s)
		}
	}
}

// -base-url runs the real Vertex client against a mock server, without
// Google credentials, so the whole path (request, usage accounting,
// schema checks, scoring) can be exercised before any real call.
func TestRun_BaseURLUsesMockServerWithoutCredentials(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if !strings.HasSuffix(r.URL.Path, "/v1/projects/p/locations/eu/publishers/google/models/m:generateContent") {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "" {
			t.Error("mock request carried credentials")
		}
		obs := `{"image_usable":"ok","label_visible":"yes","label_resi_read":"SPX1","package_sealed":"yes","item_visible":"yes","koli_seen":1}`
		_ = json.NewEncoder(w).Encode(map[string]any{
			"candidates": []any{map[string]any{
				"content":      map[string]any{"role": "model", "parts": []any{map[string]any{"text": obs}}},
				"finishReason": "STOP",
			}},
			"usageMetadata": map[string]any{"promptTokenCount": 1000, "candidatesTokenCount": 50, "thoughtsTokenCount": 150},
		})
	}))
	defer srv.Close()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "labels.csv"), []byte(header+"r1,SPX1,1,SPX,ok,yes,SPX1,yes,yes,1,kuat\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "r1"), 0o755); err != nil {
		t.Fatal(err)
	}
	var b bytes.Buffer
	_ = jpeg.Encode(&b, image.NewRGBA(image.Rect(0, 0, 64, 48)), nil)
	if err := os.WriteFile(filepath.Join(dir, "r1", "01.jpg"), b.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	outFile := filepath.Join(t.TempDir(), "res.jsonl")
	err := run([]string{"-dir", dir, "-base-url", srv.URL, "-project", "p", "-location", "eu", "-model", "m",
		"-usd-per-mtok-in", "1", "-usd-per-mtok-out", "10", "-out", outFile}, &out)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("mock server calls = %d", calls)
	}
	for _, want := range []string{"checked: 1", "kuat→kuat", "200 out, of which 150 reasoning", "MOCK SERVER"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing %q in:\n%s", want, out.String())
		}
	}

	// The -out file is the durable record of a run, so it carries the
	// reasoning share and the prompt that produced each row, not only
	// the console summary.
	raw, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}
	var row map[string]any
	if err := json.Unmarshal(raw, &row); err != nil {
		t.Fatal(err)
	}
	if row["output_tokens"] != float64(200) || row["thought_tokens"] != float64(150) || row["prompt_version"] != evidencecheck.PromptVersion {
		t.Errorf("-out row = %v", row)
	}
}

// mkSet writes a small eval set: n identical kuat records, one photo
// each. Used by the backend and stop-on-refusal tests.
func mkSet(t *testing.T, n int) string {
	t.Helper()
	dir := t.TempDir()
	csv := header
	for i := 1; i <= n; i++ {
		csv += fmt.Sprintf("r%d,SPX1,1,SPX,ok,yes,SPX1,yes,yes,1,kuat\n", i)
		if err := os.Mkdir(filepath.Join(dir, fmt.Sprintf("r%d", i)), 0o755); err != nil {
			t.Fatal(err)
		}
		var b bytes.Buffer
		_ = jpeg.Encode(&b, image.NewRGBA(image.Rect(0, 0, 64, 48)), nil)
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("r%d", i), "01.jpg"), b.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "labels.csv"), []byte(csv), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func kuatReply(w http.ResponseWriter) {
	obs := `{"image_usable":"ok","label_visible":"yes","label_resi_read":"SPX1","package_sealed":"yes","item_visible":"yes","koli_seen":1}`
	_ = json.NewEncoder(w).Encode(map[string]any{
		"candidates": []any{map[string]any{
			"content":      map[string]any{"role": "model", "parts": []any{map[string]any{"text": obs}}},
			"finishReason": "STOP",
		}},
		"usageMetadata": map[string]any{"promptTokenCount": 1000, "candidatesTokenCount": 50},
	})
}

// D11: the key comes only from $GEMINI_API_KEY, never a flag, since
// command-line arguments are visible in the process list.
func TestRun_GeminiAPIRequiresKeyFromEnv(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")
	err := run([]string{"-dir", mkSet(t, 1), "-backend", "gemini-api", "-model", "m"}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "GEMINI_API_KEY") {
		t.Fatalf("err = %v", err)
	}
	if err := run([]string{"-dir", mkSet(t, 1), "-backend", "nope", "-model", "m"}, &bytes.Buffer{}); err == nil {
		t.Fatal("unknown backend accepted")
	}
}

// With -base-url the key stays home even when it is set: a mock gets
// the Gemini request shape and no credential.
func TestRun_GeminiAPIViaMockSendsNoKey(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "real-key-must-not-leave")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1beta/models/m:generateContent" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Header.Get("x-goog-api-key") != "" || r.URL.RawQuery != "" || r.Header.Get("Authorization") != "" {
			t.Error("a credential reached the mock")
		}
		kuatReply(w)
	}))
	defer srv.Close()
	var out bytes.Buffer
	if err := run([]string{"-dir", mkSet(t, 1), "-backend", "gemini-api", "-model", "m", "-base-url", srv.URL}, &out); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"checked: 1", "MOCK SERVER"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing %q in:\n%s", want, out.String())
		}
	}
}

// A quota or permission refusal stops the run: the remaining records
// are not sent, the partial report still prints, and run returns the
// error so the exit code is non-zero.
func TestRun_StopsOnQuota(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":{"code":429,"message":"quota","status":"RESOURCE_EXHAUSTED"}}`))
			return
		}
		kuatReply(w)
	}))
	defer srv.Close()
	var out bytes.Buffer
	err := run([]string{"-dir", mkSet(t, 4), "-base-url", srv.URL, "-project", "p", "-location", "eu", "-model", "m"}, &out)
	if err == nil {
		t.Fatal("run did not report the stop")
	}
	if calls != 2 {
		t.Errorf("calls = %d, want 2 (stop after the refusal)", calls)
	}
	for _, want := range []string{"STOPPED after record 2 of 4", "checked: 1"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing %q in:\n%s", want, out.String())
		}
	}
}

// A run where every request failed (an outage, a bad key) has no
// answered records. The unsure gate must not report a pass on no data:
// 0 of 0 is not "under 15%".
func TestPrint_UnsureGateFailsWithNoAnsweredRecords(t *testing.T) {
	r := Score([]Outcome{{Err: errors.New("http 503")}, {Err: errors.New("http 503")}})
	var buf bytes.Buffer
	r.Print(&buf, 0, 0)
	line := ""
	for _, l := range strings.Split(buf.String(), "\n") {
		if strings.Contains(l, "unsure + schema failures") {
			line = l
		}
	}
	if !strings.Contains(line, "n/a") || !strings.HasSuffix(strings.TrimSpace(line), "FAIL") {
		t.Fatalf("unsure line with nothing answered = %q, want n/a and FAIL", line)
	}
}

func TestRun_RequiresDir(t *testing.T) {
	if err := run(nil, &bytes.Buffer{}); err == nil {
		t.Fatal("ran without -dir")
	}
}

// Recall is caught lemah over human-lemah records the model ANSWERED.
// A schema failure on a weak record is a miss (the seller would get no
// warning); a network error is not in the denominator, because the
// model never judged that record.
func TestScore_LemahRecall(t *testing.T) {
	obs := func(usable string) evidencecheck.Observation {
		return evidencecheck.Observation{ImageUsable: usable, LabelVisible: "yes", LabelResiRead: sp("SPX1"), PackageSealed: "yes", ItemVisible: "yes"}
	}
	lab := func(human string) Label {
		return Label{Resi: "SPX1", Koli: 1, Expected: obs("ok"), HumanVerdict: human}
	}
	mk := func(l Label, o evidencecheck.Observation) Outcome {
		return Outcome{Label: l, Obs: o, Result: evidencecheck.Decide(evidencecheck.Input{Obs: o, Resi: l.Resi, Koli: l.Koli})}
	}
	r := Score([]Outcome{
		mk(lab("lemah"), obs("dark")),                      // caught
		mk(lab("lemah"), obs("dark")),                      // caught
		mk(lab("lemah"), obs("ok")),                        // missed: called kuat
		{Label: lab("lemah"), SchemaErr: true},             // missed: no answer to warn on
		{Label: lab("lemah"), Err: errors.New("http 503")}, // not judged
		mk(lab("kuat"), obs("dark")),                       // false lemah, not in recall
	})
	if r.HumanLemah != 5 || r.HumanLemahAnswered != 4 {
		t.Fatalf("human lemah %d, answered %d; want 5 and 4", r.HumanLemah, r.HumanLemahAnswered)
	}
	got, ok := r.LemahRecall()
	if !ok || got != 0.5 {
		t.Fatalf("recall = %v ok=%v, want 2/4 = 0.5", got, ok)
	}

	var buf bytes.Buffer
	r.Print(&buf, 0, 0)
	if !strings.Contains(buf.String(), "lemah recall:             50.0% (2/4)  target >= 90%  FAIL") {
		t.Fatalf("recall line missing or wrong:\n%s", buf.String())
	}
}

// No answered human-lemah record means recall was not measured, which
// cannot pass the gate.
func TestPrint_RecallGateFailsWithNoAnsweredLemah(t *testing.T) {
	r := Score([]Outcome{{Label: Label{HumanVerdict: "lemah"}, Err: errors.New("http 503")}})
	if _, ok := r.LemahRecall(); ok {
		t.Fatal("recall reported as measured with nothing answered")
	}
	var buf bytes.Buffer
	r.Print(&buf, 0, 0)
	if !strings.Contains(buf.String(), "lemah recall:             n/a (no human-lemah record was answered)  FAIL") {
		t.Fatalf("recall line missing or wrong:\n%s", buf.String())
	}
}

// -media-resolution reaches the request and is printed in the report
// header, since a result at one level is not comparable with another.
// Without the flag the header says "default" and nothing is sent.
func TestRun_MediaResolution(t *testing.T) {
	for _, tc := range []struct {
		flag, wantLevel, wantHeader string
	}{
		{"", "", "Media resolution: default (none sent)"},
		{"medium", "MEDIA_RESOLUTION_MEDIUM", "Media resolution: medium"},
	} {
		var gotLevel string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body struct {
				Contents []struct {
					Parts []struct {
						MediaResolution *struct {
							Level string `json:"level"`
						} `json:"mediaResolution"`
					} `json:"parts"`
				} `json:"contents"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if mr := body.Contents[0].Parts[1].MediaResolution; mr != nil {
				gotLevel = mr.Level
			}
			kuatReply(w)
		}))
		args := []string{"-dir", mkSet(t, 1), "-base-url", srv.URL, "-project", "p", "-location", "eu", "-model", "m"}
		if tc.flag != "" {
			args = append(args, "-media-resolution", tc.flag)
		}
		var out bytes.Buffer
		err := run(args, &out)
		srv.Close()
		if err != nil {
			t.Fatal(err)
		}
		if gotLevel != tc.wantLevel {
			t.Errorf("flag %q: level sent = %q, want %q", tc.flag, gotLevel, tc.wantLevel)
		}
		if !strings.Contains(out.String(), tc.wantHeader) {
			t.Errorf("flag %q: header missing %q in:\n%s", tc.flag, tc.wantHeader, out.String())
		}
	}

	if err := run([]string{"-dir", mkSet(t, 1), "-base-url", "http://127.0.0.1:1", "-project", "p", "-location", "eu", "-model", "m", "-media-resolution", "ultra"}, &bytes.Buffer{}); err == nil {
		t.Fatal("unknown media resolution accepted")
	}
}

// A dry run never builds a client, so the flag is validated before the
// run starts: a report header must not name a level that does not exist.
func TestRun_DryRunRejectsUnknownMediaResolution(t *testing.T) {
	if err := run([]string{"-dir", mkSet(t, 1), "-dry-run", "-media-resolution", "ultra"}, &bytes.Buffer{}); err == nil {
		t.Fatal("dry run accepted an unknown media resolution")
	}
}
