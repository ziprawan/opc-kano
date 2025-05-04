package kanoutils

import "fmt"

func Pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("given data is empty")
	}
	if len(data)%blockSize != 0 {
		return nil, fmt.Errorf("given data is not a multiple of block size")
	}

	padLen := int(data[len(data)-1])
	if padLen == 0 || padLen > blockSize {
		return nil, fmt.Errorf("invalid padding")
	}

	for _, v := range data[len(data)-padLen:] {
		if int(v) != padLen {
			return nil, fmt.Errorf("invaling padding content")
		}
	}

	return data[:len(data)-padLen], nil
}
