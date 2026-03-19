package wordle

import (
	"image"
	"image/color"
)

var (
	BORDER_UNIFORM = image.Uniform{color.RGBA{211, 214, 218, 255}}
	GRAY_UNIFORM   = image.Uniform{color.RGBA{120, 124, 126, 255}}
	YELLOW_UNIFORM = image.Uniform{color.RGBA{201, 180, 88, 255}}
	GREEN_UNIFORM  = image.Uniform{color.RGBA{106, 170, 100, 255}}
)

var tilecolorToUniform = map[tilecolor]image.Uniform{
	gray:   GRAY_UNIFORM,
	yellow: YELLOW_UNIFORM,
	green:  GREEN_UNIFORM,
}
