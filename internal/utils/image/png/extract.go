package png

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"image"
	"image/draw"
	"image/png"
	"os"
	"slices"
)

var pngMagic = []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}

// MSB btw
func toUint(b []byte) uint {
	if len(b) > 8 {
		b = b[:8]
	}

	var res uint = 0
	for _, v := range b {
		res <<= 8
		res |= uint(v)
	}

	return res
}

func compareBytes(a []byte, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if b[i] != v {
			return false
		}
	}

	return true
}

func ExtractChunks(b []byte) ([]PNGChunk, error) {
	if !compareBytes(b[:8], pngMagic) {
		return nil, fmt.Errorf("invalid header")
	}

	chunks := []PNGChunk{}
	rawChunks := b[8:]
	for {
		if len(rawChunks) == 0 {
			break
		}
		// 4 for size, 4 for type, 4 for crc
		if len(rawChunks) < 12 {
			return nil, fmt.Errorf("broken chunk")
		}

		chunkSize := toUint(rawChunks[:4])
		chunkName := rawChunks[4:8]
		chunkData := rawChunks[8 : 8+chunkSize]
		chunkCRC := toUint(rawChunks[8+chunkSize : 12+chunkSize])
		if len(chunkData) != int(chunkSize) {
			return nil, fmt.Errorf("chunk has invalid size")
		}

		crc := crc32.Checksum(append(chunkName, chunkData...), crc32.IEEETable)
		if crc != uint32(chunkCRC) {
			return nil, fmt.Errorf("corrupted chunk, invalid CRC")
		}

		chunks = append(chunks, PNGChunk{
			Name: chunkName,
			Data: chunkData,
		})
		rawChunks = rawChunks[12+chunkSize:]

		if len(chunks) == 0 && compareBytes(chunkName, []byte{0x49, 0x48, 0x44, 0x52}) {
			return nil, fmt.Errorf("PNG not started with IHDR")
		}
		if compareBytes(chunkName, []byte{0x49, 0x45, 0x4e, 0x44}) && len(rawChunks) > 0 {
			rawChunks = []byte{}
		}
	}

	return chunks, nil
}

func buildChunk(c PNGChunk) []byte {
	chunkLength := make([]byte, 4)
	chunkCRC := make([]byte, 4)
	crc := crc32.Checksum(append(c.Name, c.Data...), crc32.IEEETable)
	binary.BigEndian.PutUint32(chunkLength, uint32(len(c.Data)))
	binary.BigEndian.PutUint32(chunkCRC, crc)

	return slices.Concat(chunkLength, c.Name, c.Data, chunkCRC)
}

func createTmpFrame(i IHDR, otherHeaders []PNGChunk, frameData frame) (image.Image, error) {
	res := make([]byte, 8)
	copy(res, pngMagic)

	if frameData.Control != nil {
		i.Width = frameData.Control.Width
		i.Height = frameData.Control.Height
	}

	res = append(res, buildChunk(PNGChunk{Name: []byte("IHDR"), Data: i.toBytes()})...)
	for _, ch := range otherHeaders {
		res = append(res, buildChunk(ch)...)
	}

	res = append(res, buildChunk(PNGChunk{Name: []byte("IDAT"), Data: frameData.Data})...)
	res = append(res, buildChunk(PNGChunk{Name: []byte("IEND"), Data: []byte{}})...)

	return png.Decode(bytes.NewReader(res))
}

// This doesn't have field checking for color type, bit depth, etc.
func parseIHDR(b []byte) (res IHDR) {
	if len(b) != 13 {
		return
	}

	res.Width = binary.BigEndian.Uint32(b[:4])
	res.Height = binary.BigEndian.Uint32(b[4:8])
	res.BitDepth = b[8]
	res.ColorType = b[9]
	res.CompressionMethod = b[10]
	res.FilterMethod = b[11]
	res.InterlaceMethod = b[12]

	return
}

