package image

import (
	"fmt"
	"kano/internal/utils/numbers"
)

type WebPImageSize struct {
	Width  uint16
	Height uint16
}

type OtherChunk struct {
	name    string
	payload []byte
}

type WebPChunk struct {
	vp8x []byte
	vp8  []byte
	vp8l []byte

	iccp []byte
	exif []byte
	xmp  []byte
	alph []byte

	anim []byte
	anmf [][]byte

	extras []OtherChunk
}

func normalizeChunkName(name string) string {
	if len(name) < 4 {
		for range 4 - len(name) {
			name += " "
		}
	} else if len(name) > 4 {
		name = name[:4]
	}

	return name
}

func (c WebPChunk) GetChunk(name string) [][]byte {
	name = normalizeChunkName(name)
	switch name {
	case "VP8 ":
		return [][]byte{c.vp8}
	case "VP8X":
		return [][]byte{c.vp8x}
	case "VP8L":
		return [][]byte{c.vp8l}
	case "ANIM":
		return [][]byte{c.anim}
	case "ANMF":
		return c.anmf
	case "ICCP":
		return [][]byte{c.iccp}
	case "EXIF":
		return [][]byte{c.exif}
	case "XMP ":
		return [][]byte{c.xmp}
	case "ALPH":
		return [][]byte{c.alph}
	default:
		found := [][]byte{}
		for _, val := range c.extras {
			if val.name == name {
				found = append(found, val.payload)
			}
		}
		return found
	}
}

func (c *WebPChunk) SetChunk(name string, payload []byte, force ...bool) error {
	isForced := len(force) == 1 && force[0]
	name = normalizeChunkName(name)
	if len(payload) == 0 {
		return ErrEmptyPayload
	}

	fmt.Println("Chunk name:", name)

	switch name {
	case "VP8 ":
		if len(c.vp8) != 0 && !isForced {
			return ErrHasPayload
		}
		c.vp8 = payload
	case "VP8X":
		fmt.Println(payload)
		if len(c.vp8x) != 0 && !isForced {
			return ErrHasPayload
		}
		c.vp8x = payload
	case "VP8L":
		if len(c.vp8l) != 0 && !isForced {
			return ErrHasPayload
		}
		c.vp8l = payload
	case "ANIM":
		if len(c.anim) != 0 && !isForced {
			return ErrHasPayload
		}
		c.anim = payload
	case "ANMF":
		c.anmf = append(c.anmf, payload)
	case "ICCP":
		if len(c.iccp) == 0 || isForced {
			c.iccp = payload
		}
	case "EXIF":
		if len(c.exif) == 0 || isForced {
			c.exif = payload
		}
	case "XMP ":
		if len(c.xmp) == 0 || isForced {
			c.xmp = payload
		}
	case "ALPH":
		if len(c.alph) == 0 || isForced {
			c.alph = payload
		}
	default:
		c.extras = append(c.extras, OtherChunk{name: name, payload: payload})
	}

	return nil
}

func (c WebPChunk) Has(name string) bool {
	name = normalizeChunkName(name)
	switch name {
	case "VP8 ":
		return len(c.vp8) != 0
	case "VP8X":
		return len(c.vp8x) != 0
	case "VP8L":
		return len(c.vp8l) != 0
	case "ANIM":
		return len(c.anim) != 0
	case "ANMF":
		return len(c.anmf) != 0
	case "ICCP":
		return len(c.iccp) != 0
	case "EXIF":
		return len(c.exif) != 0
	case "XMP ":
		return len(c.xmp) != 0
	case "ALPH":
		return len(c.alph) != 0
	default:
		for _, val := range c.extras {
			if val.name == name {
				return true
			}
		}
		return false
	}
}

func (c WebPChunk) GetImageSize() WebPImageSize {
	var width, height uint16

	if len(c.vp8l) != 0 {
		threeBytes := uint16(numbers.ByteToUint32LSB(c.vp8l[1:4]))
		width = threeBytes & 0x3fff
		height = (threeBytes >> 14) & 0x3fff
	} else if len(c.vp8) != 0 {
		width = uint16(numbers.ByteToUint16LSB(c.vp8[6:8])) & 0x3fff
		height = uint16(numbers.ByteToUint16LSB(c.vp8[8:10])) & 0x3fff
	}

	return WebPImageSize{
		Width:  width,
		Height: height,
	}
}
