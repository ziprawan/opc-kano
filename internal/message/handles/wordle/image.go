package wordle

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

func GenerateWordleImage(target string, guesses []string) ([]byte, error) {
	theFont, err := getFont()
	if err != nil {
		return nil, fmt.Errorf("failed to get font: %s", err)
	}

	img := image.NewRGBA(image.Rect(0, 0, 1960, 2300))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(theFont)
	c.SetFontSize(200)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.NewUniform(color.White))
	c.SetHinting(font.HintingNone)

	face := truetype.NewFace(theFont, &truetype.Options{
		Size: 200,
		DPI:  72,
	})
	metrics := face.Metrics()
	drawer := &font.Drawer{
		Face: face,
	}

	curPos := image.Point{150, 150} // For positioning purpose
	borderWidth := 4                // Square border length
	off := borderWidth / 2          // Just for position offset
	sLen := 300                     // Square length
	gap := 40                       // gap between the squares

	for _, guess := range guesses {
		tiles, ok := generateTiles(target, guess)
		if !ok {
			return nil, fmt.Errorf("failed to generate color tiles")
		}

		for i, tile := range tiles.Tiles {
			uniform := tilecolorToUniform[tile]

			// Square background
			draw.Draw(img, image.Rect(curPos.X-off, curPos.Y-off, curPos.X+sLen+off, curPos.Y+sLen+off), &uniform, image.Point{}, draw.Src)

			// Put the character
			chr := string(guess[i])
			x := int(drawer.MeasureString(chr) >> 6)
			baseline := curPos.Y + sLen - int(metrics.Descent>>6)
			c.DrawString(chr, freetype.Pt(curPos.X+(sLen-x)/2, baseline))

			// Shift the x coordinate to the right
			curPos.X += sLen + gap
		}

		// Shift the y coordinate to the below
		// and reset the x coordinate to the left
		curPos.Y += sLen + gap
		curPos.X = 150
	}

	// Draw the empty square border
	for range 6 - len(guesses) {
		for range 5 {
			// Top
			draw.Draw(img, image.Rect(curPos.X-off, curPos.Y-off, curPos.X+sLen+off, curPos.Y+off), &BORDER_UNIFORM, image.Point{}, draw.Src)
			// Bottom
			draw.Draw(img, image.Rect(curPos.X-off, curPos.Y+sLen-off, curPos.X+sLen+off, curPos.Y+sLen+off), &BORDER_UNIFORM, image.Point{}, draw.Src)
			// Left
			draw.Draw(img, image.Rect(curPos.X-off, curPos.Y-off, curPos.X+off, curPos.Y+sLen+off), &BORDER_UNIFORM, image.Point{}, draw.Src)
			// Right
			draw.Draw(img, image.Rect(curPos.X+sLen-off, curPos.Y-off, curPos.X+sLen+off, curPos.Y+sLen+off), &BORDER_UNIFORM, image.Point{}, draw.Src)

			// Shift the x coordinate to the right
			curPos.X += sLen + gap
		}

		curPos.Y += sLen + gap
		curPos.X = 150
	}

	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, nil)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
