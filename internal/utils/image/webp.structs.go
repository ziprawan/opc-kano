package image

import (
	"fmt"
	"image/color"
	"kano/internal/utils/numbers"
)

type WebPImageSize struct {
	Width  uint16
	Height uint16
}

type OtherChunk struct {
	Name    string
	Payload []byte
}

type WebPChunk struct {
	VP8X *VP8X
	VP8  []byte
	VP8L []byte

	ICCP []byte
	EXIF []byte
	XMP  []byte
	ALPH *ALPH

	ANIM *ANIM
	ANMF []ANMF

	Extras []OtherChunk
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
		return [][]byte{c.VP8}
	case "VP8X":
		return [][]byte{c.VP8X.toBytes()}
	case "VP8L":
		return [][]byte{c.VP8L}
	case "ANIM":
		return [][]byte{c.ANIM.toBytes()}
	case "ANMF":
		anmfs := [][]byte{}
		for _, anmf := range c.ANMF {
			anmfs = append(anmfs, anmf.toBytes())
		}
		return anmfs
	case "ICCP":
		return [][]byte{c.ICCP}
	case "EXIF":
		return [][]byte{c.EXIF}
	case "XMP ":
		return [][]byte{c.XMP}
	case "ALPH":
		return [][]byte{c.ALPH.toBytes()}
	default:
		found := [][]byte{}
		for _, val := range c.Extras {
			if val.Name == name {
				found = append(found, val.Payload)
			}
		}
		return found
	}
}

func (c *WebPChunk) SetChunk(name string, payload []byte, force ...bool) (err error) {
	isForced := len(force) == 1 && force[0]
	name = normalizeChunkName(name)
	if len(payload) == 0 {
		return ErrEmptyPayload
	}

	fmt.Println("Chunk name:", name)

	switch name {
	case "VP8 ":
		if c.Has("VP8 ") && !isForced {
			return ErrHasPayload
		}
		c.VP8 = payload
	case "VP8X":
		if c.Has("VP8X") && !isForced {
			return ErrHasPayload
		}
		c.VP8X, err = toVP8X(payload)
	case "VP8L":
		if c.Has("VP8L") && !isForced {
			return ErrHasPayload
		}
		c.VP8L = payload
	case "ANIM":
		if c.Has("ANIM") && !isForced {
			return ErrHasPayload
		}
		c.ANIM, err = toANIM(payload)
	case "ANMF":
		anmf, err := toANMF(payload)
		if err != nil {
			return err
		}
		c.ANMF = append(c.ANMF, *anmf)
	case "ICCP":
		if len(c.ICCP) == 0 || isForced {
			c.ICCP = payload
		}
	case "EXIF":
		if len(c.EXIF) == 0 || isForced {
			c.EXIF = payload
		}
	case "XMP ":
		if len(c.XMP) == 0 || isForced {
			c.XMP = payload
		}
	case "ALPH":
		if c.ALPH == nil || isForced {
			c.ALPH, err = toALPH(payload)
		}
	default:
		c.Extras = append(c.Extras, OtherChunk{Name: name, Payload: payload})
	}

	return nil
}

func (c WebPChunk) Has(name string) bool {
	name = normalizeChunkName(name)
	switch name {
	case "VP8 ":
		return len(c.VP8) != 0
	case "VP8X":
		return c.VP8X != nil
	case "VP8L":
		return len(c.VP8L) != 0
	case "ANIM":
		return c.ANIM != nil
	case "ANMF":
		return len(c.ANMF) != 0
	case "ICCP":
		return len(c.ICCP) != 0
	case "EXIF":
		return len(c.EXIF) != 0
	case "XMP ":
		return len(c.XMP) != 0
	case "ALPH":
		return c.ALPH != nil
	default:
		for _, val := range c.Extras {
			if val.Name == name {
				return true
			}
		}
		return false
	}
}

