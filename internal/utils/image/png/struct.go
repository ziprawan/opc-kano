package png

import (
	"encoding/binary"
	"image"
)

type PNGChunk struct {
	Name []byte
	Data []byte
}

type PNG struct {
	Image     []image.Image
	Delay     []int
	LoopCount int
	Disposes  []Dispose
	Blends    []Blend
}

type Dispose int

const (
	APNG_DISPOSE_OP_NONE Dispose = iota
	APNG_DISPOSE_OP_BACKGROUND
	APNG_DISPOSE_OP_PREVIOUS
)

type Blend int

const (
	APNG_BLEND_OP_SOURCE Blend = iota
	APNG_BLEND_OP_OVER
)

type stage int

const (
	seenNone stage = iota
	seenIhdr
	seenPlte
	seenIdat
	seenIend
)

type frameControl struct {
	Seq       uint32  `json:"seq"`
	Width     uint32  `json:"width"`
	Height    uint32  `json:"height"`
	XOffset   uint32  `json:"x_offset"`
	YOffset   uint32  `json:"y_offset"`
	DelayNum  uint16  `json:"delay_num"`
	DelayDen  uint16  `json:"delay_den"`
	DisposeOp Dispose `json:"dispose_op"`
	BlendOp   Blend   `json:"blend_op"`
}

type frame struct {
	Control *frameControl `json:"control"`
	Data    []byte        `json:"data"`
}

type IHDR struct {
	Width             uint32
	Height            uint32
	BitDepth          uint8
	ColorType         uint8
	CompressionMethod uint8
	FilterMethod      uint8
	InterlaceMethod   uint8
}

func (i IHDR) toBytes() []byte {
	res := make([]byte, 13)
	binary.BigEndian.PutUint32(res[:4], i.Width)
	binary.BigEndian.PutUint32(res[4:8], i.Height)
	res[8] = i.BitDepth
	res[9] = i.ColorType
	res[10] = i.CompressionMethod
	res[11] = i.FilterMethod
	res[12] = i.InterlaceMethod

	return res
}
