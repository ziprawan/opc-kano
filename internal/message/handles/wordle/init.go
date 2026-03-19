package wordle

import (
	"os"

	"github.com/golang/freetype/truetype"
)

var vFont *truetype.Font

func getFont() (*truetype.Font, error) {
	if vFont != nil {
		return vFont, nil
	}
	fontBytes, err := os.ReadFile("assets/fonts/ComicRelief-Bold.ttf")
	if err != nil {
		return nil, err
	}

	vFont, err = truetype.Parse(fontBytes)
	if err != nil {
		return nil, err
	}

	return vFont, nil
}
