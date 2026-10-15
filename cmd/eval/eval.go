package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jeejak/parcel-proof/evidencecheck"
)

// Label is one human-labelled record from labels.csv.
type Label struct {
	RecordID     string
	Resi         string
	Koli         int
	Courier      string
	Expected     evidencecheck.Observation // what a careful human sees
	HumanVerdict string                    // kuat | perlu_dicek | lemah
}

var labelHeader = []string{
	"record_id", "resi", "koli", "courier",
	"image_usable", "label_visible", "label_resi_read",
	"package_sealed", "item_visible", "koli_seen", "human_verdict",
}

// ReadLabels parses labels.csv. Empty label_resi_read / koli_seen mean
// null. Every enum is validated so a typo in the grading sheet fails
// loudly instead of silently counting as a model error.
func ReadLabels(r io.Reader) ([]Label, error) {
	cr := csv.NewReader(r)
	cr.TrimLeadingSpace = true
	head, err := cr.Read()
	if err != nil {
		return nil, fmt.Errorf("labels: header: %w", err)
	}
	if strings.Join(head, ",") != strings.Join(labelHeader, ",") {
		return nil, fmt.Errorf("labels: header must be %s", strings.Join(labelHeader, ","))
	}
	var out []Label
	seen := map[string]bool{}
	for line := 2; ; line++ {
		row, err := cr.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("labels: line %d: %w", line, err)
		}
		l, err := parseLabel(row)
		if err != nil {
			return nil, fmt.Errorf("labels: line %d: %w", line, err)
		}
		if seen[l.RecordID] {
			return nil, fmt.Errorf("labels: line %d: duplicate record_id %q", line, l.RecordID)
		}
		seen[l.RecordID] = true
		out = append(out, l)
	}
	return out, nil
}

func parseLabel(row []string) (Label, error) {
	get := func(i int) string { return strings.TrimSpace(row[i]) }
	koli, err := strconv.Atoi(get(2))
	if err != nil || koli < 1 {
		return Label{}, fmt.Errorf("koli %q must be an integer >= 1", get(2))
	}
	obs := evidencecheck.Observation{
		ImageUsable:   get(4),
		LabelVisible:  get(5),
		PackageSealed: get(7),
		ItemVisible:   get(8),
	}
	if v := get(6); v != "" {
		obs.LabelResiRead = &v
	}
	if v := get(9); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return Label{}, fmt.Errorf("koli_seen %q", v)
		}
		obs.KoliSeen = &n
	}
	// Reuse the model-side validator for the human labels.
	raw := fmt.Sprintf(`{"image_usable":%q,"label_visible":%q,"label_resi_read":%s,"package_sealed":%q,"item_visible":%q,"koli_seen":%s}`,
		obs.ImageUsable, obs.LabelVisible, jsonStrPtr(obs.LabelResiRead), obs.PackageSealed, obs.ItemVisible, jsonIntPtr(obs.KoliSeen))
	if _, err := evidencecheck.ParseObservation([]byte(raw)); err != nil {
		return Label{}, err
	}
	v := get(10)
	if v != evidencecheck.VerdictKuat && v != evidencecheck.VerdictPerluDicek && v != evidencecheck.VerdictLemah {
		return Label{}, fmt.Errorf("human_verdict %q", v)
	}
	return Label{RecordID: get(0), Resi: get(1), Koli: koli, Courier: get(3), Expected: obs, HumanVerdict: v}, nil
}

func jsonStrPtr(p *string) string {
	if p == nil {
		return "null"
	}
	return strconv.Quote(*p)
}

func jsonIntPtr(p *int) string {
	if p == nil {
		return "null"
	}
	return strconv.Itoa(*p)
}

// Outcome is one record's eval result.
type Outcome struct {
	Label     Label
	Obs       evidencecheck.Observation
	Result    evidencecheck.Result
	Usage     evidencecheck.Usage
	SchemaErr bool  // response discarded per D3
	Err       error // any other failure (network, auth, unreadable photo)
}

// D8 targets, from Plan 0016.
const (
	targetLemahPrecision = 0.90
	targetLemahRecall    = 0.90
	targetUnsureRate     = 0.15
	minRecords           = 60
	minHumanLemah        = 30
)

// Report is everything the eval prints.
type Report struct {
	Records, Checked, SchemaFailures, OtherErrors int

	// Field agreement over checked records: field -> "expected→got" -> count.
	Fields map[string]map[string]int
	// Verdict confusion over checked records: "human→model" -> count.
	Verdicts map[string]int

	ModelLemah, ModelLemahCorrect int
	HumanLemah                    int
	// HumanLemahAnswered is HumanLemah minus records lost to transport
	// errors: the recall denominator.
	HumanLemahAnswered int
	UnsureOrSchema     int

	InputTokens, OutputTokens, ThoughtTokens int
	Latencies                                []time.Duration
}

