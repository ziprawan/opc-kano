package sticker

import (
	"image"

	"golang.org/x/image/draw"
)

func ForceResize(img image.Image) *image.RGBA {
	// Get current image info
	imgBounds := img.Bounds()
	imgWidth := imgBounds.Dx()
	imgHeight := imgBounds.Dy()

	// Resizing the image into 512x512
	// First, calculate the scale first
	scale := float64(STICKER_SIZE_PX) / float64(max(imgWidth, imgHeight)) // Get the bigger value between image height and width
	// Then calculate the resized width and height
	newWidth := int(float64(imgWidth) * scale)
	newHeight := int(float64(imgHeight) * scale)

	resized := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.BiLinear.Scale(resized, resized.Bounds(), img, imgBounds, draw.Over, nil)

	// New stiker image with RGBA color.
	// If the new width or height is less than STICKER_SIZE_PX, remaining pixel will be transparent
	stk := image.NewRGBA(image.Rect(0, 0, STICKER_SIZE_PX, STICKER_SIZE_PX))

	// Calculate the offset and draw over the transparent rect image
	offsetX := (STICKER_SIZE_PX - newWidth) / 2
	offsetY := (STICKER_SIZE_PX - newHeight) / 2
	draw.Draw(stk, image.Rect(offsetX, offsetY, offsetX+newWidth, offsetY+newHeight), resized, image.Point{}, draw.Over)

	return stk
}