func DecodeAll(b []byte) (*PNG, error) {
	chunks, err := ExtractChunks(b)
	if err != nil {
		return nil, err
	}

	otherHeaders := []PNGChunk{}
	png := PNG{}
	var currentControl *frameControl
	frames := []frame{}

	actualIHDR := IHDR{}
	seenfcTL := false
	stage := seenNone
	currentSeq := -1

	for _, chunk := range chunks {
		if stage == seenIend {
			break
		}

		s := string(chunk.Name)
		switch s {
		case "IHDR":
			if stage >= seenIhdr {
				return nil, fmt.Errorf("invalid IHDR order")
			}
			stage = seenIhdr
			actualIHDR = parseIHDR(chunk.Data)

		case "eXIf", "pHYs", "sPLT":
			if stage >= seenIdat {
				return nil, fmt.Errorf("chunk %s must come before IDAT", s)
			}
			otherHeaders = append(otherHeaders, chunk)

		case "acTL":
			fmt.Printf("Expected frame numbers: %d\n", binary.BigEndian.Uint32(chunk.Data[:4]))
			png.LoopCount = int(binary.BigEndian.Uint32(chunk.Data[4:8]))

		case "fcTL":
			if stage < seenIdat {
				if seenfcTL {
					return nil, fmt.Errorf("chunk fcTL can only occur once before IDAT")
				} else {
					seenfcTL = true
				}
			}
			currentControl = &frameControl{}
			currentControl.Seq = binary.BigEndian.Uint32(chunk.Data[:4])
			currentControl.Width = binary.BigEndian.Uint32(chunk.Data[4:8])
			currentControl.Height = binary.BigEndian.Uint32(chunk.Data[8:12])
			currentControl.XOffset = binary.BigEndian.Uint32(chunk.Data[12:16])
			currentControl.YOffset = binary.BigEndian.Uint32(chunk.Data[16:20])
			currentControl.DelayNum = binary.BigEndian.Uint16(chunk.Data[20:22])
			currentControl.DelayDen = binary.BigEndian.Uint16(chunk.Data[22:24])
			currentControl.DisposeOp = Dispose(chunk.Data[24])
			currentControl.BlendOp = Blend(chunk.Data[25])

			currentSeq++
			if currentControl.Seq != uint32(currentSeq) {
				return nil, fmt.Errorf("unexpected frame sequence")
			}

		case "cHRM", "cICP", "gAMA", "iCCP", "mDCV", "cLLI", "sBIT", "sRGB":
			if stage >= seenPlte {
				return nil, fmt.Errorf("chunk %s must come before PLTE", s)
			}
			otherHeaders = append(otherHeaders, chunk)

		case "PLTE":
			if stage >= seenPlte || stage == seenNone {
				return nil, fmt.Errorf("invalid PLTE order")
			}
			stage = seenPlte
			otherHeaders = append(otherHeaders, chunk)

		case "bKGD", "hIST", "tRNS":
			if stage != seenPlte {
				return nil, fmt.Errorf("chunk %s must come after PLTE and before IDAT", s)
			}
			otherHeaders = append(otherHeaders, chunk)

		case "IDAT":
			if stage > seenIdat || stage == seenNone {
				return nil, fmt.Errorf("invalid IDAT order")
			}
			stage = seenIdat
			frames = append(frames, frame{
				Control: currentControl,
				Data:    chunk.Data,
			})

		case "fdAT":
			if stage < seenIdat {
				return nil, fmt.Errorf("chunk fdAT must come after IDAT")
			}
			currentSeq++
			seq := binary.BigEndian.Uint32(chunk.Data[:4])
			if seq != uint32(currentSeq) {
				return nil, fmt.Errorf("unexpected frame sequence")
			}
			frames = append(frames, frame{
				Control: currentControl,
				Data:    chunk.Data[4:],
			})

		case "IEND":
			stage = seenIend

		default:
			continue
		}
	}

	GenerateSeqImages(actualIHDR, otherHeaders, frames)

	return &png, nil
}

func clone(img *image.NRGBA) *image.NRGBA {
	dst := image.NewNRGBA(img.Bounds())
	copy(dst.Pix, img.Pix)
	return dst
}

func GenerateSeqImages(actualIHDR IHDR, header []PNGChunk, frames []frame) {
	var currentFrame image.Image

	for i, frame := range frames {
		curImg, err := createTmpFrame(actualIHDR, header, frame)
		if err != nil {
			panic(err)
		}
		var byteBuffer bytes.Buffer

		if currentFrame == nil {
			currentFrame = curImg
			err := png.Encode(&byteBuffer, curImg)
			if err != nil {
				panic(err)
			}
		} else {
			var prevCanvas *image.NRGBA
			canvas := image.NewNRGBA(currentFrame.Bounds())
			draw.Draw(canvas, currentFrame.Bounds(), currentFrame, image.Point{}, draw.Src)
			if frame.Control.DisposeOp == APNG_DISPOSE_OP_PREVIOUS {
				prevCanvas = clone(canvas)
				var bb bytes.Buffer
				png.Encode(&bb, prevCanvas)
				os.WriteFile(fmt.Sprintf("prev_%03d.png", i), bb.Bytes(), 0644)
			}
			r := image.Rect(
				int(frame.Control.XOffset),
				int(frame.Control.YOffset),
				int(frame.Control.XOffset+frame.Control.Width),
				int(frame.Control.YOffset+frame.Control.Height),
			)
			background := image.NewNRGBA(r)
			for i := range background.Pix {
				background.Pix[i] = 0
			}

			op := draw.Src
			if frame.Control.BlendOp == APNG_BLEND_OP_OVER {
				op = draw.Over
			}

			draw.Draw(canvas, r, curImg, image.Point{}, op)
			err := png.Encode(&byteBuffer, canvas)
			if err != nil {
				panic(err)
			}

			fmt.Println(i, "Dispose:", frame.Control.DisposeOp)
			switch frame.Control.DisposeOp {
			case APNG_DISPOSE_OP_NONE:
				// do nothing

			case APNG_DISPOSE_OP_BACKGROUND:
				draw.Draw(canvas, r, background, image.Point{}, draw.Src)

			case APNG_DISPOSE_OP_PREVIOUS:
				canvas = prevCanvas
			}

			currentFrame = canvas
		}

		var bb bytes.Buffer
		png.Encode(&bb, currentFrame)
		os.WriteFile(fmt.Sprintf("cur_%03d.png", i), bb.Bytes(), 0644)
		os.WriteFile(fmt.Sprintf("frame_%03d.png", i), byteBuffer.Bytes(), 0644)
	}
}
