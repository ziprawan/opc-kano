package kanoutils

import (
	"fmt"
	"strconv"
	"time"
)

var shortMonthNames = []string{
	"Jan",
	"Feb",
	"Mar",
	"Apr",
	"Mei",
	"Jun",
	"Jul",
	"Agu",
	"Sep",
	"Okt",
	"Nov",
	"Des",
}

func FormatISO8601Date(dateStr string) (string, error) {
	layout := "2006-01-02T15:04:05Z"
	dateFormat := "Monday, 02 January 2006"
	date, err := time.Parse(layout, dateStr)
	if err != nil {
		return "", err
	} else {
		return date.Format(dateFormat), nil
	}
}

func FormatNumber(num int64) string {
	strNum := strconv.FormatInt(num, 10)
	strNumLen := len(strNum)
	formatted := ""

	for i := range strNumLen {
		idx := strNumLen - i - 1
		char := string(strNum[idx])

		formatted = char + formatted
		if i%3 == 2 && i != strNumLen-1 {
			formatted = "," + formatted
		}
	}

	return formatted
}

func FormatMonthIndoShort(m time.Month) string {
	if time.January <= m && m <= time.December {
		return shortMonthNames[m-1]
	}
	return "???"
}

func FormatRangeDateOnly(firstDate, lastDate string) string {
	parsedFirstDate, err := time.Parse(time.DateOnly, firstDate)
	if err != nil {
		return ""
	}

	parsedLastDate, err := time.Parse(time.DateOnly, lastDate)
	if err != nil {
		return ""
	}

	fYear, fMonth, fDay := parsedFirstDate.Date()
	lYear, lMonth, lDay := parsedLastDate.Date()

	rightSide := fmt.Sprintf("%d %s %d", lDay, FormatMonthIndoShort(lMonth), lYear)
	leftSide := ""

	if fYear != lYear {
		leftSide = fmt.Sprintf("%d %s %d", fDay, FormatMonthIndoShort(fMonth), fYear)
	} else if fMonth != lMonth {
		leftSide = fmt.Sprintf("%d %s", fDay, FormatMonthIndoShort(fMonth))
	} else if fDay != lDay {
		leftSide = fmt.Sprintf("%d", fDay)
	}
	if leftSide != "" {
		leftSide += " - "
	}

	return fmt.Sprintf("%s%s", leftSide, rightSide)
}
