package image

import (
	"kano/internal/utils/numbers"
)

func generateChunk(name string, payload []byte) []byte {
	if len(payload) == 0 {
		return nil
	}

	if len(name) < 4 {
		for range 4 - len(name) {
			name += " "
		}
	} else if len(name) > 4 {
		name = name[:4]
	}

	resBytes := []byte(name)
	resBytes = append(resBytes, numbers.Int32ToByteLSB(len(payload))...)
	resBytes = append(resBytes, payload...)

	if len(payload)%2 == 1 {
		resBytes = append(resBytes, 0) // Add padding if the payload length is not even
	}

	return resBytes
}

func (c WebPChunk) Rebuild() ([]byte, error) {
	return BuildWebPFromChunks(c)
}

func BuildWebPFromChunks(chunks WebPChunk) ([]byte, error) {
	isExtended := chunks.Has("VP8X")

	hasVP8 := chunks.Has("VP8")
	hasVP8L := chunks.Has("VP8L")
	isAnim := chunks.Has("ANIM") && chunks.Has("ANMF")

	hasImageData := hasVP8 || hasVP8L || isAnim
	hasDoubleData := (hasVP8 && hasVP8L) || (hasVP8 && isAnim) || (hasVP8L && isAnim)
	if !hasImageData {
		return nil, ErrNoImageData
	}
	if hasDoubleData {
		return nil, ErrDoubleImageData
	}

	riffPayload := []byte("WEBP")

	if isExtended {
		riffPayload = append(riffPayload, generateChunk("VP8X", chunks.vp8x)...)
	}

	riffPayload = append(riffPayload, generateChunk("VP8 ", chunks.vp8)...)
	riffPayload = append(riffPayload, generateChunk("VP8L", chunks.vp8l)...)
	riffPayload = append(riffPayload, generateChunk("ANIM", chunks.anim)...)
	for _, anmf := range chunks.anmf {
		riffPayload = append(riffPayload, generateChunk("ANMF", anmf)...)
	}
	riffPayload = append(riffPayload, generateChunk("ALPH", chunks.alph)...)
	riffPayload = append(riffPayload, generateChunk("XMP ", chunks.xmp)...)
	riffPayload = append(riffPayload, generateChunk("EXIF", chunks.exif)...)
	riffPayload = append(riffPayload, generateChunk("ICCP", chunks.iccp)...)
	for _, extra := range chunks.extras {
		riffPayload = append(riffPayload, generateChunk(extra.name, extra.payload)...)
	}

	riffSizeByte := numbers.Int32ToByteLSB(len(riffPayload))

	webpBytes := []byte("RIFF")
	webpBytes = append(webpBytes, riffSizeByte...)
	webpBytes = append(webpBytes, riffPayload...)

	return webpBytes, nil
}
