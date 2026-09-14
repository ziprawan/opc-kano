package schedules

import (
	"kano/internal/database/models"

	"gorm.io/gorm"
)

func handleRemoved(dbx *gorm.DB, semId uint, removed []Subject) error {
	mapped := make([][2]uint, len(removed))
	for i := range removed {
		mapped[i] = [2]uint{semId, removed[i].ID}
	}

	if len(mapped) > 0 {
		tx := dbx.
			Unscoped().
			Where("(semester_id, subject_id) IN ?", mapped).
			Delete(&models.SubjectClass{})
		return tx.Error
	}

	return nil
}
