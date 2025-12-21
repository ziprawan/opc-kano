package image

import (
	"encoding/binary"
	"kano/internal/utils/numbers"
)

func readBits(b byte, n, m uint) uint8 {
	return (b >> n) & ((1 << (m - n + 1)) - 1)
}

func readBit(b byte, pos uint) uint8 {
	if pos > 7 {
		pos = 7
	}
	return (b >> pos) & 1
}

func writeBit(b *byte, pos uint, value bool) {
	if pos > 7 {
		pos = 7
	}

	if value {
		*b |= (1 << pos)
	} else {
		*b &= (0xff - (1 << pos))
	}
}

func uint24ToBytes(num uint) []byte {
	var MAX = uint(1 << 24)
	if num >= MAX {
		num = MAX
	}

	numByte := make([]byte, 4)
	binary.LittleEndian.PutUint32(numByte, uint32(num))

	res := make([]byte, 3)
	res[0] = numByte[0]
	res[1] = numByte[1]
	res[2] = numByte[2]

	return res
}

func bytesToUint24(bytes []byte) uint {
	tmp := bytes[:3]
	return uint(numbers.ByteToUint32LSB(tmp))
}
