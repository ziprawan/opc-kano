package image

import (
	"kano/internal/utils/numbers"
)

var BOOL_STR = map[bool]string{true: "1", false: "0"}

func changeAttr(attr byte, idx int, newAttr bool) byte {
	if newAttr {
		attr |= (1 << idx)
	} else {
		attr &= (0xff - (1 << idx))
	}

	return attr
}

// This will assume that the riff payload is ALL VALID.
// Just blindly take the payload length and put it in the RIFF header
func FixRIFFHeader(riffBytes []byte) ([]byte, error) {
	riffHeader := riffBytes[:8]
	if string(riffHeader[:4]) != "RIFF" {
		return nil, ErrNotRIFF
	}

	riffPayload := riffBytes[8:]
	riffPayloadSize := numbers.ByteToUint32LSB(riffBytes[4:8])
	if int(riffPayloadSize) != len(riffPayload) {
		newRiffBytes := []byte("RIFF")
		newRiffBytes = append(newRiffBytes, numbers.Uint32ToByteLSB(uint(len(riffPayload)))...)
		newRiffBytes = append(newRiffBytes, riffPayload...)

		return newRiffBytes, nil
	} else {
		return riffBytes, nil
	}
}

// Fix the VP8X data in the WebPChunk
func FixWebPExtendedChunks(chunks WebPChunk) WebPChunk {
	vp8x := chunks.VP8X
	if !chunks.Has("VP8X") {
		vp8x = &VP8X{}
	}

	vp8x.HasICCProfile = chunks.Has("ICCP")
	vp8x.HasAlpha = chunks.Has("ALPH")
	vp8x.HasExif = chunks.Has("EXIF")
	vp8x.HasXMP = chunks.Has("XMP ")
	vp8x.HasAnimation = chunks.Has("ANIM") && chunks.Has("ANMF")

	size := chunks.GetImageSize()
	vp8x.Width = uint(size.Width)
	vp8x.Height = uint(size.Height)

	chunks.VP8X = vp8x
	return chunks
}
