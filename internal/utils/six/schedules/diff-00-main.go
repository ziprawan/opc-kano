package schedules

import (
	"fmt"
)

func GetScheduleDiff(scheds []SemesterSubject) ([]SemesterDiff, error) {
	var err error

	// Initialize needed table data
	err = initLecturers(scheds)
	if err != nil {
		return nil, err
	}
	err = initRooms(scheds)
	if err != nil {
		return nil, err
	}
	err = initSubjects(scheds)
	if err != nil {
		return nil, err
	}
	err = initMajors(scheds)
	if err != nil {
		return nil, err
	}
	err = initMajorConstraints(scheds)
	if err != nil {
		return nil, err
	}

	// The real diff generator
	semsDiffs := make([]SemesterDiff, 0, len(scheds))
	for _, semester := range scheds {
		semsDiff, err := generateSemesterDiff(semester)
		if err != nil {
			return nil, fmt.Errorf("generateSemesterDiff %d-%d: %s", semester.Year, semester.Semester, err)
		}
		semsDiffs = append(semsDiffs, semsDiff)
	}

	return semsDiffs, nil
}
