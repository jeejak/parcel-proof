package evidencecheck

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"math/rand"
	"testing"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/oned"
)

// testPhoto builds a 200x150 JPEG: a random grey texture (many edges),
// scaled by brightness, or a flat field when flat is true.
func testPhoto(t *testing.T, brightness float64, flat bool) []byte {
	t.Helper()
	img := image.NewGray(image.Rect(0, 0, 200, 150))
	r := rand.New(rand.NewSource(1))
	for y := 0; y < 150; y++ {
		for x := 0; x < 200; x++ {
			v := 128.0
			if !flat {
				v = float64(r.Intn(256))
			}
			img.SetGray(x, y, color.Gray{Y: uint8(v * brightness)})
		}
	}
	var b bytes.Buffer
	if err := jpeg.Encode(&b, img, &jpeg.Options{Quality: 95}); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func TestBasicChecker(t *testing.T) {
	c := BasicChecker{DarkBelow: 78.5, BlurBelow: 31.5}
	sharp, dark, flat := testPhoto(t, 1, false), testPhoto(t, 0.2, false), testPhoto(t, 1, true)
	cases := map[string]struct {
		photos [][]byte
		want   string
	}{
		"bright and sharp":        {[][]byte{sharp, sharp}, UsableOK},
		"one dark photo":          {[][]byte{sharp, dark}, UsableDark},
		"one flat (blurry) photo": {[][]byte{sharp, flat}, UsableBlurry},
		"dark wins over blurry":   {[][]byte{flat, dark}, UsableDark},
	}
	for name, tc := range cases {
		obs, _, err := c.Check(context.Background(), Request{Photos: tc.photos, Koli: 1})
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if obs.ImageUsable != tc.want {
			t.Errorf("%s: image_usable = %q, want %q", name, obs.ImageUsable, tc.want)
		}
		// Everything it cannot see is unsure, which Decide never turns
		// into a reason (D3).
		if obs.LabelVisible != Unsure || obs.PackageSealed != Unsure || obs.ItemVisible != Unsure || obs.LabelResiRead != nil {
			t.Errorf("%s: non-pixel fields not unsure: %+v", name, obs)
		}
	}
	if _, _, err := c.Check(context.Background(), Request{}); err == nil {
		t.Error("no photos accepted")
	}
	if _, _, err := c.Check(context.Background(), Request{Photos: [][]byte{[]byte("not a jpeg")}}); err == nil {
		t.Error("garbage accepted")
	}
}

func TestBasicVerdict(t *testing.T) {
	// A dark record decides lemah with only image_dark; an ok record with
	// every other field unsure decides kuat.
	dark := Decide(Input{Obs: Observation{ImageUsable: UsableDark, LabelVisible: Unsure, PackageSealed: Unsure, ItemVisible: Unsure}, Resi: "X", Koli: 1})
	ok := Decide(Input{Obs: Observation{ImageUsable: UsableOK, LabelVisible: Unsure, PackageSealed: Unsure, ItemVisible: Unsure}, Resi: "X", Koli: 1})
	if dark.Verdict != VerdictLemah || len(dark.Reasons) != 1 || ok.Verdict != VerdictKuat {
		t.Fatalf("dark %+v ok %+v", dark, ok)
	}
}

// labelPhoto renders barcodes stacked on a white page: each value as
// Code 128 unless it is a 13-digit EAN, which is how product packaging
// prints it.
func labelPhoto(t *testing.T, values ...string) []byte {
	t.Helper()
	const w, rowH = 600, 160
	page := image.NewGray(image.Rect(0, 0, w, rowH*len(values)+40))
	for i := range page.Pix {
		page.Pix[i] = 255
	}
	for i, v := range values {
		var m *gozxing.BitMatrix
		var err error
		if len(v) == 13 {
			m, err = oned.NewEAN13Writer().Encode(v, gozxing.BarcodeFormat_EAN_13, 500, 120, nil)
		} else {
			m, err = oned.NewCode128Writer().Encode(v, gozxing.BarcodeFormat_CODE_128, 500, 120, nil)
		}
		if err != nil {
			t.Fatal(err)
		}
		for y := 0; y < m.GetHeight(); y++ {
			for x := 0; x < m.GetWidth(); x++ {
				if m.Get(x, y) {
					page.SetGray(50+x, 20+i*rowH+y, color.Gray{})
				}
			}
		}
	}
	var b bytes.Buffer
	if err := jpeg.Encode(&b, page, &jpeg.Options{Quality: 95}); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func TestBasicChecker_ReadsOneShippingBarcode(t *testing.T) {
	c := BasicChecker{DarkBelow: 78.5, BlurBelow: 0, ReadBarcodes: true}
	read := func(photos ...[]byte) Observation {
		t.Helper()
		obs, _, err := c.Check(context.Background(), Request{Photos: photos, Koli: 1})
		if err != nil {
			t.Fatal(err)
		}
		return obs
	}

	o := read(labelPhoto(t, "SPXID5774637064"))
	if o.LabelVisible != Yes || o.LabelResiRead == nil || *o.LabelResiRead != "SPXID5774637064" {
		t.Fatalf("one Code 128: %+v", o)
	}
	// The same value in two photos is still one value.
	if o := read(labelPhoto(t, "JP3304452940"), labelPhoto(t, "JP3304452940")); o.LabelResiRead == nil {
		t.Fatalf("same value twice: %+v", o)
	}
	// Resi plus order id: both are found (the halves scan), and which
	// one is the resi is unknown, so the label stays unsure.
	if got := code128Values([][]byte{labelPhoto(t, "JP3304452940", "ORDER260911QFN")}); len(got) != 2 {
		t.Fatalf("two-barcode label decoded to %v, want both values", got)
	}
	if o := read(labelPhoto(t, "JP3304452940", "ORDER260911QFN")); o.LabelVisible != Unsure || o.LabelResiRead != nil {
		t.Fatalf("two Code 128 values read as one: %+v", o)
	}
	// A product EAN is not a shipping label.
	if o := read(labelPhoto(t, "8992761111144")); o.LabelVisible != Unsure || o.LabelResiRead != nil {
		t.Fatalf("EAN read as a resi: %+v", o)
	}
	// Off by default.
	c.ReadBarcodes = false
	if o := read(labelPhoto(t, "SPXID5774637064")); o.LabelResiRead != nil {
		t.Fatalf("read with ReadBarcodes off: %+v", o)
	}
}
