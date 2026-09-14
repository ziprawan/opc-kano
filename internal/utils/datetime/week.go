package datetime

import (
	"time"
)

var jakarta *time.Location

func init() {
	var err error
	jakarta, err = time.LoadLocation("Asia/Jakarta")
	if err != nil {
		panic(err)
	}
}

func StartOfWeek(t time.Time) time.Time {
	t = t.In(jakarta)
	newT := time.Date(
		t.Year(), t.Month(), t.Day()-((int(t.Weekday())+6)%7),
		0, 0, 0, 0, jakarta,
	)
	return newT
}
