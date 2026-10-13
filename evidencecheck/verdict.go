package evidencecheck

import "strings"

// Verdict values, stored and sent on the wire as-is (D4, D7).
const (
	VerdictKuat       = "kuat"
	VerdictPerluDicek = "perlu_dicek"
	VerdictLemah      = "lemah"
)

// Reason codes. Seller-facing text is a fixed string per code, drafted
// in Phase 1 (Plan 0016 §7 Q5); the model's own words never reach a
// seller.
const (
	ReasonImageDark       = "image_dark"
	ReasonImageBlurry     = "image_blurry"
	ReasonImageObstructed = "image_obstructed"
	ReasonLabelNotVisible = "label_not_visible"
	ReasonLabelUnreadable = "label_unreadable"
	ReasonLabelMismatch   = "label_mismatch"
	ReasonItemNotVisible  = "item_not_visible"
	ReasonNotSealed       = "not_sealed"
	ReasonKoliMismatch    = "koli_mismatch"
)

// weakReasons are the ones that, on their own, mean the evidence would
// likely lose a dispute: the photos cannot be read, nothing ties them
// to this resi, or the contents are not shown. Everything else is a
// gap worth a look but not disqualifying.
var weakReasons = map[string]bool{
	ReasonImageDark:       true,
	ReasonImageBlurry:     true,
	ReasonImageObstructed: true,
	ReasonLabelNotVisible: true,
	ReasonLabelMismatch:   true,
	ReasonItemNotVisible:  true,
}

// Input is what Decide needs: the model's observation plus the two
// facts from the record the model was deliberately not given.
type Input struct {
	Obs  Observation
	Resi string // evidence_records.resi_number
	Koli int    // evidence_records.koli, >= 1 by schema
}

// Result is a verdict with the reasons that produced it, in a stable
// order so stored rows and tests compare cleanly.
type Result struct {
	Verdict string
	Reasons []string
}

// Decide is the whole server-side rule (D4). Pure and deterministic.
//
//   - "unsure" never produces a reason (D3).
//   - The resi comparison is done here, on normalised strings. A read
//     that does not match is label_mismatch; a visible label with no
//     read is label_unreadable.
//   - Any weak reason makes the verdict lemah; any other reason makes it
//     perlu_dicek; no reasons is kuat.
func Decide(in Input) Result {
	o := in.Obs
	var reasons []string

	switch o.ImageUsable {
	case UsableDark:
		reasons = append(reasons, ReasonImageDark)
	case UsableBlurry:
		reasons = append(reasons, ReasonImageBlurry)
	case UsableObstructed:
		reasons = append(reasons, ReasonImageObstructed)
	}

	read := ""
	if o.LabelResiRead != nil {
		read = NormaliseResi(*o.LabelResiRead)
	}
	switch {
	case read != "":
		if read != NormaliseResi(in.Resi) {
			reasons = append(reasons, ReasonLabelMismatch)
		}
	case o.LabelVisible == No:
		reasons = append(reasons, ReasonLabelNotVisible)
	case o.LabelVisible == Yes:
		reasons = append(reasons, ReasonLabelUnreadable)
	}

	if o.ItemVisible == No {
		reasons = append(reasons, ReasonItemNotVisible)
	}
	if o.PackageSealed == No {
		reasons = append(reasons, ReasonNotSealed)
	}
	if o.KoliSeen != nil && in.Koli > 0 && *o.KoliSeen != in.Koli {
		reasons = append(reasons, ReasonKoliMismatch)
	}

	verdict := VerdictKuat
	for _, r := range reasons {
		if weakReasons[r] {
			verdict = VerdictLemah
			break
		}
		verdict = VerdictPerluDicek
	}
	return Result{Verdict: verdict, Reasons: reasons}
}

// NormaliseResi uppercases and drops everything that is not a letter
// or digit, so "spxid 0123-456" and "SPXID0123456" compare equal. A
// waybill number's meaning is in its alphanumerics; spacing and dashes
// vary by courier print template and by how a model transcribes them.
func NormaliseResi(s string) string {
	var b strings.Builder
	for _, r := range strings.ToUpper(s) {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}
