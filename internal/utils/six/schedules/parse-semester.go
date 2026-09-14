package schedules

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func parsePerSemesterContext(semsCtx string, page *goquery.Selection) (SemesterSubject, error) {
	optgroups := page.Find("#prodi optgroup")
	res := SemesterSubject{}

	prodiTotal := 0
	optgroups.Each(func(i int, s *goquery.Selection) {
		prodiTotal += s.Find("option").Length()
	})

	// Get all available prodis
	var err error
	res.Majors = make([]Major, 0, prodiTotal)
	optgroups.Each(func(i int, g *goquery.Selection) {
		// Faculty name
		faculty, ok := g.Attr("label")
		if !ok {
			err = fmt.Errorf("unable to resolve optgroup label at idx %d", i)
			return
		}
		faculty = strings.TrimSpace(faculty)

		g.Find("option").Each(func(j int, o *goquery.Selection) {
			// Major name
			_, name, _ := strings.Cut(o.Text(), " - ")
			if name == "" {
				err = fmt.Errorf("unable to resolve major name at optgroup idx %d and option idx %d", i, j)
				return
			}
			name = strings.TrimSpace(name)

			// Major Id
			value, ok := o.Attr("value")
			if !ok {
				err = fmt.Errorf("unable to resolve option value at optgroup idx %d and option idx %d", i, j)
				return
			}
			majorId, err := strconv.ParseUint(value, 10, 0)
			if err != nil {
				return
			}

			// Append the result
			res.Majors = append(res.Majors, Major{
				ID:      uint(majorId),
				Name:    name,
				Faculty: faculty,
			})
		})
	})
	if err != nil {
		return res, err
	}

	// Fetch per major schedule page and parse them
	for _, major := range res.Majors {
		thePage, err := schedQuery(semsCtx, fmt.Sprintf("%d", major.ID))
		if err != nil {
			return res, fmt.Errorf("major %d: SchedQuery: %s", major.ID, err)
		}
		subs, err := parsePerMajorSchedulePage(thePage)
		if err != nil {
			return res, fmt.Errorf("major %d: parsePerMajorSchedulePage: %s", major.ID, err)
		}

		// Append subject data to the semester data
		for _, sub := range subs {
			appendSubjectToSemsData(&res, sub)
		}
	}

	// deleteTmpFiles()

	return res, nil
}
