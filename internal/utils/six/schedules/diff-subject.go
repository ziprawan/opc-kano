package schedules

import (
	"errors"
	"fmt"
	"kano/internal/database/models"
	"slices"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm/clause"
)

// Should change this later. I still manually inserting curricula year into database.
// TODO: Put this in each subject struct instead and automatically insert this to database.
const CURRICULA_YEAR = 2024

func initSubjects(scheds []SemesterSubject) error {
	subjects := []models.Subject{}
	for _, semester := range scheds {
		for _, subject := range semester.Subjects {
			if !slices.ContainsFunc(subjects, func(a models.Subject) bool { return a.Code == subject.Code && a.CurriculaYear == CURRICULA_YEAR }) {
				subjects = append(subjects, models.Subject{
					ID:            subject.ID,
					Code:          subject.Code,
					Name:          subject.Name,
					SKS:           subject.SKS,
					CurriculaYear: CURRICULA_YEAR,
				})
			}
		}
	}

	tx := db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&subjects)
	if tx.Error != nil {
		var pgErr *pgconn.PgError
		fmt.Printf("The error type: %T\n", tx.Error)
		if errors.As(tx.Error, &pgErr) {
			fmt.Println("Code:", pgErr.Code)
			fmt.Println("Constraint:", pgErr.ConstraintName)
			fmt.Println("Column:", pgErr.ColumnName)
			fmt.Println("Table:", pgErr.TableName)
			fmt.Println("Detail:", pgErr.Detail)
		}
	}

	return tx.Error
}
