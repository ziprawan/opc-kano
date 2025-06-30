package kanoutils

import "time"

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
