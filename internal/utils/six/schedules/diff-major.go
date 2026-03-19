package schedules

import (
	"kano/internal/database/models"
	"slices"

	"gorm.io/gorm/clause"
)

func initMajors(scheds []SemesterSubject) error {
	majors := []models.Major{}
	for _, semester := range scheds {
		for _, major := range semester.Majors {
			obj := models.Major{ID: major.ID, Name: major.Name, Faculty: major.Faculty}
			if !slices.Contains(majors, obj) {
				majors = append(majors, obj)
			}
		}
	}

	tx := db.
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoNothing: true,
		}).
		Create(&majors)

	return tx.Error
}
