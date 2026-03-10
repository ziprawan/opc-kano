package grouputil

import (
	"kano/internal/database"
	"kano/internal/database/models"
)

type GroupSettings struct {
	models.GroupSettings
}

func InitSettings(groupId uint) (*GroupSettings, error) {
	var settings GroupSettings

	db := database.GetInstance()
	tx := db.Where("id = ?", groupId).FirstOrCreate(&settings)

	return &settings, tx.Error
}

func (gs *GroupSettings) Save() error {
	settings := models.GroupSettings{
		ID:            gs.ID,
		IsGameAllowed: gs.IsGameAllowed,
	}

	db := database.GetInstance()
	tx := db.Save(&settings)

	return tx.Error
}
