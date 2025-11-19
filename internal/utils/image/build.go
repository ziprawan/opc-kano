package image

import (
	"kano/internal/utils/numbers"
)

func BuildWebPFromChunks(chunks Chunks) ([]byte, error) {
	rawChunks := []byte{}

	_, isExtended := chunks["VP8X"]
	_, hasImageData := chunks["VP8 "]
	if !hasImageData {
		_, hasImageData = chunks["VP8L"]
	} else if _, hasAnother := chunks["VP8L"]; hasAnother {
		return nil, ErrDoubleImageData
	}

	if !hasImageData {
		return nil, ErrNoImageData
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
