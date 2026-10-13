package evidencecheck

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg" // decode the downscaled JPEGs
	"time"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/oned"
)

// PhotoStats is what the basic check measures on one photo, with no
// model involved: mean brightness (0 to 255) and sharpness (variance of
// the Laplacian of the grayscale image; a blurred photo has few edges,
// so a low variance).
type PhotoStats struct {
	Luma      float64
	Sharpness float64
}

// Measure computes PhotoStats for one JPEG. Photos go through Downscale
// first (D2), so sharpness is compared at the same scale for every
// record.
func Measure(jpegBytes []byte) (PhotoStats, error) {
	img, _, err := image.Decode(bytes.NewReader(jpegBytes))
	if err != nil {
		return PhotoStats{}, fmt.Errorf("decode photo: %w", err)
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < 3 || h < 3 {
		return PhotoStats{}, fmt.Errorf("photo too small: %dx%d", w, h)
	}
	gray := make([]float64, w*h)
	var sum float64
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, bl, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
			// Rec. 601 luma on 8-bit channels.
			l := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(bl)) / 257
			gray[y*w+x] = l
			sum += l
		}
	}
	var lsum, lsq float64
	n := 0
	for y := 1; y < h-1; y++ {
		for x := 1; x < w-1; x++ {
			i := y*w + x
			lap := gray[i-w] + gray[i+w] + gray[i-1] + gray[i+1] - 4*gray[i]
			lsum += lap
			lsq += lap * lap
			n++
		}
	}
	mean := lsum / float64(n)
	return PhotoStats{
		Luma:      sum / float64(w*h),
		Sharpness: lsq/float64(n) - mean*mean,
	}, nil
}

// BasicChecker is the no-cost check: it judges only whether the photos
// are dark or blurry, from pixel statistics, and answers "unsure" for
// every other field, which Decide never turns into a reason (D3). No
// photo leaves the server and no model is called, so it needs no new
// data processor.
//
// A record is dark if any photo's mean brightness is below DarkBelow,
// otherwise blurry if any photo's sharpness is below BlurBelow. The
// thresholds are calibrated on the eval set, never guessed.
type BasicChecker struct {
	DarkBelow float64
	BlurBelow float64
	// ReadBarcodes also decodes Code 128 barcodes, the shipping-label
	// format (product EAN/UPC codes are ignored). When exactly one
	// distinct value is found across the photos it is reported as
	// label_resi_read, and Decide compares it with the record's resi.
	// Two or more values (a label that also prints the order id) or
	// none at all leave the label unsure. Off by default: its false
	// mismatch rate is unmeasured until real photos exist.
	ReadBarcodes bool
	// OnStats, if set, receives each photo's stats (for calibration).
	OnStats func(i int, s PhotoStats)
}

var _ Checker = BasicChecker{}

// Check implements Checker.
func (c BasicChecker) Check(ctx context.Context, req Request) (Observation, Usage, error) {
	start := time.Now()
	if len(req.Photos) == 0 {
		return Observation{}, Usage{}, fmt.Errorf("evidencecheck: no photos")
	}
	usable := UsableOK
	for i, p := range req.Photos {
		if err := ctx.Err(); err != nil {
			return Observation{}, Usage{}, err
		}
		s, err := Measure(p)
		if err != nil {
			return Observation{}, Usage{}, err
		}
		if c.OnStats != nil {
			c.OnStats(i, s)
		}
		switch {
		case s.Luma < c.DarkBelow:
			usable = UsableDark
		case s.Sharpness < c.BlurBelow && usable != UsableDark:
			usable = UsableBlurry
		}
	}
	obs := Observation{
		ImageUsable:  usable,
		LabelVisible: Unsure,
	}
	if c.ReadBarcodes {
		if v, ok := singleCode128(req.Photos); ok {
			obs.LabelVisible = Yes
			obs.LabelResiRead = &v
		}
	}
	obs = Observation{
		ImageUsable:   obs.ImageUsable,
		LabelVisible:  obs.LabelVisible,
		LabelResiRead: obs.LabelResiRead,
		PackageSealed: Unsure,
		ItemVisible:   Unsure,
	}
	return obs, Usage{Model: "basic", Latency: time.Since(start)}, nil
}

// singleCode128 returns the one distinct Code 128 value found across the
// photos, if there is exactly one.
func singleCode128(photos [][]byte) (string, bool) {
	seen := code128Values(photos)
	if len(seen) != 1 {
		return "", false
	}
	for v := range seen {
		return v, true
	}
	return "", false
}

// code128Values returns every distinct Code 128 value in the photos.
// The reader decodes one barcode per image, so each photo is also tried
// in halves, which finds a second barcode (an order id printed beside
// the resi) when there is one.
func code128Values(photos [][]byte) map[string]bool {
	seen := map[string]bool{}
	reader := oned.NewCode128Reader()
	hints := map[gozxing.DecodeHintType]interface{}{gozxing.DecodeHintType_TRY_HARDER: true}
	decode := func(img image.Image) {
		bmp, err := gozxing.NewBinaryBitmapFromImage(img)
		if err != nil {
			return
		}
		r, err := reader.Decode(bmp, hints)
		if err == nil && r.GetBarcodeFormat() == gozxing.BarcodeFormat_CODE_128 && r.GetText() != "" {
			seen[r.GetText()] = true
		}
	}
	for _, p := range photos {
		img, _, err := image.Decode(bytes.NewReader(p))
		if err != nil {
			continue
		}
		decode(img)
		sub, ok := img.(interface {
			SubImage(image.Rectangle) image.Image
		})
		if !ok {
			continue
		}
		b := img.Bounds()
		mx, my := (b.Min.X+b.Max.X)/2, (b.Min.Y+b.Max.Y)/2
		for _, r := range []image.Rectangle{
			image.Rect(b.Min.X, b.Min.Y, b.Max.X, my), image.Rect(b.Min.X, my, b.Max.X, b.Max.Y),
			image.Rect(b.Min.X, b.Min.Y, mx, b.Max.Y), image.Rect(mx, b.Min.Y, b.Max.X, b.Max.Y),
		} {
			decode(sub.SubImage(r))
		}
	}
	return seen
}
