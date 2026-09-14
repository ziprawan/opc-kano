package six

import (
	"fmt"
	"kano/internal/database"
	"kano/internal/utils/word"
	"strconv"
	"strings"
)

var db = database.GetInstance().Debug()

func parseClassCtx(classCtx string) (string, uint, error) {
	classCode, classNumStr, ok := strings.Cut(classCtx, "-")
	if !ok {
		return "", 0, fmt.Errorf("tidak ada strip tidak terdeteksi")
	}
	classCode = strings.ToUpper(classCode)
	checkResult := checkClassCode(classCode)
	if checkResult != "" {
		return "", 0, fmt.Errorf("%s", checkResult)
	}
	classNum, err := strconv.ParseUint(classNumStr, 10, 0)
	if err != nil {
		return "", 0, fmt.Errorf("tidak dapat mengubah %q menjadi angka", classNumStr)
	}

	return classCode, uint(classNum), nil
}

func checkClassCode(classCode string) string {
	if len(classCode) != 6 {
		return "panjang kode matkul bukan 6"
	}
	if !word.IsCharUpper(classCode[0]) || !word.IsCharUpper(classCode[1]) {
		return "dua karakter pertama bukan a-z"
	}
	if _, err := strconv.ParseUint(classCode[2:], 10, 0); err != nil {
		return "empat karakter terakhir bukan angka"
	}

	return ""
}
