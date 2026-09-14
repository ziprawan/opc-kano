package schedules

import (
	"database/sql"
	"fmt"
	"kano/internal/database/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func handleModifiedClasses(dbx *gorm.DB, modClasses []ClassDiff) error {
	constraintsToInsert := make([]models.ClassConstraint, 0, len(modClasses))
	constraintsToRemove := make([]uint, 0, len(modClasses))
	addedLecturersTotal := 0
	removedLecturersTotal := 0
	addedSchedsTotal := 0
	removedSchedsTotal := 0
	modifiedSchedsTotal := 0
	for _, modClass := range modClasses {
		// Check all the modified class fields
		hasChange := false
		theChange := models.SubjectClass{ID: modClass.ID}
		q := dbx

		if modClass.Number.HasDiff {
			q = q.Select("number")
			hasChange = true
			theChange.Number = modClass.Number.After
		}
		if modClass.Quota.HasDiff {
			q = q.Select("quota")
			hasChange = true
			theChange.Quota = modClass.Quota.After
		}
		if modClass.Links.HasDiff {
			q = q.Select("edunex_class_id", "teams_link")
			hasChange = true
			theChange.EdunexClassID = modClass.Links.After.EdunexClassId
			theChange.TeamsLink = sql.NullString{String: modClass.Links.After.Teams, Valid: len(modClass.Links.After.Teams) > 0}
		}
		if hasChange {
			tx := q.Updates(&theChange)
			if tx.Error != nil {
				return fmt.Errorf("failed to update class id %d: %s", modClass.ID, tx.Error)
			}
		}

		// The class constraint
		if modClass.Constraints.HasDiff {
			cnstrnt := modClass.Constraints.After
			// Uhm, yeah, special case if the constraint major was different
			// I will just remove and reinsert it again
			// Yep, I need to reorder the db calls, I must call delete first, then insert (Done tho)
			// if !compareArr(modClass.Constraints.After.Constraint.Majors, modClass.Constraints.Before.Constraint.Majors) {
			// 	constraintsToRemove = append(constraintsToRemove, modClass.ID)
			// 	if cnstrnt.Valid {
			// 		modelConstraint, err := constraintModel(Class{ID: modClass.ID, Constraints: cnstrnt})
			// 		if err != nil {
			// 			return fmt.Errorf("constraintModel %d: %s", modClass.ID, err)
			// 		}
			// 		constraintsToInsert = append(constraintsToInsert, modelConstraint)
			// 	}
			// } else {
			// 	if cnstrnt.Valid {
			// 		modelConstraint, err := constraintModel(Class{ID: modClass.ID, Constraints: cnstrnt})
			// 		if err != nil {
			// 			return fmt.Errorf("constraintModel %d: %s", modClass.ID, err)
			// 		}
			// 		constraintsToInsert = append(constraintsToInsert, modelConstraint)
			// 	} else {
			// 		constraintsToRemove = append(constraintsToRemove, modClass.ID)
			// 	}
			// }

			// Code below is refactored from the above by chatgpt
			majorsChanged := !compareArr(
				modClass.Constraints.After.Constraint.Majors,
				modClass.Constraints.Before.Constraint.Majors,
			)

			if majorsChanged || !cnstrnt.Valid {
				constraintsToRemove = append(constraintsToRemove, modClass.ID)
			}

			if cnstrnt.Valid {
				modelConstraint, err := constraintModel(
					Class{ID: modClass.ID, Constraints: cnstrnt},
				)
				if err != nil {
					return fmt.Errorf("constraintModel %d: %w", modClass.ID, err)
				}
				constraintsToInsert = append(constraintsToInsert, modelConstraint)
			}

		}

		addedLecturersTotal += len(modClass.AddedLecturers)
		removedLecturersTotal += len(modClass.RemovedLecturers)
		addedSchedsTotal += len(modClass.AddedSchedules)
		removedSchedsTotal += len(modClass.RemovedSchedules)
		modifiedSchedsTotal += len(modClass.ModifiedSchedules)
	}
	lecturersToInsert := make([]models.LecturerInClass, 0, addedLecturersTotal) // Lecturer In Class models
	lecturersToRemove := make([][2]uint, 0, removedLecturersTotal)              // Array of Subject Class ID and Lecturer ID
	schedsToInsert := make([]models.ClassSchedule, 0, addedSchedsTotal)         // Schedule models
	schedsToRemove := make([]uint, 0, removedSchedsTotal)                       // Schedule IDs
	for _, modClass := range modClasses {
		for _, al := range modClass.AddedLecturers {
			id := findLecturerId(al)
			if id == 0 {
				return fmt.Errorf("unable to find lecturer id for name=%q", al)
			}
			lecturersToInsert = append(lecturersToInsert, models.LecturerInClass{SubjectClassID: modClass.ID, LecturerID: id})
		}
		for _, rl := range modClass.RemovedLecturers {
			id := findLecturerId(rl)
			if id == 0 {
				return fmt.Errorf("unable to find lecturer id for name=%q", rl)
			}
			lecturersToRemove = append(lecturersToRemove, [2]uint{modClass.ID, id})
		}

		sched, err := schedulesModel(Class{ID: modClass.ID, Schedules: modClass.AddedSchedules})
		if err != nil {
			return fmt.Errorf("schedulesModel %d: %s", modClass.ID, err)
		}
		schedsToInsert = append(schedsToInsert, sched...)
		for _, rs := range modClass.RemovedSchedules {
			schedsToRemove = append(schedsToRemove, rs.ID)
		}
	}

	// Remove constraints
	if len(constraintsToRemove) > 0 {
		tx := dbx.Where("subject_class_id IN ?", constraintsToRemove).Delete(&models.ClassConstraint{})
		if tx.Error != nil {
			return fmt.Errorf("failed to remove constraints: %s", tx.Error)
		}
	}

	// Remove lecturer in class
	if len(lecturersToRemove) > 0 {
		tx := dbx.Where("(subject_class_id, lecturer_id) IN ?", lecturersToRemove).Delete(&models.LecturerInClass{})
		if tx.Error != nil {
			return fmt.Errorf("failed to remove lecturers: %s", tx.Error)
		}
	}

	// Remove class schedules
	if len(schedsToRemove) > 0 {
		tx := dbx.Where("id IN ?", schedsToRemove).Delete(&models.ClassSchedule{})
		if tx.Error != nil {
			return fmt.Errorf("failed to remove schedules: %s", tx.Error)
		}
	}

	// Insert constraints
	if len(constraintsToInsert) > 0 {
		tx := dbx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "subject_class_id"}}, UpdateAll: true}).CreateInBatches(&constraintsToInsert, 500)
		if tx.Error != nil {
			return fmt.Errorf("failed to insert constraints: %s", tx.Error)
		}
	}

	// Insert lecturer in class
	if len(lecturersToInsert) > 0 {
		tx := dbx.Clauses(clause.OnConflict{UpdateAll: true}).CreateInBatches(&lecturersToInsert, 500)
		if tx.Error != nil {
			return fmt.Errorf("failed to insert lecturers: %s", tx.Error)
		}
	}

	// Insert class schedules
	if len(schedsToInsert) > 0 {
		tx := dbx.Clauses(clause.OnConflict{UpdateAll: true}).CreateInBatches(&schedsToInsert, 500)
		if tx.Error != nil {
			return fmt.Errorf("failed to insert schedules: %s", tx.Error)
		}
	}

	// Let's handle the modified schedule one
	if modifiedSchedsTotal > 0 {
		modScheds := make([]ScheduleDiff, 0, modifiedSchedsTotal)
		for _, c := range modClasses {
			for _, s := range c.ModifiedSchedules {
				modScheds = append(modScheds, s)
			}
		}

		err := handleModifiedSchedules(dbx, modScheds)
		if err != nil {
			return fmt.Errorf("handleModifiedSchedules: %s", err)
		}
	}

	return nil
}
