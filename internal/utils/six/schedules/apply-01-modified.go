package schedules

import (
	"fmt"
	"kano/internal/database/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func handleModified(dbx *gorm.DB, semsId uint, modified []SubjectDiff) error {
	// Counting, so I can easily allocate the arrays
	addedClassesTotal := 0
	addedConstraintsTotal := 0
	addedSchedulesTotal := 0
	removedClassesTotal := 0
	modifiedClassesTotal := 0
	for _, mod := range modified {
		removedClassesTotal += len(mod.RemovedClasses)
		addedClassesTotal += len(mod.AddedClasses)
		modifiedClassesTotal += len(mod.ModifiedClasses)

		for _, a := range mod.AddedClasses {
			if a.Constraints.Valid {
				addedConstraintsTotal++
			}
			addedSchedulesTotal += len(a.Schedules)
		}
	}

	// Allocation and filling
	removedClassIds := make([]uint, 0, removedClassesTotal)
	classesToInsert := make([]models.SubjectClass, 0, addedClassesTotal)
	constraintsToInsert := make([]models.ClassConstraint, 0, addedConstraintsTotal)
	schedulesToInsert := make([]models.ClassSchedule, 0, addedSchedulesTotal)
	modifiedClasses := make([]ClassDiff, 0, modifiedClassesTotal)
	for _, mod := range modified {
		// Uhh
		modifiedClasses = append(modifiedClasses, mod.ModifiedClasses...)

		// Removed classes
		for _, rc := range mod.RemovedClasses {
			removedClassIds = append(removedClassIds, rc.ID)
		}

		// Added classes
		for _, c := range mod.AddedClasses {
			// The literal classes
			sc, err := classModel(semsId, mod.ID, c)
			if err != nil {
				return fmt.Errorf("classModel %d: %s", c.ID, err)
			}
			classesToInsert = append(classesToInsert, sc)

			// Available class constraints
			if c.Constraints.Valid {
				cnstrnt, err := constraintModel(c)
				if err != nil {
					return fmt.Errorf("constraintModel %d: %s", c.ID, err)
				}
				constraintsToInsert = append(constraintsToInsert, cnstrnt)
			}

			// Class schedules
			scheds, err := schedulesModel(c)
			if err != nil {
				return fmt.Errorf("schedulesModel %d: %s", c.ID, err)
			}
			schedulesToInsert = append(schedulesToInsert, scheds...)
		}
	}

	// Remove classes
	// This will also delete the class constraint and class schedules :P
	if len(removedClassIds) > 0 {
		tx := dbx.
			Unscoped().
			Where("id IN ?", removedClassIds).
			Delete(&models.SubjectClass{})
		if tx.Error != nil {
			return tx.Error
		}
	}

	// Added classes
	if len(classesToInsert) > 0 {
		tx := dbx.Clauses(clause.OnConflict{UpdateAll: true}).CreateInBatches(&classesToInsert, 500)
		if tx.Error != nil {
			return fmt.Errorf("failed to insert classes: %s", tx.Error)
		}
		if tx.RowsAffected != int64(addedClassesTotal) {
			return fmt.Errorf("insert class: invalid total of affected rows is %d when the expected is %d rows", tx.RowsAffected, addedClassesTotal)
		}
	}

	// Added classes' constraints
	if len(constraintsToInsert) > 0 {
		tx := dbx.Clauses(clause.OnConflict{UpdateAll: true}).CreateInBatches(&constraintsToInsert, 500)
		if tx.Error != nil {
			return fmt.Errorf("failed to insert constraints: %s", tx.Error)
		}
		if tx.RowsAffected != int64(addedConstraintsTotal) {
			return fmt.Errorf("insert constraint: invalid total of affected rows is %d when the expected is %d rows", tx.RowsAffected, addedClassesTotal)
		}
	}
	// Added classes' schedules
	if len(schedulesToInsert) > 0 {
		tx := dbx.Clauses(clause.OnConflict{UpdateAll: true}).CreateInBatches(&schedulesToInsert, 500)
		if tx.Error != nil {
			return fmt.Errorf("failed to insert schedules: %s", tx.Error)
		}
		if tx.RowsAffected != int64(addedSchedulesTotal) {
			return fmt.Errorf("insert schedules: invalid total of affected rows is %d when the expected is %d rows", tx.RowsAffected, addedClassesTotal)
		}
	}

	// Now just handle the modified classes =w=
	if len(modifiedClasses) > 0 {
		err := handleModifiedClasses(dbx, modifiedClasses)
		if err != nil {
			return fmt.Errorf("handleModifiedClasses: %s", err)
		}
	}

	return nil
}
