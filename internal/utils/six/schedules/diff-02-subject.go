package schedules

import (
	"database/sql"
	"fmt"
	"kano/internal/database/models"
	"slices"
)

// Somehow this only generates the added, removed, and modified classes.
// Yeah, as described by SubjectDiff's fields.
func generateSubjectDiff(subject Subject) (SubjectDiff, error) {
	res := SubjectDiff{ID: subject.ID}
	if dbSems.ID == 0 {
		return res, fmt.Errorf("generateSubjectDiff called before generateSemesterDiff")
	}

	var dbSubjClasses []models.SubjectClass
	tx := db.
		Preload("Subject").
		Preload("Lecturers").
		Preload("Constraint.Majors").
		Preload("Schedules.Rooms").
		Where(models.SubjectClass{SubjectID: subject.ID, SemesterID: dbSems.ID}).
		Find(&dbSubjClasses)
	if tx.Error != nil {
		return res, tx.Error
	}

	// m, _ := json.MarshalIndent(dbSubjClasses, "", "\t")
	// fmt.Println(string(m))

	// Find removed classes
	removed := make([]Class, 0, len(dbSubjClasses))
	for _, dbSubjClass := range dbSubjClasses {
		if !slices.ContainsFunc(subject.Classes, func(a Class) bool { return a.ID == dbSubjClass.ID }) {
			removed = append(removed, Class{
				// I think I just need the class ID and Number
				ID:                 dbSubjClass.ID,
				Number:             dbSubjClass.Number,
				Quota:              dbSubjClass.Quota,
				AvailableAtMajorId: dbSubjClass.AvailableAtMajorID,
			})
		}
	}

	// Find added and existing
	exists := make([]Class, 0, len(subject.Classes))
	added := make([]Class, 0, len(subject.Classes))
	for _, class := range subject.Classes {
		if !slices.ContainsFunc(dbSubjClasses, func(a models.SubjectClass) bool { return a.ID == class.ID }) {
			added = append(added, class)
		} else {
			exists = append(exists, class)
		}
	}

	res.ModifiedClasses = make([]ClassDiff, 0, len(exists))
	for _, classMod := range exists {
		var classDiff ClassDiff

		classDiff.ID = classMod.ID
		idx := slices.IndexFunc(dbSubjClasses, func(a models.SubjectClass) bool { return a.ID == classMod.ID })
		if idx == -1 {
			return res, fmt.Errorf("classDiff %d: invalid index -1", classMod.Number)
		}
		dbClass := dbSubjClasses[idx]

		err := scheduleDiff(&classDiff, classMod, dbClass)
		if err != nil {
			return res, fmt.Errorf("classDiff %d: schedule: %s", classMod.Number, err)
		}
		linksDiff(&classDiff, classMod, dbClass)
		constraintDiff(&classDiff, classMod, dbClass)
		lecturerDiff(&classDiff, classMod, dbClass)

		// Just write it here for number and quota, since it is quite short
		if classMod.Number != dbClass.Number {
			classDiff.Number = varDiff[uint]{
				Before:  dbClass.Number,
				After:   classMod.Number,
				HasDiff: true,
			}
		} else { // Umm, special case, because I want to know the class number
			classDiff.Number = varDiff[uint]{
				Before:  classMod.Number,
				After:   classMod.Number,
				HasDiff: false,
			}
		}
		if classMod.Quota != dbClass.Quota {
			classDiff.Quota = varDiff[sql.NullInt32]{
				Before:  dbClass.Quota,
				After:   classMod.Quota,
				HasDiff: true,
			}
		}

		anyDiff := classDiff.Number.HasDiff ||
			classDiff.Quota.HasDiff ||
			classDiff.Links.HasDiff ||
			classDiff.Constraints.HasDiff ||
			len(classDiff.AddedLecturers) > 0 ||
			len(classDiff.RemovedLecturers) > 0 ||
			len(classDiff.AddedSchedules) > 0 ||
			len(classDiff.RemovedSchedules) > 0 ||
			len(classDiff.ModifiedSchedules) > 0

		if anyDiff {
			res.ModifiedClasses = append(res.ModifiedClasses, classDiff)
		}
	}

	res.AddedClasses = added
	res.RemovedClasses = removed

	return res, nil
}
