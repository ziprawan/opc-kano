package schedules

import (
	"fmt"
	"kano/internal/database/models"
	"slices"
)

// See vars.go for related variables

func initLecturers(scheds []SemesterSubject) error {
	strLecturers := []string{}
	for _, sched := range scheds {
		for _, subject := range sched.Subjects {
			for _, class := range subject.Classes {
				for _, lecturer := range class.Lecturers {
					if !slices.Contains(strLecturers, lecturer) {
						strLecturers = append(strLecturers, lecturer)
					}
				}
			}
		}
	}

	var found []models.Lecturer
	tx := db.Find(&found)
	if tx.Error != nil {
		return tx.Error
	}

	// if len(found) == len(strLecturers) {
	// 	lecturers = found
	// 	return nil
	// }

	toInsert := make([]models.Lecturer, 0, len(strLecturers))
	i := 0
	for _, name := range strLecturers {
		if !slices.ContainsFunc(found, func(a models.Lecturer) bool { return a.Name == name }) {
			toInsert = append(toInsert, models.Lecturer{Name: name})
			i++
		}
	}

	lecturers = make([]models.Lecturer, 0, len(strLecturers)+len(found))
	for _, f := range found {
		lecturers = append(lecturers, f)
	}

	if len(toInsert) > 0 {
		fmt.Printf("Missing %d lecturer(s) from database\n", len(toInsert))
		tx = db.Create(&toInsert)
		if tx.Error != nil {
			return tx.Error
		}
		fmt.Printf("Inserted %d lecture(s) into database\n", tx.RowsAffected)

		for _, in := range toInsert {
			lecturers = append(lecturers, in)
		}
	}

	return nil
}

func findLecturerId(lecturerName string) uint {
	for _, lecturer := range lecturers {
		if lecturer.Name == lecturerName {
			return lecturer.ID
		}
	}

	return 0
}
