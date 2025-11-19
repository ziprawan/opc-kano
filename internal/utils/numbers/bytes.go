package numbers

import (
	"encoding/binary"
)

func ByteToInt32LSB(bytes []byte) int {
	return int(binary.LittleEndian.Uint32(bytes))
}

func Int32ToByteLSB(num int) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(num))
	return b
}

func ByteToInt16LSB(bytes []byte) int {
	return int(binary.LittleEndian.Uint16(bytes))
}

func Int16ToByteLSB(num int) []byte {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, uint16(num))
	return b
}
