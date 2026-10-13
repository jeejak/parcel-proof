package evidencecheck

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"reflect"
	"strings"
	"testing"
)

func sp(s string) *string { return &s }
func ip(i int) *int       { return &i }

// good is an observation that should produce "kuat" for resi SPX123.
func good() Observation {
	return Observation{
		ImageUsable:   UsableOK,
		LabelVisible:  Yes,
		LabelResiRead: sp("SPX123"),
		PackageSealed: Yes,
		ItemVisible:   Yes,
		KoliSeen:      ip(1),
	}
}

func TestDecide(t *testing.T) {
	mod := func(f func(o *Observation)) Observation {
		o := good()
		f(&o)
		return o
	}
	cases := []struct {
		name    string
		obs     Observation
		resi    string
		koli    int
		verdict string
		reasons []string
	}{
		{"all good", good(), "SPX123", 1, VerdictKuat, nil},
		{"resi formatting differs", mod(func(o *Observation) { o.LabelResiRead = sp("spx 1-23") }), "SPX123", 1, VerdictKuat, nil},
		{"dark", mod(func(o *Observation) { o.ImageUsable = UsableDark }), "SPX123", 1, VerdictLemah, []string{ReasonImageDark}},
		{"blurry", mod(func(o *Observation) { o.ImageUsable = UsableBlurry }), "SPX123", 1, VerdictLemah, []string{ReasonImageBlurry}},
		{"obstructed", mod(func(o *Observation) { o.ImageUsable = UsableObstructed }), "SPX123", 1, VerdictLemah, []string{ReasonImageObstructed}},
		{"wrong resi on label", mod(func(o *Observation) { o.LabelResiRead = sp("JNT999") }), "SPX123", 1, VerdictLemah, []string{ReasonLabelMismatch}},
		{"no label", mod(func(o *Observation) { o.LabelVisible = No; o.LabelResiRead = nil }), "SPX123", 1, VerdictLemah, []string{ReasonLabelNotVisible}},
		{"label visible unreadable", mod(func(o *Observation) { o.LabelResiRead = nil }), "SPX123", 1, VerdictPerluDicek, []string{ReasonLabelUnreadable}},
		{"empty read treated as null", mod(func(o *Observation) { o.LabelResiRead = sp(" - ") }), "SPX123", 1, VerdictPerluDicek, []string{ReasonLabelUnreadable}},
		{"label unsure no read", mod(func(o *Observation) { o.LabelVisible = Unsure; o.LabelResiRead = nil }), "SPX123", 1, VerdictKuat, nil},
		{"item not visible", mod(func(o *Observation) { o.ItemVisible = No }), "SPX123", 1, VerdictLemah, []string{ReasonItemNotVisible}},
		{"not sealed", mod(func(o *Observation) { o.PackageSealed = No }), "SPX123", 1, VerdictPerluDicek, []string{ReasonNotSealed}},
		{"koli mismatch", mod(func(o *Observation) { o.KoliSeen = ip(2) }), "SPX123", 1, VerdictPerluDicek, []string{ReasonKoliMismatch}},
		{"koli null ignored", mod(func(o *Observation) { o.KoliSeen = nil }), "SPX123", 3, VerdictKuat, nil},
		{"all unsure is kuat", mod(func(o *Observation) {
			o.LabelVisible, o.PackageSealed, o.ItemVisible = Unsure, Unsure, Unsure
			o.LabelResiRead, o.KoliSeen = nil, nil
		}), "SPX123", 1, VerdictKuat, nil},
		{"weak wins over soft", mod(func(o *Observation) { o.PackageSealed = No; o.ItemVisible = No }), "SPX123", 1, VerdictLemah, []string{ReasonItemNotVisible, ReasonNotSealed}},
		{"reason order stable", mod(func(o *Observation) {
			o.ImageUsable = UsableDark
			o.LabelResiRead = sp("X")
			o.ItemVisible = No
			o.PackageSealed = No
			o.KoliSeen = ip(4)
		}), "SPX123", 1, VerdictLemah, []string{ReasonImageDark, ReasonLabelMismatch, ReasonItemNotVisible, ReasonNotSealed, ReasonKoliMismatch}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Decide(Input{Obs: c.obs, Resi: c.resi, Koli: c.koli})
			if got.Verdict != c.verdict {
				t.Errorf("verdict = %q, want %q", got.Verdict, c.verdict)
			}
			if !reflect.DeepEqual(got.Reasons, c.reasons) {
				t.Errorf("reasons = %v, want %v", got.Reasons, c.reasons)
			}
		})
	}
}

