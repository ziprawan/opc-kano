package schedules

import (
	"database/sql"
	"fmt"
	"kano/internal/database/models"
	"slices"
)

// See vars.go for related variables

func initMajorConstraints(scheds []SemesterSubject) error {
	rawMajorConstraints := []MajorConstraint{}
	for _, sched := range scheds {
		for _, subject := range sched.Subjects {
			for _, class := range subject.Classes {
				c := class.Constraints
				if !c.Valid {
					continue
				}
				constraint := c.Constraint

				if len(constraint.Majors) == 0 {
					continue
				}

				for _, cm := range constraint.Majors {
					if !slices.Contains(rawMajorConstraints, cm) {
						rawMajorConstraints = append(rawMajorConstraints, cm)
					}
				}
			}
		}
	}

	found := make([]models.ConstraintMajor, 0, len(rawMajorConstraints)) // RETURN
	ors := make([]models.ConstraintMajor, len(rawMajorConstraints))      // WHERE
	for i, cm := range rawMajorConstraints {
		ors[i] = models.ConstraintMajor{MajorID: sql.NullInt32{Int32: int32(cm.ID), Valid: true}, AddonData: cm.Addon}
		if cm.ID == 0 {
			ors[i].MajorID.Valid = false
		}
	}
	q := db
	for i := range ors {
		q = q.Or(ors[i])
	}
	tx := q.Find(&found)
	if tx.Error != nil {
		return tx.Error
	}

	if len(found) == len(rawMajorConstraints) {
		majorConstraints = found
		return nil
	}

	missing := len(rawMajorConstraints) - len(found)
	fmt.Printf("Missing %d room(s) from database\n", missing)
	toInsert := make([]models.ConstraintMajor, missing)
	i := 0
	for _, cm := range ors {
		if !slices.ContainsFunc(found, func(a models.ConstraintMajor) bool {
			return a.MajorID == cm.MajorID && a.AddonData == cm.AddonData
		}) {
			toInsert[i] = cm
			i++
		}
	}

	tx = db.Create(&toInsert)
	if tx.Error != nil {
		return tx.Error
	}
	fmt.Printf("Inserted %d room(s) into database\n", tx.RowsAffected)

	majorConstraints = make([]models.ConstraintMajor, len(rawMajorConstraints))
	i = 0
	for _, f := range found {
		majorConstraints[i] = f
		i++
	}
	for _, in := range toInsert {
		majorConstraints[i] = in
		i++
	}

	return nil
}

func findMajorConstraintId(majorId uint, addon string) uint {
	for _, mc := range majorConstraints {
		if mc.MajorID.Int32 == int32(majorId) && mc.AddonData == addon {
			return mc.ID
		}
	}

	return 0
}