func (c WebPChunk) GetImageSize() WebPImageSize {
	var width, height uint16

	if len(c.VP8L) != 0 {
		threeBytes := uint16(numbers.ByteToUint32LSB(c.VP8L[1:4]))
		width = threeBytes & 0x3fff
		height = (threeBytes >> 14) & 0x3fff
	} else if len(c.VP8) != 0 {
		width = uint16(numbers.ByteToUint16LSB(c.VP8[6:8])) & 0x3fff
		height = uint16(numbers.ByteToUint16LSB(c.VP8[8:10])) & 0x3fff
	} else if len(c.ANMF) != 0 {
		for _, anmf := range c.ANMF {
			w := anmf.X + anmf.W
			h := anmf.Y + anmf.H
			if w > uint(width) {
				width = uint16(w)
			}
			if h > uint(height) {
				height = uint16(h)
			}
		}
	}

	return WebPImageSize{
		Width:  width,
		Height: height,
	}
}

type uint24 uint32

func (u uint24) Clamp() uint24 {
	max := uint24((1 << 24) - 1)
	if u > max {
		return uint24(max)
	}

	return u
}

type VP8X struct {
	HasICCProfile bool
	HasAlpha      bool
	HasExif       bool
	HasXMP        bool
	HasAnimation  bool

	Width  uint
	Height uint
}

func toVP8X(bytes []byte) (*VP8X, error) {
	if len(bytes) != 10 {
		return nil, fmt.Errorf("VP8X byte length must be 10")
	}

	props := bytes[0]
	res := VP8X{
		HasICCProfile: readBit(props, 5) > 0,
		HasAlpha:      readBit(props, 4) > 0,
		HasExif:       readBit(props, 3) > 0,
		HasXMP:        readBit(props, 2) > 0,
		HasAnimation:  readBit(props, 1) > 0,

		Width:  bytesToUint24(bytes[4:7]) + 1,
		Height: bytesToUint24(bytes[7:10]) + 1,
	}

	return &res, nil
}

func (v *VP8X) toBytes() []byte {
	if v == nil {
		return nil
	}

	res := make([]byte, 10)
	writeBit(&res[0], 5, v.HasICCProfile)
	writeBit(&res[0], 4, v.HasAlpha)
	writeBit(&res[0], 3, v.HasExif)
	writeBit(&res[0], 2, v.HasXMP)
	writeBit(&res[0], 1, v.HasAnimation)

	widthByte := uint24ToBytes(uint(v.Width) - 1)
	heightByte := uint24ToBytes(uint(v.Height) - 1)
	res[4] = widthByte[0]
	res[5] = widthByte[1]
	res[6] = widthByte[2]
	res[7] = heightByte[0]
	res[8] = heightByte[1]
	res[9] = heightByte[2]

	return res
}

// For an animated image, this chunk contains the global parameters of the animation.
//
// This chunk MUST appear if the Animation flag in the 'VP8X' Chunk is set.
// If the Animation flag is not set and this chunk is present, it MUST be ignored.
type ANIM struct {
	// The default background color of the canvas in [Blue, Green, Red, Alpha] byte order.
	// This color MAY be used to fill the unused space on the canvas around the frames,
	// as well as the transparent pixels of the first frame.
	// The background color is also used when the Disposal method is 1.
	BackgroundColor color.RGBA

	// The number of times to loop the animation. If it is 0, this means infinitely.
	LoopCount uint16
}

func toANIM(bytes []byte) (*ANIM, error) {
	if len(bytes) != 6 {
		return nil, fmt.Errorf("ANIM byte length must be 6")
	}

	res := ANIM{
		BackgroundColor: color.RGBA{
			B: bytes[0],
			G: bytes[1],
			R: bytes[2],
			A: bytes[3],
		},
		LoopCount: numbers.ByteToUint16LSB(bytes[4:]),
	}

	return &res, nil
}

func (a *ANIM) toBytes() []byte {
	if a == nil {
		return nil
	}

	res := make([]byte, 6)
	res[0] = a.BackgroundColor.B
	res[1] = a.BackgroundColor.G
	res[2] = a.BackgroundColor.R
	res[3] = a.BackgroundColor.A
	loop := numbers.Uint16ToByteLSB(uint(a.LoopCount))
	res[4] = loop[0]
	res[5] = loop[1]

	return res
}

