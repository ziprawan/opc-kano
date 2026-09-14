package schedules

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var subjects []Subject
var unresolvedSubjectIds []string

func getSubjectId(code string) uint {
	if len(subjects) == 0 {
		UpdateSubjects()
	}

	for _, sub := range subjects {
		if sub.Code == code {
			return sub.ID
		}
	}

	return 0
}

func UpdateSubjects() {
	subjects = nil
	bt, err := os.ReadFile("dumps/six/subject-id_map.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = json.Unmarshal(bt, &subjects)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func parseSubjectRow(row *goquery.Selection) (ScheduleRow, error) {
	res := ScheduleRow{}

	columns := row.Children()
	if columns.Length() != 10 {
		return res, fmt.Errorf("unexpected columns length %d, expected 10", columns.Length())
	}

	textEq := func(idx int) string { return strings.TrimSpace(columns.Eq(idx).Text()) }
	uintEq := func(idx int) (uint, error) { a, b := strconv.ParseUint(textEq(idx), 10, 0); return uint(a), b }

	res.Subject.Code = textEq(2)
	res.Subject.Name = textEq(3)

	// Resolve subject id
	sId := getSubjectId(res.Subject.Code)
	if sId == 0 {
		unresolvedSubjectIds = append(unresolvedSubjectIds, res.Subject.Code)
	}
	res.Subject.ID = sId

	sks, err := uintEq(4)
	if err != nil {
		return res, fmt.Errorf("failed to parse %q as uint for sks", textEq(4))
	}
	res.Subject.SKS = sks

	classNum, err := uintEq(5)
	if err != nil {
		return res, fmt.Errorf("failed to parse %q as uint for classNum", textEq(5))
	}
	res.Class.Number = classNum

	classQuotaStr := textEq(6)
	if classQuotaStr == "-" {
		res.Class.Quota = sql.NullInt32{Valid: false}
	} else {
		classQuota, err := strconv.ParseUint(classQuotaStr, 10, 0)
		if err != nil {
			return res, fmt.Errorf("failed to parse %q as uint for classQuota", classQuotaStr)
		}
		iClassQuota := int32(classQuota)
		res.Class.Quota = sql.NullInt32{Int32: iClassQuota, Valid: true}
	}

	lecturers := columns.Eq(7).Find("li")
	res.Class.Lecturers = make([]string, lecturers.Length())
	for i, lecturer := range lecturers.EachIter() {
		res.Class.Lecturers[i] = strings.TrimSpace(lecturer.Text())
	}

	constraints, err := parseConstraints(columns.Eq(8))
	if err != nil {
		return res, fmt.Errorf("parseConstraints: %s", err)
	}
	res.Class.Constraints = constraints

	lastRow, err := parseSchedules(columns.Eq(9))
	if err != nil {
		return res, fmt.Errorf("parseSchedules: %s", err)
	}
	// fmt.Println(lastRow)

	res.Class.Links = lastRow.ClassLink
	res.Class.Schedules = lastRow.Schedules
	res.Class.ID = lastRow.ClassId

	return res, nil
}
