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
	resBytes = append(resBytes, numbers.Uint32ToByteLSB(uint(len(payload)))...)
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
		riffPayload = append(riffPayload, generateChunk("VP8X", chunks.VP8X.toBytes())...)
	}

	// All chunks necessary for reconstruction and color correction, that is,
	// 'VP8X', 'ICCP', 'ANIM', 'ANMF', 'ALPH', 'VP8 ', and 'VP8L', MUST appear in the order described earlier.
	// Readers SHOULD fail when chunks necessary for reconstruction and color correction are out of order.
	riffPayload = append(riffPayload, generateChunk("ICCP", chunks.ICCP)...)
	riffPayload = append(riffPayload, generateChunk("ANIM", chunks.ANIM.toBytes())...)
	for _, anmf := range chunks.ANMF {
		riffPayload = append(riffPayload, generateChunk("ANMF", anmf.toBytes())...)
	}
	riffPayload = append(riffPayload, generateChunk("ALPH", chunks.ALPH.toBytes())...)
	riffPayload = append(riffPayload, generateChunk("VP8 ", chunks.VP8)...)
	riffPayload = append(riffPayload, generateChunk("VP8L", chunks.VP8L)...)

	riffPayload = append(riffPayload, generateChunk("XMP ", chunks.XMP)...)
	riffPayload = append(riffPayload, generateChunk("EXIF", chunks.EXIF)...)
	for _, extra := range chunks.Extras {
		riffPayload = append(riffPayload, generateChunk(extra.Name, extra.Payload)...)
	}

	riffSizeByte := numbers.Uint32ToByteLSB(uint(len(riffPayload)))

	webpBytes := []byte("RIFF")
	webpBytes = append(webpBytes, riffSizeByte...)
	webpBytes = append(webpBytes, riffPayload...)

	return webpBytes, nil
}
