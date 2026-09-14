package six

import (
	"fmt"
	"kano/internal/utils/word"
)

func validOffsetChars(char byte) bool {
	return char == '^' || char == '+' || char == '-' || // Symbols
		(char >= '0' && char <= '9') || // Numbers
		char == 'm' || char == 'h' || char == 'd' // Days
}

func unit(char byte) int {
	switch char {
	case 'm':
		return 1
	case 'h':
		return 60
	case 'd':
		return 1440
	default:
		return 1
	}
}

func parseOffset(timeStr string) (offsetMinutes int, anchorAtEnd bool, err error) {
	if len(timeStr) == 0 {
		return
	}

	for i := range timeStr {
		if !validOffsetChars(timeStr[i]) {
			err = fmt.Errorf("terdapat karakter yang tidak didukung: %q", timeStr[i])
			return
		}
	}

	if timeStr[0] == '^' {
		anchorAtEnd = true
		timeStr = timeStr[1:]
	}

	if len(timeStr) == 0 {
		return
	}

	isNegative := false
	if timeStr[0] == '+' || timeStr[0] == '-' {
		if timeStr[0] == '-' {
			isNegative = true
		}
		timeStr = timeStr[1:]
	}

	curUnit := (byte)(0)
	tmpNum := 0
	hasNum := false
	for i := range len(timeStr) {
		c := timeStr[i]
		if word.IsCharNumber(c) {
			hasNum = true
			tmpNum *= 10
			tmpNum += int(c - '0')
			continue
		}

		if c != 'm' && c != 'h' && c != 'd' {
			err = fmt.Errorf("unit tidak terduga: %q", c)
			return
		}

		if !hasNum {
			err = fmt.Errorf("unit terdeteksi sebelum adanya angka: %q", c)
			return
		}

		// Validasi urutan unit
		if curUnit == 'd' && c == 'd' {
			err = fmt.Errorf("duplikat unit 'd'")
			return
		}
		if curUnit == 'h' && (c == 'h' || c == 'd') {
			if c == 'h' {
				err = fmt.Errorf("duplikat unit 'h'")
			} else {
				err = fmt.Errorf("unit 'd' tidak diperbolehkan setelah unit 'h'")
			}
			return
		}
		if curUnit == 'm' && (c == 'm' || c == 'h' || c == 'd') {
			switch c {
			case 'm':
				err = fmt.Errorf("duplikat unit 'm'")
			case 'h':
				err = fmt.Errorf("unit 'h' tidak diperbolehkan setelah unit 'm'")
			default:
				err = fmt.Errorf("unit 'd' tidak diperbolehkan setelah unit 'm'")
			}
			return
		}
		curUnit = c

		// Masukin ke offset
		offsetMinutes += unit(curUnit) * tmpNum
		tmpNum = 0
		hasNum = false
	}

	if tmpNum != 0 {
		if curUnit != 0 {
			err = fmt.Errorf("unit tidak dispesifikasikan untuk angka terakhir")
			return
		} else {
			offsetMinutes = tmpNum
		}
	}

	if isNegative {
		offsetMinutes = -offsetMinutes
	}

	return
}