// Score builds the report from outcomes. Pure.
func Score(outs []Outcome) Report {
	rep := Report{Fields: map[string]map[string]int{}, Verdicts: map[string]int{}}
	add := func(field, exp, got string) {
		if rep.Fields[field] == nil {
			rep.Fields[field] = map[string]int{}
		}
		rep.Fields[field][exp+"→"+got]++
	}
	for _, o := range outs {
		rep.Records++
		if o.Label.HumanVerdict == evidencecheck.VerdictLemah {
			rep.HumanLemah++
		}
		rep.InputTokens += o.Usage.InputTokens
		rep.OutputTokens += o.Usage.OutputTokens
		rep.ThoughtTokens += o.Usage.ThoughtTokens
		if o.Usage.Latency > 0 {
			rep.Latencies = append(rep.Latencies, o.Usage.Latency)
		}
		if o.Label.HumanVerdict == evidencecheck.VerdictLemah && o.Err == nil {
			rep.HumanLemahAnswered++
		}
		switch {
		case o.SchemaErr:
			rep.SchemaFailures++
			rep.UnsureOrSchema++
			continue
		case o.Err != nil:
			rep.OtherErrors++
			continue
		}
		rep.Checked++
		if o.Obs.HasUnsure() {
			rep.UnsureOrSchema++
		}
		e := o.Label.Expected
		add("image_usable", e.ImageUsable, o.Obs.ImageUsable)
		add("label_visible", e.LabelVisible, o.Obs.LabelVisible)
		add("label_resi_read", resiCell(e.LabelResiRead, o.Label.Resi), resiCell(o.Obs.LabelResiRead, o.Label.Resi))
		add("package_sealed", e.PackageSealed, o.Obs.PackageSealed)
		add("item_visible", e.ItemVisible, o.Obs.ItemVisible)
		add("koli_seen", intCell(e.KoliSeen), intCell(o.Obs.KoliSeen))

		rep.Verdicts[o.Label.HumanVerdict+"→"+o.Result.Verdict]++
		if o.Result.Verdict == evidencecheck.VerdictLemah {
			rep.ModelLemah++
			if o.Label.HumanVerdict == evidencecheck.VerdictLemah {
				rep.ModelLemahCorrect++
			}
		}
	}
	return rep
}

// resiCell buckets a read so the confusion table stays small: whether
// it matched the record, was some other string, or was null.
func resiCell(read *string, resi string) string {
	if read == nil || evidencecheck.NormaliseResi(*read) == "" {
		return "null"
	}
	if evidencecheck.NormaliseResi(*read) == evidencecheck.NormaliseResi(resi) {
		return "match"
	}
	return "other"
}

func intCell(p *int) string {
	if p == nil {
		return "null"
	}
	return strconv.Itoa(*p)
}

// LemahPrecision is the D8 headline number; ok is false with no lemah
// verdicts to measure.
func (r Report) LemahPrecision() (float64, bool) {
	if r.ModelLemah == 0 {
		return 0, false
	}
	return float64(r.ModelLemahCorrect) / float64(r.ModelLemah), true
}

// LemahRecall is the share of answered human-lemah records the model
// also called lemah. A schema failure on a weak record counts as a
// miss, since the seller gets no warning; a record lost to a transport
// error is left out, since the model never judged it. ok is false when
// no human-lemah record was answered.
func (r Report) LemahRecall() (float64, bool) {
	if r.HumanLemahAnswered == 0 {
		return 0, false
	}
	return float64(r.ModelLemahCorrect) / float64(r.HumanLemahAnswered), true
}

// Answered is the number of records the model actually responded to:
// a parsed observation or a schema failure. Records lost to a network
// error, a timeout or an unreadable photo are NOT in it, because the
// model never judged them.
//
// This is the denominator for every rate the eval reports, including
// the token and cost averages.
func (r Report) Answered() int {
	return r.Checked + r.SchemaFailures
}

// UnsureRate is the share of ANSWERED records that came back unsure on
// any field or failed the schema.
//
// The denominator deliberately excludes OtherErrors. Dividing by every
// record would let transport failures pad the denominator and drag the
// rate down, which is the optimistic direction for a D8 gate: 8 unsure
// out of 60 records with 10 network errors reads as 13% and passes,
// when the model was unsure on 8 of the 50 it actually answered, which
// is 16% and fails.
func (r Report) UnsureRate() float64 {
	if r.Answered() == 0 {
		return 0
	}
	return float64(r.UnsureOrSchema) / float64(r.Answered())
}

// Percentile returns the p-th percentile latency (nearest rank).
func (r Report) Percentile(p float64) time.Duration {
	if len(r.Latencies) == 0 {
		return 0
	}
	l := append([]time.Duration(nil), r.Latencies...)
	sort.Slice(l, func(i, j int) bool { return l[i] < l[j] })
	idx := int(p*float64(len(l))+0.999999) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(l) {
		idx = len(l) - 1
	}
	return l[idx]
}

