package word

import (
	"slices"
	"strings"
)

const (
	QUOTATION_MARK      = rune(0x22)   // (")
	APOSTROPHE          = rune(0x27)   // (')
	GRAVE_ACCENT        = rune(0x60)   // (`)
	LEFT_SINGLE_MARK    = rune(0x2018) // (‘)
	RIGHT_SINGLE_MARK   = rune(0x2019) // (’)
	LEFT_DOUBLE_MARK    = rune(0x201c) // (“)
	RIGHT_DOUBLE_MARK   = rune(0x201d) // (”)
	MODIFIER_APOSTROPHE = rune(0x2bc)  // (ʼ)
)

var closers = map[rune][]rune{
	0x22:   {0x22, 0x201c, 0x201d},
	0x27:   {0x27, 0x2018, 0x2019},
	0x60:   {0x60, 0x2bc},
	0x201c: {0x201c, 0x201d},
	0x201d: {0x201c, 0x201d},
	0x2018: {0x2018, 0x2019},
	0x2019: {0x2018, 0x2019},
	0x2bc:  {0x2bc, 0x60},
}

func NormalizeQuote(s string) string {
	s = strings.ReplaceAll(s, string(LEFT_SINGLE_MARK), string(APOSTROPHE))
	s = strings.ReplaceAll(s, string(RIGHT_SINGLE_MARK), string(APOSTROPHE))
	s = strings.ReplaceAll(s, string(LEFT_DOUBLE_MARK), string(QUOTATION_MARK))
	s = strings.ReplaceAll(s, string(RIGHT_DOUBLE_MARK), string(QUOTATION_MARK))
	s = strings.ReplaceAll(s, string(MODIFIER_APOSTROPHE), string(GRAVE_ACCENT))

	return s
}

func IsQuote(char rune) bool {
	// iPhone's smart quotes pmo
	return char == QUOTATION_MARK ||
		char == APOSTROPHE ||
		char == GRAVE_ACCENT ||
		char == LEFT_SINGLE_MARK ||
		char == RIGHT_SINGLE_MARK ||
		char == LEFT_DOUBLE_MARK ||
		char == RIGHT_DOUBLE_MARK ||
		char == MODIFIER_APOSTROPHE
}

func IsTwinQuote(char, other rune) bool {
	closer, ok := closers[char]
	if !ok {
		return false
	}
	return slices.Contains(closer, other)
}

func IsCharUpper(char byte) bool {
	return char >= 'A' && char <= 'Z'
}

func IsCharLower(char byte) bool {
	return char >= 'a' && char <= 'z'
}

func IsCharNumber(char byte) bool {
	return char >= '0' && char <= '9'
}
