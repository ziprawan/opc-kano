package schedules

import (
	"database/sql"
	"fmt"
	"kano/internal/database/models"

	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func handleAdded(dbx *gorm.DB, semsId uint, subjectAdds []Subject) error {
	totalClasses, totalConstraints, totalSchedules := counter(subjectAdds)

	classesToInsert := make([]models.SubjectClass, 0, totalClasses)
	constraintsToInsert := make([]models.ClassConstraint, 0, totalConstraints)
	schedulesToInsert := make([]models.ClassSchedule, 0, totalSchedules)
	for _, s := range subjectAdds {
		for _, c := range s.Classes {
			sc, err := classModel(semsId, s.ID, c)
			if err != nil {
				return fmt.Errorf("classModel %d: %s", c.ID, err)
			}
			classesToInsert = append(classesToInsert, sc)

			if c.Constraints.Valid {
				cnstrnt, err := constraintModel(c)
				if err != nil {
					return fmt.Errorf("constraintModel %d: %s", c.ID, err)
				}
				constraintsToInsert = append(constraintsToInsert, cnstrnt)
			}

			scheds, err := schedulesModel(c)
			if err != nil {
				return fmt.Errorf("schedulesModel %d: %s", c.ID, err)
			}
			schedulesToInsert = append(schedulesToInsert, scheds...)
		}
	}

	if len(classesToInsert) > 0 {
		tx := dbx.Clauses(clause.OnConflict{UpdateAll: true}).CreateInBatches(&classesToInsert, 500)
		if tx.Error != nil {
			return fmt.Errorf("failed to insert classes: %s", tx.Error)
		}
		if tx.RowsAffected != int64(totalClasses) {
			return fmt.Errorf("insert class: invalid total of affected rows is %d when the expected is %d rows", tx.RowsAffected, totalClasses)
		}
	}

	if len(constraintsToInsert) > 0 {
		tx := dbx.Clauses(clause.OnConflict{UpdateAll: true}).CreateInBatches(&constraintsToInsert, 500)
		if tx.Error != nil {
			return fmt.Errorf("failed to insert constraints: %s", tx.Error)
		}
		if tx.RowsAffected != int64(totalConstraints) {
			return fmt.Errorf("insert constraint: invalid total of affected rows is %d when the expected is %d rows", tx.RowsAffected, totalClasses)
		}
	}

	if len(schedulesToInsert) > 0 {
		tx := dbx.Clauses(clause.OnConflict{UpdateAll: true}).CreateInBatches(&schedulesToInsert, 500)
		if tx.Error != nil {
			return fmt.Errorf("failed to insert schedules: %s", tx.Error)
		}
		if tx.RowsAffected != int64(totalSchedules) {
			return fmt.Errorf("insert schedules: invalid total of affected rows is %d when the expected is %d rows", tx.RowsAffected, totalClasses)
		}
	}

	return nil
}

// Counter
func counter(subjectAdds []Subject) (int, int, int) {
	totalClasses := 0
	totalConstraints := 0
	totalSchedules := 0

	for _, s := range subjectAdds {
		for _, c := range s.Classes {
			// Class total
			totalClasses++

			// Class that has constraints
			if c.Constraints.Valid {
				totalConstraints++
			}

			// Schedules total from all classes
			totalSchedules += len(c.Schedules)
		}
	}

	return totalClasses, totalConstraints, totalSchedules
}

// Modelling
func classModel(semsId uint, sId uint, c Class) (models.SubjectClass, error) {
	theLecturers, err := lecturerMapModel(c.Lecturers)
	if err != nil {
		return models.SubjectClass{}, fmt.Errorf("lecturerModel: %s", err)
	}

	sc := models.SubjectClass{
		ID:                 c.ID,
		Number:             c.Number,
		Quota:              c.Quota,
		EdunexClassID:      c.Links.EdunexClassId,
		AvailableAtMajorID: c.AvailableAtMajorId,
		SubjectID:          sId,
		SemesterID:         semsId,
		Lecturers:          theLecturers,
	}

	if c.Links.Teams != "" {
		sc.TeamsLink = sql.NullString{String: c.Links.Teams, Valid: true}
	}

	return sc, nil
}

func constraintModel(c Class) (models.ClassConstraint, error) {
	res := models.ClassConstraint{}
	theConstraint := c.Constraints
	if !theConstraint.Valid {
		return res, fmt.Errorf("source constraint is not valid")
	}
	srcConstraint := theConstraint.Constraint

	res.SubjectClassID = c.ID
	res.Other = srcConstraint.Others
	res.Faculties = srcConstraint.Faculties
	res.Semester = semesterMapModel(srcConstraint.Semesters)
	res.Campuses = campusMapModel(srcConstraint.Campuses)
	res.Stratas = strataMapModel(srcConstraint.Stratas)

	var err error
	res.Majors, err = constraintMajorMapmodel(srcConstraint.Majors)
	if err != nil {
		return res, fmt.Errorf("constraintMajorMapModel: %s", err)
	}

	return res, nil
}

func schedulesModel(c Class) ([]models.ClassSchedule, error) {
	srcScheds := c.Schedules
	scheds := make([]models.ClassSchedule, len(srcScheds))
	if len(c.Schedules) == 0 {
		return scheds, nil
	}

	for i := range srcScheds {
		sched := srcScheds[i]
		scheds[i] = models.ClassSchedule{
			Start:    sched.Start,
			End:      sched.End,
			Activity: models.ScheduleActivity(sched.Activity),
			Method:   models.ScheduleMethod(sched.Method),

			UnixStart: uint(sched.Start.Unix()),
			UnixEnd:   uint(sched.End.Unix()),

			SubjectClassID: c.ID,
		}

		var err error
		scheds[i].Rooms, err = roomMapModel(sched.Rooms)
		if err != nil {
			return nil, fmt.Errorf("roomMapModel %d: %s", i, err)
		}
	}

	return scheds, nil
}

// Mapper
func semesterMapModel(cs []uint) pq.Int32Array {
	res := make([]int32, len(cs))
	for i := range cs {
		res[i] = int32(cs[i])
	}

	return pq.Int32Array(res)
}

func campusMapModel(c []Campus) []models.Campus {
	r := make([]models.Campus, len(c))
	for i := range c {
		r[i] = models.Campus(c[i])
	}
	return r
}

func strataMapModel(s []Strata) []models.Strata {
	r := make([]models.Strata, len(s))
	for i := range s {
		r[i] = models.Strata(s[i])
	}
	return r
}

func lecturerMapModel(lecturers []string) ([]models.Lecturer, error) {
	if len(lecturers) == 0 {
		return []models.Lecturer{}, nil
	}

	res := make([]models.Lecturer, len(lecturers))
	for i := range lecturers {
		name := lecturers[i]
		id := findLecturerId(name)
		if id == 0 {
			return nil, fmt.Errorf("unable to find lecturer id for name=%q", name)
		}
		res[i] = models.Lecturer{ID: id}
	}

	return res, nil
}

func constraintMajorMapmodel(m []MajorConstraint) ([]models.ConstraintMajor, error) {
	r := make([]models.ConstraintMajor, len(m))
	for i := range m {
		mc := m[i]
		id := findMajorConstraintId(mc.ID, mc.Addon)
		if id == 0 {
			return nil, fmt.Errorf("unable to find major constraint id for majorID=%d, addon=%q", mc.ID, mc.Addon)
		}
		r[i] = models.ConstraintMajor{ID: id}
	}
	return r, nil
}

func roomMapModel(rooms []string) ([]models.Room, error) {
	if len(rooms) == 0 {
		return []models.Room{}, nil
	}

	res := make([]models.Room, len(rooms))
	for i := range rooms {
		name := rooms[i]
		id := findRoomId(name)
		if id == 0 {
			return nil, fmt.Errorf("unable to find major id for name=%q", name)
		}
		res[i] = models.Room{ID: id}
	}

	return res, nil
}
