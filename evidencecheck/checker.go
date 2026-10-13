package evidencecheck

import (
	"context"
	"fmt"
	"time"
)

// PromptVersion is part of the idempotency key (D6): changing the
// prompt or schema means every record is due for a re-check. Bump it on
// any edit to SystemInstruction, UserText or ResponseSchema.
const PromptVersion = "2026-10-07.1"

// MaxPhotos is how many photos one check sends (D2).
const MaxPhotos = 4

// Request is one record's check. Photos are already selected and
// downscaled JPEG bytes. Koli and courier are the only record facts
// sent; the resi, recipient and operator are deliberately absent (D2).
type Request struct {
	Photos  [][]byte
	Koli    int
	Courier string
}

// Usage is what a check cost and took, recorded per row and summed by
// the eval (D8, D9).
type Usage struct {
	Model       string
	InputTokens int
	// OutputTokens is everything billed at the output rate: the answer
	// plus the model's reasoning. ThoughtTokens is the reasoning part,
	// kept apart so the eval can show how much of the cost is thinking.
	OutputTokens  int
	ThoughtTokens int
	Latency       time.Duration
}

// Checker returns the model's observation for one record. A response
// that fails the schema returns an error wrapping ErrSchema; callers
// discard it and leave the record unchecked.
type Checker interface {
	Check(ctx context.Context, req Request) (Observation, Usage, error)
}

// SystemInstruction tells the model what to look for and, as
// importantly, to say "unsure" rather than guess. It never mentions a
// resi to expect.
const SystemInstruction = `You inspect packing photos taken by an online seller in Indonesia before a parcel is handed to a courier. The photos of one parcel are given together. Judge the set as a whole, not each photo: different photos usually show different steps, for example the item before sealing and the sealed box with its shipping label.

Report only what is visible. When you cannot tell, answer "unsure". Do not guess.

Fields:
- image_usable: the most serious problem found in any photo: "dark", "blurry", or "obstructed", or "ok" if there is none. "obstructed" means a hand, finger, bag, cloth or other object covers what a photo is meant to show: the item, the shipping label, or the seal. A hand that is packing or taping, with nothing important hidden, is not an obstruction.
- label_visible: whether a printed shipping label (waybill) is visible in any photo. Answer "no" if the label is covered so that its printed number cannot be seen in any photo.
- label_resi_read: the tracking number printed on the label, copied exactly as printed, or null if no label is visible or the number cannot be read with confidence. Never invent or complete a partial number.
- package_sealed: whether any photo shows the package closed and sealed. If the place where the box would be sealed is hidden in every photo, answer "unsure". Do not infer a seal you cannot see.
- item_visible: whether any photo shows the goods themselves before the package was closed.
- koli_seen: the number of separate packages (koli) visible, or null if it cannot be counted.

Return only the JSON object.`

// UserText is the text part sent with the photos.
func UserText(req Request) string {
	courier := req.Courier
	if courier == "" {
		courier = "unknown"
	}
	return fmt.Sprintf("Photos: %d. Courier: %s. The seller recorded %d koli for this shipment.",
		len(req.Photos), courier, req.Koli)
}

// ResponseSchema is the Vertex AI responseSchema (OpenAPI subset) that
// constrains the model to the Observation shape (D3).
func ResponseSchema() map[string]any {
	tri := func() map[string]any {
		return map[string]any{"type": "STRING", "enum": []string{Yes, No, Unsure}}
	}
	return map[string]any{
		"type": "OBJECT",
		"properties": map[string]any{
			"image_usable": map[string]any{
				"type": "STRING",
				"enum": []string{UsableOK, UsableDark, UsableBlurry, UsableObstructed},
			},
			"label_visible":   tri(),
			"label_resi_read": map[string]any{"type": "STRING", "nullable": true},
			"package_sealed":  tri(),
			"item_visible":    tri(),
			"koli_seen":       map[string]any{"type": "INTEGER", "nullable": true},
		},
		"required":         requiredFields,
		"propertyOrdering": requiredFields,
	}
}

// FakeChecker is a Checker for tests and for dry runs of the eval.
type FakeChecker struct {
	Fn func(ctx context.Context, req Request) (Observation, Usage, error)
}

// Check implements Checker.
func (f FakeChecker) Check(ctx context.Context, req Request) (Observation, Usage, error) {
	if f.Fn == nil {
		return Observation{}, Usage{}, fmt.Errorf("evidencecheck: FakeChecker has no Fn")
	}
	return f.Fn(ctx, req)
}
