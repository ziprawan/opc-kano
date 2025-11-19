package image

import (
	"kano/internal/utils/numbers"
)

type Chunks = map[string][]byte

// Chunk name must be printable ASCII characters
func ValidateChunkName(name []byte) bool {
	for _, rn := range name {
		// Must be in range 0x20 - 0x7E
		if rn < 0x20 || rn > 0x7e {
			return false
		}
	}

	return true
}

func ExtractChunksFromWebP(webpData []byte) (Chunks, error) {
	// Check the first twelve bytes of webp data
	// First four byte must be "RIFF"
	// Second four byte represent the RIFF payload size
	riffHeader := webpData[:8]
	if string(riffHeader[:4]) != "RIFF" {
		return nil, ErrNotRIFF
	}

	// Take the payload from offest 8 and compare the length
	riffPayload := webpData[8:]
	riffPayloadSize := numbers.ByteToInt32LSB(webpData[4:8])
	if riffPayloadSize != len(riffPayload) {
		return nil, ErrInvalidRIFFLength
	}

	// Next, from [8:12] must be "WEBP"
	if string(riffPayload[:4]) != "WEBP" {
		return nil, ErrNotWebP
	}

	// Take the chunk payloads
	rawChunks := riffPayload[4:]
	chunks := Chunks{}

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
		chunkSize := numbers.ByteToInt32LSB(chunkHeader[4:])

		// remainingBytes is a rawChunks with a header cut off
		remainingBytes := rawChunks[8:]

		// Chunk name must be printable ASCII characters
		if !ValidateChunkName(chunkName) {
			return nil, ErrInvalidChunkName
		}
		// The remaining bytes should be equal or more than the specified chunk size
		if len(remainingBytes) < chunkSize {
			return nil, ErrInvalidChunkLength
		}

		// Take the payload from remainingBytes from 0 until the chunkSize
		chunkPayload := remainingBytes[:chunkSize]
		chunks[string(chunkName)] = chunkPayload

		// Remove padding if the length is not even
		if chunkSize%2 == 1 {
			chunkSize += 1
		}
		rawChunks = remainingBytes[chunkSize:]
	}

	return chunks, nil
}
