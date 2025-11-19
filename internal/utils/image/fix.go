package image

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
)

var BOOL_STR = map[bool]string{true: "1", false: "0"}

func changeAttr(attr string, idx int, newAttr bool) string {
	changedAttr := attr[:idx] + BOOL_STR[newAttr]
	if idx != len(attr)-1 {
		changedAttr += attr[idx:]
	}
	return changedAttr
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

func FixWebPExtendedChunks(chunks Chunks, imageWidth, imageHeight uint) Chunks {
	vp8x, ok := chunks["VP8X"]
	if !ok || len(vp8x) != 10 {
		vp8x = make([]byte, 10)
	}

	_, hasICCP := chunks["ICCP"]
	_, hasALPH := chunks["ALPH"]
	_, hasEXIF := chunks["EXIF"]
	_, hasXMP := chunks["XMP "]
	_, hasANIM := chunks["ANIM"]

	attr := fmt.Sprintf("%08b", vp8x[0])
	attr = changeAttr(attr, 2, hasICCP)
	attr = changeAttr(attr, 3, hasALPH)
	attr = changeAttr(attr, 4, hasEXIF)
	attr = changeAttr(attr, 5, hasXMP)
	attr = changeAttr(attr, 6, hasANIM)

	attrByte, _ := strconv.ParseUint(attr, 2, 0)
	vp8x[0] = byte(attrByte)

	widthByte := uintTo24Bit(imageWidth - 1)
	heightByte := uintTo24Bit(imageHeight - 1)

	vp8x[4] = widthByte[0]
	vp8x[5] = widthByte[1]
	vp8x[6] = widthByte[2]
	vp8x[7] = heightByte[0]
	vp8x[8] = heightByte[1]
	vp8x[9] = heightByte[2]

	chunks["VP8X"] = vp8x
	return chunks
}
