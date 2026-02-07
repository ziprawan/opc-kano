package schedules

import (
	"fmt"
	"kano/internal/database/models"
	"kano/internal/utils/datetime"
	"slices"
	"time"
)

// Self reminder: diff generator is just get something from database and generate the diff from them,
// never insert or update them
// Except for room, lecturer, majorConstraint, and subject, since they are all mandatory
// for class schedule foreign key

var dbSems models.Semester
var semsWeekStart time.Time

func generateSemesterDiff(semester SemesterSubject) (SemesterDiff, error) {
	res := SemesterDiff{}

	tx := db.
		Where(models.Semester{Year: semester.Year, Semester: semester.Semester}).
		FirstOrCreate(&dbSems)
	if tx.Error != nil {
		return res, tx.Error
	}
	res.ID = dbSems.ID
	semsWeekStart = datetime.StartOfWeek(dbSems.Start)

	// Ts might slow af
	var dbClasses []models.SubjectClass
	q := db.Preload("Subject").Select("subject_id").Group("subject_id")
	tx = q.Find(&dbClasses)
	if tx.Error != nil {
		return res, tx.Error
	}

	// Find removed subjects
	removed := make([]Subject, 0, len(dbClasses))
	for _, dbClass := range dbClasses {
		if !slices.ContainsFunc(semester.Subjects, func(a Subject) bool { return a.ID == dbClass.SubjectID }) {
			removed = append(removed, Subject{
				BasicSubject: BasicSubject{
					ID:   dbClass.Subject.ID,
					Code: dbClass.Subject.Code,
					Name: dbClass.Subject.Name,
					SKS:  dbClass.Subject.SKS,
				},
			})
		}
	}

	// Find added subjects
	exists := make([]Subject, 0, len(semester.Subjects))
	added := make([]Subject, 0, len(semester.Subjects))
	for _, subject := range semester.Subjects {
		if !slices.ContainsFunc(dbClasses, func(a models.SubjectClass) bool { return a.SubjectID == subject.ID }) {
			added = append(added, subject)
		} else {
			exists = append(exists, subject)
		}
	}

	// Debugging purpose
	fmt.Println(len(semester.Subjects))
	fmt.Println(len(added))
	fmt.Println(len(removed))
	fmt.Println(len(exists))

	// Find modified subject (and their classes :D)
	res.ModifiedSubjects = make([]SubjectDiff, 0, len(exists)) // Pre allocate
	for _, mod := range exists {
		subDiff, err := generateSubjectDiff(mod)
		if err != nil {
			return res, fmt.Errorf("subjectDiff %s: %s", mod.Code, err)
		}

		anyDiff := len(subDiff.AddedClasses) > 0 ||
			len(subDiff.RemovedClasses) > 0 ||
			len(subDiff.ModifiedClasses) > 0

		if anyDiff {
			res.ModifiedSubjects = append(res.ModifiedSubjects, subDiff)
		}
	}

	res.AddedSubjects = added
	res.RemovedSubjects = removed

	return res, nil
}
