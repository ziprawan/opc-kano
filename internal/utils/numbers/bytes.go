package numbers

import (
	"encoding/binary"
)

func ByteToUint32LSB(bytes []byte) int {
	copied := make([]byte, 4)
	for i := range len(bytes) {
		copied[i] = bytes[i]
	}
	return int(binary.LittleEndian.Uint32(copied))
}

func Int32ToByteLSB(num int) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(num))
	return b
}

func ByteToUint16LSB(bytes []byte) int {
	copied := make([]byte, 2)
	for i := range len(bytes) {
		copied[i] = bytes[i]
	}
	return int(binary.LittleEndian.Uint16(copied))
}

func Int16ToByteLSB(num int) []byte {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, uint16(num))
	return b
}
