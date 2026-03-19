package png

import (
	"bytes"
	"fmt"
)

func GetHeader(imgBytes []byte) (IHDR, error) {
	chunks, err := ExtractChunks(imgBytes)
	if err != nil {
		return IHDR{}, err
	}

	for _, c := range chunks {
		if bytes.Equal(c.Name, []byte("IHDR")) {
			return parseIHDR(c.Data), nil
		}
	}

	return IHDR{}, fmt.Errorf("chunk IHDR is not found")
}
