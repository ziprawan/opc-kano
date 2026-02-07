package schedules

import (
	"fmt"
	"kano/internal/utils/six/fetcher"
	"strings"
)

var activeOnly bool = true

func GetSchedules() ([]SemesterSubject, error) {
	path, err := fetcher.BuildAppPath([]string{"kelas"})
	if err != nil {
		return nil, err
	}

	_, mainClassPage, err := fetcher.GetPage(path, nil)
	if err != nil {
		return nil, err
	}

	semsList := mainClassPage.Find(`#navbar >:first-child > :nth-child(3) > ul > li:not(.divider)`)
	if semsList.Length() == 0 {
		return nil, fmt.Errorf("failed to find the semester list, maybe the layout was changed?")
	}

	total := 0
	for _, sems := range semsList.EachIter() {
		isActive := sems.HasClass("active")
		if activeOnly && !isActive {
			continue
		}
		total++
	}
	if total == 0 {
		return []SemesterSubject{}, nil
	}

	semesters := make([]SemesterSubject, 0, total)
	for _, sems := range semsList.EachIter() {
		isActive := sems.HasClass("active")
		if activeOnly && !isActive {
			continue
		}

		// Get the semester context (e.g. 2024-1)
		semsYear, semsNum, err := parseSemsData(sems)
		if err != nil {
			return nil, fmt.Errorf("parseSemsData: %s", err)
		}
		semsCtx := fmt.Sprintf("%d-%d", semsYear, semsNum)
		semsData := SemesterSubject{
			BasicSemester: BasicSemester{Year: semsYear, Semester: semsNum},
		}

		// Fetch per semester page and parse it
		thePage, err := schedQuery(semsCtx, "")
		if err != nil {
			return nil, fmt.Errorf("SchedQuery: %s", err)
		}
		semsPageData, err := parsePerSemesterContext(semsCtx, thePage)
		if err != nil {
			return nil, fmt.Errorf("parsePerSemesterContext: %s", err)
		}
		semsData.Majors = semsPageData.Majors
		semsData.Subjects = semsPageData.Subjects

		if len(unresolvedSubjectIds) > 0 {
			return nil, fmt.Errorf("unable to resolve these subject ids: %s", strings.Join(unresolvedSubjectIds, ", "))
		}

		semesters = append(semesters, semsData)
	}

	return semesters, nil
}
