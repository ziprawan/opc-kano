package schedules

import (
	"fmt"

	"gorm.io/gorm"
)

func ApplyDiff(sems []SemesterDiff) error {
	fmt.Println("There is", len(lecturers), "lecturers in the cache")
	err := db.Transaction(func(tx *gorm.DB) error {
		for _, sem := range sems {
			err := applyPerSemester(tx, sem)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

func applyPerSemester(dbx *gorm.DB, sem SemesterDiff) error {
	if len(sem.AddedSubjects) > 0 {
		err := handleAdded(dbx, sem.ID, sem.AddedSubjects)
		if err != nil {
			return err
		}
	}

	if len(sem.RemovedSubjects) > 0 {
		err := handleRemoved(dbx, sem.ID, sem.RemovedSubjects)
		if err != nil {
			return err
		}
	}

	if len(sem.ModifiedSubjects) > 0 {
		err := handleModified(dbx, sem.ID, sem.ModifiedSubjects)
		if err != nil {
			return err
		}
	}

	return nil
}
