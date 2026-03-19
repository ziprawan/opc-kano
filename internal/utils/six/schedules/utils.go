package schedules

import (
	"errors"
	"fmt"
	"kano/internal/utils/six/fetcher"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func appendSubjectToSemsData(semsData *SemesterSubject, subject Subject) {
	if semsData == nil {
		return
	}

	idx := -1
	for i, sub := range semsData.Subjects {
		if sub.Code == subject.Code {
			idx = i
			break
		}
	}

	if idx == -1 {
		semsData.Subjects = append(semsData.Subjects, subject)
	} else {
		semsData.Subjects[idx].Classes = append(semsData.Subjects[idx].Classes, subject.Classes...)
	}
}

func parseSemsData(li *goquery.Selection) (uint, uint, error) {
	semsText := li.Text()
	semsNumTxt, semsYearTxt, ok := strings.Cut(semsText, "-")
	if !ok {
		return 0, 0, fmt.Errorf("unexpected semester data: %q", semsText)
	}

	// Trimming the spaces
	semsNumTxt = strings.TrimSpace(semsNumTxt)
	semsYearTxt = strings.TrimSpace(semsYearTxt)

	// Getting semester number
	semsNumTxt = strings.ReplaceAll(semsNumTxt, "Semester ", "")
	semsNum, err := strconv.ParseUint(semsNumTxt, 10, 0)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot parse semester number as uint: %s", semsNumTxt)
	}

	left, _, ok := strings.Cut(semsYearTxt, "/")
	if !ok {
		return 0, 0, fmt.Errorf("unexpected semester year data: %q", semsYearTxt)
	}
	semsYear, err := strconv.ParseUint(left, 10, 0)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot parse semester start year as uint: %s", semsYearTxt)
	}

	return uint(semsYear), uint(semsNum), nil
}

func schedQuery(semsCtx, prodiId string) (*goquery.Selection, error) {
	fPath := path.Join("tmp", "schedules")
	_, err := os.Stat(fPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}

		err = os.MkdirAll(fPath, 0777)
		if err != nil {
			return nil, err
		}
	}

	fPath = path.Join(fPath, fmt.Sprintf("%s+%s", semsCtx, prodiId))
	file, err := os.OpenFile(fPath, os.O_RDONLY, 0644)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	} else {
		d, e := goquery.NewDocumentFromReader(file)
		if e != nil {
			return nil, e
		}
		return d.Selection, e
	}

	thePath, err := fetcher.BuildAppPathWithSems([]string{"kelas", "jadwal", "kuliah"}, semsCtx)
	if err != nil {
		return nil, err
	}
	query := map[string][]string{
		"fakultas": {""},
		"prodi":    {prodiId},
	}
	_, p, e := fetcher.GetPage(thePath, query)

	htm, err := p.Html()
	if err == nil && !errors.Is(e, fetcher.ErrInvalidCredential) {
		os.WriteFile(fPath, []byte(htm), 0644)
	}

	return p, e
}

func schedRowSorter(a, b ScheduleRow) int {
	if a.Subject.Code > b.Subject.Code {
		return 1
	} else if a.Subject.Code < b.Subject.Code {
		return -1
	} else {
		return 0
	}
}

func text(s *goquery.Selection) string {
	return strings.TrimSpace(s.Text())
}

func isText(s *goquery.Selection) bool {
	return goquery.NodeName(s) == "#text"
}

func dataToMajorConstraint(data string) (MajorConstraint, error) {
	data = strings.ReplaceAll(data, "\u00A0", " ")
	data = strings.TrimSpace(data)
	idStr, addon, _ := strings.Cut(data, " ")
	id, _ := strconv.ParseUint(idStr, 10, 0)

	// By default, if the strings.Cut failed, the addon would be ""
	// and I assume idStr contains the major ID we want.
	// Also the id is 0 if the strconv.ParseUint is failed.
	//
	// Maybe I should make a validation where addon == "" and id == 0
	// cannot be in the same time.
	if addon == "" && id == 0 {
		return MajorConstraint{ID: 0, Addon: idStr}, nil // Special case for "CCE"
	}

	return MajorConstraint{ID: uint(id), Addon: addon}, nil
}

func splitTrim(str string, sep string) []string {
	strs := strings.Split(str, sep)
	for i := range strs {
		strs[i] = strings.TrimSpace(strs[i])
	}

	return strs
}

func splitTrimN(str string, sep string, n int) []string {
	strs := strings.SplitN(str, sep, n)
	for i := range strs {
		strs[i] = strings.TrimSpace(strs[i])
	}

	return strs
}

func parseDateTime(dateStr, timeStr string) (time.Time, error) {
	dateStr = strings.ToLower(dateStr)
	dateStr = strings.Replace(dateStr, "mei", "may", 1)
	dateStr = strings.Replace(dateStr, "agu", "aug", 1)
	dateStr = strings.Replace(dateStr, "okt", "oct", 1)
	dateStr = strings.Replace(dateStr, "des", "dec", 1)

	timeStr = strings.ReplaceAll(timeStr, ".", ":")

	layout := "2 Jan 2006 15:04 -0700"
	return time.Parse(layout, fmt.Sprintf("%s %s +0700", dateStr, timeStr))
}

func safeHtml(s *goquery.Selection) string {
	html, err := s.Html()
	if err != nil {
		return ""
	} else {
		return html
	}
}

func CleanupTmpFiles() {
	dirs, err := os.ReadDir(path.Join("tmp", "schedules"))
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, dir := range dirs {
		if dir.Type().IsRegular() {
			// fmt.Println(dir.Name())
			err = os.Remove(path.Join("tmp", "schedules", dir.Name()))
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
