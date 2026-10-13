// Package evidencecheck asks a multimodal model what it can see in a
// record's packing photos and turns the answer into a verdict on
// whether that evidence would hold up in a marketplace dispute
// (Plan 0016, G1).
//
// The split is the whole design. The model only observes: it fills a
// fixed set of enum fields (Observation) and nothing else. The server
// decides: Decide is a pure function from those fields plus the
// record's own resi and koli to a verdict and reason codes. The model
// is never told which resi to expect, and the resi comparison happens
// here, so a misread label can only ever produce "label_mismatch",
// never a wrong resi in anything a seller sees (D4).
//
// Nothing in this package does I/O except the Checker implementations,
// so the rules that matter are pinned by table tests that run without
// a network or a database.
package evidencecheck

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// Field values the response schema allows. "unsure" is a real answer
// and never contributes to a warning (D3).
const (
	UsableOK         = "ok"
	UsableDark       = "dark"
	UsableBlurry     = "blurry"
	UsableObstructed = "obstructed"

	Yes    = "yes"
	No     = "no"
	Unsure = "unsure"
)

// Observation is exactly what the model returns for one record's photo
// set, judged as a whole (D3). Pointer fields are nullable on the wire.
type Observation struct {
	ImageUsable   string  `json:"image_usable"`
	LabelVisible  string  `json:"label_visible"`
	LabelResiRead *string `json:"label_resi_read"`
	PackageSealed string  `json:"package_sealed"`
	ItemVisible   string  `json:"item_visible"`
	KoliSeen      *int    `json:"koli_seen"`
}

// ErrSchema marks a model response that does not match the schema.
// Such a response is discarded and the record stays unchecked; it is
// counted, because the schema-failure rate is half of a D8 target.
var ErrSchema = errors.New("evidencecheck: response does not match schema")

// ParseObservation decodes and validates a model response strictly:
// unknown fields, missing fields, out-of-enum values, and trailing data
// are all ErrSchema. Being lenient here would let a model that drifted
// into prose leak half an answer into a verdict.
func ParseObservation(raw []byte) (Observation, error) {
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(raw, &probe); err != nil {
		return Observation{}, fmt.Errorf("%w: %w", ErrSchema, err)
	}
	for _, k := range requiredFields {
		if _, ok := probe[k]; !ok {
			return Observation{}, fmt.Errorf("%w: missing %s", ErrSchema, k)
		}
	}

	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var o Observation
	if err := dec.Decode(&o); err != nil {
		return Observation{}, fmt.Errorf("%w: %w", ErrSchema, err)
	}
	if dec.More() {
		return Observation{}, fmt.Errorf("%w: trailing data", ErrSchema)
	}

	if !oneOf(o.ImageUsable, UsableOK, UsableDark, UsableBlurry, UsableObstructed) {
		return Observation{}, fmt.Errorf("%w: image_usable=%q", ErrSchema, o.ImageUsable)
	}
	for name, v := range map[string]string{
		"label_visible":  o.LabelVisible,
		"package_sealed": o.PackageSealed,
		"item_visible":   o.ItemVisible,
	} {
		if !oneOf(v, Yes, No, Unsure) {
			return Observation{}, fmt.Errorf("%w: %s=%q", ErrSchema, name, v)
		}
	}
	if o.KoliSeen != nil && *o.KoliSeen < 0 {
		return Observation{}, fmt.Errorf("%w: koli_seen=%d", ErrSchema, *o.KoliSeen)
	}
	return o, nil
}

var requiredFields = []string{
	"image_usable", "label_visible", "label_resi_read",
	"package_sealed", "item_visible", "koli_seen",
}

func oneOf(v string, allowed ...string) bool {
	for _, a := range allowed {
		if v == a {
			return true
		}
	}
	return false
}

// HasUnsure reports whether any tri-state field came back "unsure".
// Used by the eval to measure the D8 "unsure plus schema failures"
// rate; it has no effect on the verdict.
func (o Observation) HasUnsure() bool {
	return o.LabelVisible == Unsure || o.PackageSealed == Unsure || o.ItemVisible == Unsure
}
