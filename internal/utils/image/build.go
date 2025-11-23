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

func BuildWebPFromChunksOld(chunks Chunks) ([]byte, error) {
	rawChunks := []byte{}

	_, hasVP8 := chunks["VP8 "]
	_, hasVP8L := chunks["VP8L"]
	_, hasVP8X := chunks["VP8X"]
	_, hasANIM := chunks["ANIM"]
	_, hasANMF := chunks["ANMF"]

	isExtended := hasVP8X
	isAnim := hasANIM && hasANMF

	hasImageData := hasVP8 || hasVP8L || isAnim
	if !hasImageData {
		return nil, ErrNoImageData
	}

	hasDoubleImageData := (hasVP8 && hasVP8L) || (hasVP8 && isAnim) || (hasVP8L && isAnim) // Uhh yeah
	if hasDoubleImageData {
		return nil, ErrDoubleImageData
	}

	for name, payload := range chunks {
		if len(name) != 4 {
			return nil, ErrInvalidChunkNameLength
		}

		// Header and payload
		chunkName := []byte(name)
		chunkSize := numbers.Int32ToByteLSB(len(payload))
		chunk := append(chunkName, chunkSize...)
		chunk = append(chunk, payload...)

		// If VP8X doesn't exists
		if !isExtended {
			// Check if current chunk name is VP8 or VP8L
			if name == "VP8 " || name == "VP8L" {
				// Add VP8 or VP8L to the first position of chunks
				if len(payload)%2 == 1 {
					chunk = append(chunk, 0)
				}
				rawChunks = append(chunk, rawChunks...)
			}

			continue
		}

		// This should be executed if isExtended is true
		// or VP8X chunk is exists

		// Put the VP8X chunk to the first position of chunks
		if name == "VP8X" {
			if len(payload)%2 == 1 {
				chunk = append(chunk, 0)
			}
			rawChunks = append(chunk, rawChunks...)
			continue
		}

		rawChunks = append(rawChunks, chunk...)
		if len(payload)%2 == 1 {
			rawChunks = append(rawChunks, 0) // Add padding if the payload length is not even
		}
	}

	// Get RIFF payload and size
	riffPayload := []byte("WEBP")
	riffPayload = append(riffPayload, rawChunks...)
	riffSize := numbers.Int32ToByteLSB(len(riffPayload))

	// Build the WebP data
	webpData := []byte("RIFF")
	webpData = append(webpData, riffSize...)
	webpData = append(webpData, riffPayload...)

	return webpData, nil
}