func TestParseObservation(t *testing.T) {
	ok := `{"image_usable":"ok","label_visible":"yes","label_resi_read":"SPX1","package_sealed":"no","item_visible":"unsure","koli_seen":null}`
	o, err := ParseObservation([]byte(ok))
	if err != nil {
		t.Fatalf("valid response: %v", err)
	}
	if o.LabelResiRead == nil || *o.LabelResiRead != "SPX1" || o.KoliSeen != nil || !o.HasUnsure() {
		t.Fatalf("decoded wrong: %+v", o)
	}

	bad := map[string]string{
		"not json":      `the photo looks fine`,
		"missing field": `{"image_usable":"ok","label_visible":"yes","label_resi_read":null,"package_sealed":"no","item_visible":"yes"}`,
		"unknown field": strings.Replace(ok, `"koli_seen":null`, `"koli_seen":null,"note":"x"`, 1),
		"bad enum":      strings.Replace(ok, `"image_usable":"ok"`, `"image_usable":"fine"`, 1),
		"bad tri":       strings.Replace(ok, `"package_sealed":"no"`, `"package_sealed":"maybe"`, 1),
		"negative koli": strings.Replace(ok, `"koli_seen":null`, `"koli_seen":-1`, 1),
		"trailing data": ok + `{}`,
		"wrong type":    strings.Replace(ok, `"koli_seen":null`, `"koli_seen":"two"`, 1),
	}
	for name, raw := range bad {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseObservation([]byte(raw)); !errors.Is(err, ErrSchema) {
				t.Fatalf("err = %v, want ErrSchema", err)
			}
		})
	}
}

func TestSelectPhotos(t *testing.T) {
	cases := map[int][]int{
		0: {},
		3: {0, 1, 2},
		4: {0, 1, 2, 3},
		5: {0, 1, 3, 4},
		9: {0, 1, 7, 8},
	}
	for n, want := range cases {
		in := make([]int, n)
		for i := range in {
			in[i] = i
		}
		got := SelectPhotos(in)
		if len(got) == 0 && len(want) == 0 {
			continue
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("n=%d: got %v, want %v", n, got, want)
		}
	}
}

func TestDownscale(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 3000, 1500))
	for y := 0; y < 1500; y++ {
		for x := 0; x < 3000; x++ {
			src.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}
	var in bytes.Buffer
	if err := jpeg.Encode(&in, src, nil); err != nil {
		t.Fatal(err)
	}
	out, err := Downscale(&in, MaxEdge)
	if err != nil {
		t.Fatal(err)
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	if format != "jpeg" || cfg.Width != 1024 || cfg.Height != 512 {
		t.Fatalf("got %s %dx%d, want jpeg 1024x512", format, cfg.Width, cfg.Height)
	}

	// Never enlarges.
	small := image.NewRGBA(image.Rect(0, 0, 200, 100))
	in.Reset()
	if err := jpeg.Encode(&in, small, nil); err != nil {
		t.Fatal(err)
	}
	out, err = Downscale(&in, MaxEdge)
	if err != nil {
		t.Fatal(err)
	}
	cfg, _, _ = image.DecodeConfig(bytes.NewReader(out))
	if cfg.Width != 200 || cfg.Height != 100 {
		t.Fatalf("small image resized to %dx%d", cfg.Width, cfg.Height)
	}

	if _, err := Downscale(strings.NewReader("not an image"), MaxEdge); err == nil {
		t.Fatal("garbage input decoded")
	}
}

// endlessReader never returns EOF. Downscale must not try to read it
// all: an unbounded io.ReadAll here would hang until the process died,
// so reaching the assertion at all is the proof that the read is
// bounded. Phase 1 feeds Downscale from object storage, where the
// object's length is whatever was written, not whatever the upload
// surface accepted.
type endlessReader struct{}

func (endlessReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 'A'
	}
	return len(p), nil
}

func TestDownscaleRefusesOversizedInput(t *testing.T) {
	_, err := Downscale(endlessReader{}, MaxEdge)
	if err == nil {
		t.Fatal("an endless reader was accepted")
	}
	if !errors.Is(err, ErrImageTooLarge) {
		t.Fatalf("got %v, want ErrImageTooLarge", err)
	}
}

