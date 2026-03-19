package png

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/kettek/apng"
)

// RenderedFrame adalah hasil akhir: canvas yang sudah jadi per frame
type RenderedFrame struct {
	Image    *image.RGBA
	DelayNum uint16
	DelayDen uint16
}

// RenderAPNGFrames menerima slice of Frame dan mengembalikan
// slice of RenderedFrame — tiap elemen adalah "foto" canvas yang siap pakai.
func RenderAPNGFrames(frames []apng.Frame, canvasWidth, canvasHeight int) []RenderedFrame {
	if len(frames) == 0 {
		return nil
	}

	canvas := image.NewRGBA(image.Rect(0, 0, canvasWidth, canvasHeight))
	var prevCanvas *image.RGBA // snapshot untuk dispose=PREVIOUS

	results := make([]RenderedFrame, 0, len(frames))

	for _, f := range frames {
		bounds := f.Image.Bounds()
		fw, fh := bounds.Dx(), bounds.Dy()
		dst := image.Rect(f.XOffset, f.YOffset, f.XOffset+fw, f.YOffset+fh)

		// --- 1. Simpan snapshot SEBELUM render jika dispose=PREVIOUS ---
		if f.DisposeOp == 2 {
			prevCanvas = cloneRGBA(canvas)
		}

		// --- 2. Render frame ke canvas ---
		switch f.BlendOp {
		case 0: // SOURCE: timpa langsung, abaikan alpha canvas
			draw.Draw(canvas, dst, f.Image, bounds.Min, draw.Src)
		case 1: // OVER: alpha composite normal
			draw.Draw(canvas, dst, f.Image, bounds.Min, draw.Over)
		}

		// --- 3. Snapshot canvas ini = output frame ---
		results = append(results, RenderedFrame{
			Image:    cloneRGBA(canvas),
			DelayNum: f.DelayNumerator,
			DelayDen: f.DelayDenominator,
		})

		// --- 4. Post-dispose: siapkan canvas untuk frame berikutnya ---
		switch f.DisposeOp {
		case apng.DISPOSE_OP_NONE: // NONE: biarkan canvas seperti adanya
			// tidak ada yang dilakukan

		case apng.DISPOSE_OP_BACKGROUND: // BACKGROUND: bersihkan area frame ke transparan
			draw.Draw(canvas, dst, image.Transparent, image.Point{}, draw.Src)

		case apng.DISPOSE_OP_PREVIOUS: // PREVIOUS: kembalikan canvas ke snapshot
			if prevCanvas != nil {
				canvas = cloneRGBA(prevCanvas)
			} else {
				// Frame pertama tidak punya "previous", fallback ke transparan
				canvas = image.NewRGBA(image.Rect(0, 0, canvasWidth, canvasHeight))
			}
		}
	}

	return results
}

// cloneRGBA membuat salinan independen dari *image.RGBA
func cloneRGBA(src *image.RGBA) *image.RGBA {
	dst := image.NewRGBA(src.Bounds())
	copy(dst.Pix, src.Pix)
	return dst
}

// clearRect membersihkan area tertentu ke warna transparan.
// (Helper opsional, saat ini draw.Draw sudah cukup)
func clearRect(img *image.RGBA, r image.Rectangle) {
	draw.Draw(img, r, image.NewUniform(color.Transparent), image.Point{}, draw.Src)
}
