package numbers

import (
	"encoding/binary"
)

func ByteToUint32LSB(bytes []byte) uint32 {
	copied := make([]byte, 4)
	for i := range len(bytes) {
		copied[i] = bytes[i]
	}
	return binary.LittleEndian.Uint32(copied)
}

func ByteToUint32MSB(bytes []byte) uint32 {
	copied := make([]byte, 4)
	for i := range len(bytes) {
		copied[i] = bytes[i]
	}
	return binary.BigEndian.Uint32(copied)
}

func Uint32ToByteLSB(num uint) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(num))
	return b
}

func Uint32ToByteMSB(num uint) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(num))
	return b
}

func ByteToUint16LSB(bytes []byte) uint16 {
	copied := make([]byte, 2)
	for i := range len(bytes) {
		copied[i] = bytes[i]
	}
	return binary.LittleEndian.Uint16(copied)
}

func ByteToUint16MSB(bytes []byte) uint16 {
	copied := make([]byte, 2)
	for i := range len(bytes) {
		copied[i] = bytes[i]
	}
	return binary.BigEndian.Uint16(copied)
}

func Uint16ToByteLSB(num uint) []byte {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, uint16(num))
	return b
}

func Uint16ToByteMSB(num uint) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(num))
	return b
}