// For animated images, this chunk contains information about a single frame.
// If the Animation flag is not set, then this chunk SHOULD NOT be present.
type ANMF struct {
	// The X coordinate of the upper left corner of the frame
	X uint
	// The Y coordinate of the upper left corner of the frame
	Y uint
	// The 1-based width of the frame
	W uint
	// The 1-based height of the frame
	H uint
	// The time to wait before displaying the next frame, in 1-millisecond units.
	// Note that the interpretation of the Frame Duration of 0 (and often <= 10) is defined by the implementation.
	// Many tools and browsers assign a minimum duration similar to GIF.
	D uint

	// Indicates how transparent pixels of the current frame are to be blended with corresponding pixels of the previous canvas.
	//
	// 0: Use alpha-blending. After disposing of the previous frame, render the current frame on the canvas using alpha-blending (see below).
	// If the current frame does not have an alpha channel, assume the alpha value is 255, effectively replacing the rectangle.
	//
	// 1: Do not blend. After disposing of the previous frame,
	// render the current frame on the canvas by overwriting the rectangle convered by the current frame.
	BlendingMethod bool

	// Indicates how the current frame is to be treated after it has been displayed
	// (before rendering the next frame) on the canvas:
	//
	// 0: Do not dispose. Leave the canvas as is.
	//
	// 1: Dispose to the background color. Fill the rectangle on the canvas covered by the current frame
	// with the background color specified in the 'ANIM' chunk
	DisposalMethod bool

	ALPH          *ALPH
	VP8           []byte
	VP8L          []byte
	UnknownChunks []OtherChunk
}

func toANMF(bytes []byte) (*ANMF, error) {
	res := ANMF{}

	x := bytesToUint24(bytes[0:3])
	y := bytesToUint24(bytes[3:6])
	w := bytesToUint24(bytes[6:9])
	h := bytesToUint24(bytes[9:12])
	d := bytesToUint24(bytes[12:15])
	res.X = x * 2
	res.Y = y * 2
	res.W = w + 1
	res.H = h + 1
	res.D = d

	res.BlendingMethod = readBit(bytes[15], 1) > 0
	res.DisposalMethod = readBit(bytes[15], 0) > 0

	rawChunks := bytes[16:]
	for {
		// The remaining bytes is less than 8
		// Which is insufficient for chunk header
		if len(rawChunks) < 8 {
			if len(rawChunks) != 0 {
				return nil, ErrInvalidChunkLength
			}
			// Except if it was 0, then stop the loop
			break
		}

		// Take the chunk header (name and size)
		chunkHeader := rawChunks[:8]
		chunkName := chunkHeader[:4]
		chunkSize := numbers.ByteToUint32LSB(chunkHeader[4:])

		// remainingBytes is a rawChunks with a header cut off
		remainingBytes := rawChunks[8:]

		// Chunk name must be printable ASCII characters
		if !ValidateChunkName(chunkName) {
			return nil, ErrInvalidChunkName
		}
		// The remaining bytes should be equal or more than the specified chunk size
		if len(remainingBytes) < int(chunkSize) {
			return nil, ErrInvalidChunkLength
		}

		// Take the payload from remainingBytes from 0 until the chunkSize
		chunkPayload := remainingBytes[:chunkSize]
		chunkNameStr := string(chunkName)

		switch chunkNameStr {
		case "ALPH":
			var err error
			res.ALPH, err = toALPH(chunkPayload)
			if err != nil {
				return nil, err
			}
		case "VP8 ":
			if len(res.VP8L) != 0 {
				return nil, fmt.Errorf("alph: VP8L bitstream exists")
			}
			res.VP8 = chunkPayload
		case "VP8L":
			if len(res.VP8) != 0 {
				return nil, fmt.Errorf("alph: VP8  bitstream exists")
			}
			res.VP8L = chunkPayload
			res.ALPH = nil
		default:
			res.UnknownChunks = append(res.UnknownChunks, OtherChunk{
				Name:    chunkNameStr,
				Payload: chunkPayload,
			})
		}

		// Remove padding if the length is not even
		if chunkSize%2 == 1 {
			chunkSize += 1
		}
		rawChunks = remainingBytes[chunkSize:]
	}

	if res.ALPH != nil && len(res.VP8) == 0 {
		return nil, fmt.Errorf("alph: no image bitstream given")
	}
	if len(res.VP8) == 0 && len(res.VP8L) == 0 {
		return nil, fmt.Errorf("alph: no image bitstream given")
	}

	return &res, nil
}

