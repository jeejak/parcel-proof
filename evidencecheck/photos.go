package evidencecheck

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/png" // evidence captured as PNG is rarer but legal
	"io"
)

// SelectPhotos picks which of a record's packing photos to send (D2).
// photos must already be in order_index order. Up to MaxPhotos are
// sent as-is; beyond that, the first half and the last half of that
// budget (so first two and last two at MaxPhotos = 4). The first photos
// usually show the item, the last ones the sealed box with its label,
// and taking only the first four (the dispute PDF's order) would report
// "not sealed" on a correctly packed record.
func SelectPhotos[T any](photos []T) []T {
	if len(photos) <= MaxPhotos {
		return photos
	}
	// Split derived from MaxPhotos rather than hardcoded at 2+2: with a
	// literal first-two-and-last-two, raising MaxPhotos would silently
	// keep sending four photos, and lowering it to 3 would emit four
	// entries with a duplicate in the middle.
	head := MaxPhotos / 2
	tail := MaxPhotos - head
	out := make([]T, 0, MaxPhotos)
	out = append(out, photos[:head]...)
	return append(out, photos[len(photos)-tail:]...)
}

// MaxEdge is the long-edge size photos are downscaled to before they
// are sent (D2). Smaller means fewer image tokens; 1024 keeps a
// thermal waybill's printed number legible.
const MaxEdge = 1024

// maxDecodePixels bounds decoding. Photos are capped at 5 MiB on
// upload, but a small, highly compressible file can still declare an
// enormous canvas; refusing it before allocating is cheaper than an OOM.
const maxDecodePixels = 40_000_000

// MaxEncodedBytes bounds how many bytes Downscale will read from its
// reader. The 5 MiB upload cap is enforced at the upload surface, not
// here, and Downscale takes an io.Reader: Phase 1 feeds it from object
// storage, where the stored object's length is whatever was written.
// Reading without a limit makes one oversized object an OOM for the
// whole binary, so the limit lives with the read. 8 MiB leaves room
// above the upload cap for re-encoding without inviting a large read.
const MaxEncodedBytes = 8 << 20

// ErrImageTooLarge is returned for an image whose declared canvas
// exceeds maxDecodePixels, or whose encoded form exceeds
// MaxEncodedBytes.
var ErrImageTooLarge = errors.New("evidencecheck: image dimensions too large")

// Downscale decodes a JPEG or PNG, shrinks it so the long edge is at
// most maxEdge (never enlarges), and re-encodes it as JPEG. The output
// exists only in memory and is never written anywhere (D9).
func Downscale(r io.Reader, maxEdge int) ([]byte, error) {
	// Read one byte past the limit so a file AT the limit still decodes
	// while anything larger is refused instead of being silently
	// truncated into a corrupt image.
	raw, err := io.ReadAll(io.LimitReader(r, MaxEncodedBytes+1))
	if err != nil {
		return nil, err
	}
	if len(raw) > MaxEncodedBytes {
		return nil, fmt.Errorf("%w: encoded size over %d bytes", ErrImageTooLarge, MaxEncodedBytes)
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("evidencecheck: decode config: %w", err)
	}
	if cfg.Width <= 0 || cfg.Height <= 0 || cfg.Width*cfg.Height > maxDecodePixels {
		return nil, ErrImageTooLarge
	}
	src, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("evidencecheck: decode: %w", err)
	}

	dst := shrink(src, maxEdge)
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 85}); err != nil {
		return nil, fmt.Errorf("evidencecheck: encode: %w", err)
	}
	return buf.Bytes(), nil
}

// shrink box-filters src down so its long edge is maxEdge. Box
// averaging is enough for a 2x to 4x reduction of a phone photo and
// keeps this package on the standard library.
func shrink(src image.Image, maxEdge int) image.Image {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	long := w
	if h > long {
		long = h
	}
	if maxEdge <= 0 || long <= maxEdge {
		return src
	}
	nw := w * maxEdge / long
	nh := h * maxEdge / long
	if nw < 1 {
		nw = 1
	}
	if nh < 1 {
		nh = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
	for y := 0; y < nh; y++ {
		y0 := b.Min.Y + y*h/nh
		y1 := b.Min.Y + (y+1)*h/nh
		if y1 <= y0 {
			y1 = y0 + 1
		}
		for x := 0; x < nw; x++ {
			x0 := b.Min.X + x*w/nw
			x1 := b.Min.X + (x+1)*w/nw
			if x1 <= x0 {
				x1 = x0 + 1
			}
			var rs, gs, bs, n uint64
			for sy := y0; sy < y1; sy++ {
				for sx := x0; sx < x1; sx++ {
					r, g, bl, _ := src.At(sx, sy).RGBA()
					rs += uint64(r)
					gs += uint64(g)
					bs += uint64(bl)
					n++
				}
			}
			dst.SetRGBA(x, y, color.RGBA{
				R: uint8((rs / n) >> 8), //nolint:gosec // G115: RGBA() channels are <= 0xffff, so mean>>8 <= 0xff
				G: uint8((gs / n) >> 8), //nolint:gosec // G115: RGBA() channels are <= 0xffff, so mean>>8 <= 0xff
				B: uint8((bs / n) >> 8), //nolint:gosec // G115: RGBA() channels are <= 0xffff, so mean>>8 <= 0xff
				A: 0xff,
			})
		}
	}
	return dst
}