// The selection budget is MaxPhotos, and the split is derived from it.
// These are the properties that must hold at ANY MaxPhotos, so changing
// the constant cannot quietly start sending a different count or the
// same photo twice.
func TestSelectPhotosRespectsBudgetAtAnySize(t *testing.T) {
	for n := 0; n <= 3*MaxPhotos+1; n++ {
		in := make([]int, n)
		for i := range in {
			in[i] = i
		}
		got := SelectPhotos(in)
		if len(got) > MaxPhotos {
			t.Fatalf("n=%d returned %d photos, over the MaxPhotos=%d budget", n, len(got), MaxPhotos)
		}
		if n <= MaxPhotos && len(got) != n {
			t.Fatalf("n=%d returned %d; sets within budget must be sent whole", n, len(got))
		}
		if n > MaxPhotos && len(got) != MaxPhotos {
			t.Fatalf("n=%d returned %d; oversized sets must fill the budget", n, len(got))
		}
		seen := map[int]bool{}
		for _, v := range got {
			if seen[v] {
				t.Fatalf("n=%d sent index %d twice: %v", n, v, got)
			}
			seen[v] = true
		}
	}
}

func TestUserTextOmitsRecordIdentity(t *testing.T) {
	// D2: the model is told photo count, courier and koli, nothing that
	// identifies the record, the buyer, or the resi to expect.
	got := UserText(Request{Photos: make([][]byte, 2), Koli: 3, Courier: "SPX"})
	want := "Photos: 2. Courier: SPX. The seller recorded 3 koli for this shipment."
	if got != want {
		t.Fatalf("UserText = %q", got)
	}
	if strings.Contains(strings.ToLower(SystemInstruction), "expected resi") {
		t.Fatal("system instruction must not reference an expected resi")
	}
}

func TestResponseSchemaMatchesObservation(t *testing.T) {
	// Every Observation JSON field is in the schema and required, so a
	// field added to one and not the other fails here, not in production.
	props := ResponseSchema()["properties"].(map[string]any)
	ty := reflect.TypeOf(Observation{})
	if ty.NumField() != len(props) || len(requiredFields) != len(props) {
		t.Fatalf("Observation has %d fields, schema %d, required %d", ty.NumField(), len(props), len(requiredFields))
	}
	for i := 0; i < ty.NumField(); i++ {
		tag := ty.Field(i).Tag.Get("json")
		if _, ok := props[tag]; !ok {
			t.Errorf("schema missing %q", tag)
		}
	}
}

// PromptVersion is part of the D6 idempotency key: a record is due for
// a re-check when the prompt that judged it changed. Nothing enforced
// that, so editing SystemInstruction, UserText or ResponseSchema while
// leaving PromptVersion alone would silently leave every already-checked
// record holding a verdict from a prompt that no longer exists, and no
// re-check would ever be scheduled.
//
// This digest covers the whole prompt surface. When it fails, that is
// correct: bump PromptVersion, then paste the printed digest here. The
// digest is not a value to keep stable, it is a tripwire on changing one
// without the other.
func TestPromptVersionCoversThePromptSurface(t *testing.T) {
	const (
		wantVersion = "2026-10-07.1"
		wantDigest  = "37d414932c05f27e883197d49ccfd71cab2f33f775401379349fe753aae834a8"
	)
	if PromptVersion != wantVersion {
		t.Fatalf("PromptVersion is %q but this test pins %q; update wantVersion and wantDigest together", PromptVersion, wantVersion)
	}

	schema, err := json.Marshal(ResponseSchema())
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.New()
	h.Write([]byte(SystemInstruction))
	h.Write([]byte{0})
	// A fixed request so the digest tracks UserText's wording, not the
	// values a caller happens to pass.
	h.Write([]byte(UserText(Request{Photos: make([][]byte, 3), Koli: 2, Courier: "jne"})))
	h.Write([]byte{0})
	h.Write(schema)
	got := hex.EncodeToString(h.Sum(nil))

	if got != wantDigest {
		t.Fatalf("the prompt surface changed but PromptVersion is still %q.\n"+
			"Bump PromptVersion (every checked record becomes due for a re-check), then set:\n"+
			"    wantDigest = %q", PromptVersion, got)
	}
}