func (a *ANMF) toBytes() []byte {
	res := make([]byte, 16)

	x := uint24ToBytes(a.X / 2)
	y := uint24ToBytes(a.Y / 2)
	w := uint24ToBytes(a.W - 1)
	h := uint24ToBytes(a.H - 1)
	d := uint24ToBytes(a.D)

	res[0] = x[0]
	res[1] = x[1]
	res[2] = x[2]
	res[3] = y[0]
	res[4] = y[1]
	res[5] = y[2]
	res[6] = w[0]
	res[7] = w[1]
	res[8] = w[2]
	res[9] = h[0]
	res[10] = h[1]
	res[11] = h[2]
	res[12] = d[0]
	res[13] = d[1]
	res[14] = d[2]

	writeBit(&res[15], 0, a.DisposalMethod)
	writeBit(&res[15], 1, a.BlendingMethod)

	if a.ALPH != nil {
		res = append(res, generateChunk("ALPH", a.ALPH.toBytes())...)
	}
	if len(a.VP8) != 0 {
		res = append(res, generateChunk("VP8 ", a.VP8)...)
	} else if len(a.VP8L) != 0 {
		res = append(res, generateChunk("VP8L", a.VP8L)...)
	}

	return res
}

type ALPH struct {
	Preprocessing   Preprocessing
	FilteringMethod FilterMethod
	Compression     Compression
	Bitstream       []byte
}

func toALPH(bytes []byte) (*ALPH, error) {
	res := ALPH{}

	if readBits(bytes[0], 6, 7) != 0 {
		return nil, fmt.Errorf("alph: reserved bit is not 0")
	}

	P := readBits(bytes[0], 4, 5)
	F := readBits(bytes[0], 2, 3)
	C := readBits(bytes[0], 0, 1)

	switch P {
	case 0:
		res.Preprocessing = PreprocessingNone
	case 1:
		res.Preprocessing = PreprocessingReduction
	default:
		return nil, fmt.Errorf("alph: unknown preprocessing %d", P)
	}

	switch F {
	case 0:
		res.FilteringMethod = FilterNone
	case 1:
		res.FilteringMethod = FilterHorizontal
	case 2:
		res.FilteringMethod = FilterVertical
	case 3:
		res.FilteringMethod = FilterGradient
	default:
		return nil, fmt.Errorf("alph: unknown filtering method %d", F)
	}

	switch C {
	case 0:
		res.Compression = CompressionNone
	case 1:
		res.Compression = CompressionWebPLossless
	default:
		return nil, fmt.Errorf("alph: unknown compression method %d", C)
	}

	res.Bitstream = bytes[1:]

	return &res, nil
}

func (a *ALPH) toBytes() []byte {
	if a == nil {
		return nil
	}

	res := make([]byte, 1, len(a.Bitstream)+1)

	if a.Preprocessing == PreprocessingReduction {
		writeBit(&res[0], 4, true)
	}
	if a.Compression == CompressionWebPLossless {
		writeBit(&res[0], 0, true)
	}

	if a.FilteringMethod == FilterHorizontal || a.FilteringMethod == FilterGradient {
		writeBit(&res[0], 2, true)
	}
	if a.FilteringMethod == FilterVertical || a.FilteringMethod == FilterGradient {
		writeBit(&res[0], 3, true)
	}

	res = append(res, a.Bitstream...)

	return res
}

type Preprocessing uint

const (
	PreprocessingNone Preprocessing = iota
	PreprocessingReduction
)

type FilterMethod uint

const (
	FilterNone FilterMethod = iota
	FilterHorizontal
	FilterVertical
	FilterGradient
)

type Compression uint

const (
	CompressionNone Compression = iota
	CompressionWebPLossless
)