// Print writes the human-readable report. Prices are per million
// tokens in USD; zero means "not given" and cost is not computed, since
// the plan records measured cost and never an assumed one.
func (r Report) Print(w io.Writer, usdPerMIn, usdPerMOut float64) {
	fmt.Fprintf(w, "Records: %d  checked: %d  schema failures: %d  other errors: %d\n\n",
		r.Records, r.Checked, r.SchemaFailures, r.OtherErrors)

	fields := make([]string, 0, len(r.Fields))
	for f := range r.Fields {
		fields = append(fields, f)
	}
	sort.Strings(fields)
	for _, f := range fields {
		cells := r.Fields[f]
		agree, total := 0, 0
		keys := make([]string, 0, len(cells))
		for k, n := range cells {
			keys = append(keys, k)
			total += n
			if p := strings.SplitN(k, "→", 2); p[0] == p[1] {
				agree += n
			}
		}
		sort.Strings(keys)
		fmt.Fprintf(w, "%s: %d/%d agree (%.0f%%)\n", f, agree, total, pct(agree, total))
		for _, k := range keys {
			fmt.Fprintf(w, "    %-24s %d\n", k, cells[k])
		}
	}

	fmt.Fprintf(w, "\nVerdict (human→model):\n")
	vk := make([]string, 0, len(r.Verdicts))
	for k := range r.Verdicts {
		vk = append(vk, k)
	}
	sort.Strings(vk)
	for _, k := range vk {
		fmt.Fprintf(w, "    %-24s %d\n", k, r.Verdicts[k])
	}

	fmt.Fprintf(w, "\nD8 targets\n")
	if p, ok := r.LemahPrecision(); ok {
		fmt.Fprintf(w, "  lemah precision:          %.1f%% (%d/%d)  target >= %.0f%%  %s\n",
			p*100, r.ModelLemahCorrect, r.ModelLemah, targetLemahPrecision*100, pass(p >= targetLemahPrecision))
	} else {
		fmt.Fprintf(w, "  lemah precision:          n/a (model returned no lemah verdicts)  FAIL\n")
	}
	if rc, ok := r.LemahRecall(); ok {
		fmt.Fprintf(w, "  lemah recall:             %.1f%% (%d/%d)  target >= %.0f%%  %s\n",
			rc*100, r.ModelLemahCorrect, r.HumanLemahAnswered, targetLemahRecall*100, pass(rc >= targetLemahRecall))
	} else {
		fmt.Fprintf(w, "  lemah recall:             n/a (no human-lemah record was answered)  FAIL\n")
	}
	if r.Answered() > 0 {
		ur := r.UnsureRate()
		fmt.Fprintf(w, "  unsure + schema failures: %.1f%% (%d/%d answered)  target < %.0f%%  %s\n",
			ur*100, r.UnsureOrSchema, r.Answered(), targetUnsureRate*100, pass(ur < targetUnsureRate))
	} else {
		// 0 of 0 is not "under 15%": a run where every request failed has
		// measured nothing, so the gate cannot pass.
		fmt.Fprintf(w, "  unsure + schema failures: n/a (no record was answered)  FAIL\n")
	}
	fmt.Fprintf(w, "  eval set size:            %d records, %d human-lemah  need >= %d and >= %d  %s\n",
		r.Records, r.HumanLemah, minRecords, minHumanLemah, pass(r.Records >= minRecords && r.HumanLemah >= minHumanLemah))

	fmt.Fprintf(w, "\nUsage\n")
	n := r.Answered()
	if n > 0 {
		fmt.Fprintf(w, "  tokens: %d in, %d out, of which %d reasoning  (per record: %.0f in, %.0f out)\n",
			r.InputTokens, r.OutputTokens, r.ThoughtTokens, float64(r.InputTokens)/float64(n), float64(r.OutputTokens)/float64(n))
	}
	fmt.Fprintf(w, "  latency: p50 %s  p95 %s\n", r.Percentile(0.50).Round(time.Millisecond), r.Percentile(0.95).Round(time.Millisecond))
	if usdPerMIn > 0 || usdPerMOut > 0 {
		cost := float64(r.InputTokens)/1e6*usdPerMIn + float64(r.OutputTokens)/1e6*usdPerMOut
		per := 0.0
		if n > 0 {
			per = cost / float64(n)
		}
		fmt.Fprintf(w, "  cost: $%.4f total, $%.6f per record (at the prices passed in)\n", cost, per)
	} else {
		fmt.Fprintf(w, "  cost: not computed; pass -usd-per-mtok-in and -usd-per-mtok-out from the current Vertex price list\n")
	}
}

func pct(a, b int) float64 {
	if b == 0 {
		return 0
	}
	return float64(a) * 100 / float64(b)
}

func pass(ok bool) string {
	if ok {
		return "PASS"
	}
	return "FAIL"
}
