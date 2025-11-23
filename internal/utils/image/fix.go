package image

import (
	"encoding/binary"
	"kano/internal/utils/numbers"
	"math"
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

func uintTo24Bit(num uint) [3]byte {
	var MAX = uint(math.Pow(2, 24))
	if num >= MAX {
		num = MAX
	}

	numByte := make([]byte, 4)
	binary.LittleEndian.PutUint32(numByte, uint32(num))

	res := [3]byte{}
	res[0] = numByte[0]
	res[1] = numByte[1]
	res[2] = numByte[2]

	return res
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
	if riffPayloadSize != len(riffPayload) {
		newRiffBytes := []byte("RIFF")
		newRiffBytes = append(newRiffBytes, numbers.Int32ToByteLSB(len(riffPayload))...)
		newRiffBytes = append(newRiffBytes, riffPayload...)

		return newRiffBytes, nil
	} else {
		return riffBytes, nil
	}
}

// Fix the VP8X data in the WebPChunk
func FixWebPExtendedChunks(chunks WebPChunk) WebPChunk {
	vp8x := chunks.vp8x
	if len(vp8x) == 0 {
		// We might not need VP8X
		return chunks
	}
	if len(vp8x) != 10 {
		vp8x = make([]byte, 10)
	}

	hasICCP := len(chunks.iccp) != 0
	hasALPH := len(chunks.alph) != 0
	hasEXIF := len(chunks.exif) != 0
	hasXMP := len(chunks.xmp) != 0
	hasANIM := len(chunks.anim) != 0 && len(chunks.anmf) != 0

	vp8x[0] = changeAttr(vp8x[0], 1, hasANIM)
	vp8x[0] = changeAttr(vp8x[0], 2, hasXMP)
	vp8x[0] = changeAttr(vp8x[0], 3, hasEXIF)
	vp8x[0] = changeAttr(vp8x[0], 4, hasALPH)
	vp8x[0] = changeAttr(vp8x[0], 5, hasICCP)

	size := chunks.GetImageSize()
	widthByte := uintTo24Bit(uint(size.Width) - 1)
	heightByte := uintTo24Bit(uint(size.Height) - 1)

	vp8x[4] = widthByte[0]
	vp8x[5] = widthByte[1]
	vp8x[6] = widthByte[2]
	vp8x[7] = heightByte[0]
	vp8x[8] = heightByte[1]
	vp8x[9] = heightByte[2]

	chunks.vp8x = vp8x
	return chunks
}
