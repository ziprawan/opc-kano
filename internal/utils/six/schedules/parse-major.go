package schedules

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

func parsePerMajorSchedulePage(page *goquery.Selection) ([]Subject, error) {
	prodVal, ok := page.Find("#prodi option[selected]").Attr("value")
	if !ok {
		return nil, fmt.Errorf("failed to get prodi context")
	}
	majorId, err := strconv.ParseUint(prodVal, 10, 0)
	if err != nil {
		return nil, fmt.Errorf("ParseUint: failed to parse major id as uint: %q", prodVal)
	}

	rows := page.Find("tbody tr")
	if rows.Length() == 0 {
		return []Subject{}, nil
	}

	parsedRows := make([]ScheduleRow, rows.Length())
	for i, row := range rows.EachIter() {
		parsedRow, err := parseSubjectRow(row)
		if err != nil {
			return nil, fmt.Errorf("row %d: parseSubjectRow: %s", i, err)
		}
		parsedRow.Class.AvailableAtMajorId = uint(majorId)
		parsedRows[i] = parsedRow
	}

	// Sort by code so I can so sure know the subjects count
	slices.SortFunc(parsedRows, schedRowSorter)

	subjTotal := 0
	curCode := ""
	for _, pr := range parsedRows {
		if curCode != pr.Subject.Code {
			subjTotal++
			curCode = pr.Subject.Code
		}
	}

	classTots := make([]int, subjTotal)
	i := -1
	curCode = ""
	for _, pr := range parsedRows {
		if curCode != pr.Subject.Code {
			curCode = pr.Subject.Code
			i++
			classTots[i] = 1
		} else {
			classTots[i]++
		}
	}

	subjects := make([]Subject, subjTotal)
	i = -1
	j := 0
	curCode = ""
	for _, pr := range parsedRows {
		if curCode != pr.Subject.Code {
			curCode = pr.Subject.Code

			i++
			j = 0

			subjects[i] = pr.Subject
			subjects[i].Classes = make([]Class, classTots[i])
			subjects[i].Classes[j] = pr.Class
		} else {
			j++
			subjects[i].Classes[j] = pr.Class
		}
	}

	return subjects, nil
}
