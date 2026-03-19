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
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &settings, nil
}

func (gs *GroupSettings) Save() error {
	settings := models.GroupSettings{
		ID:               gs.ID,
		IsGameAllowed:    gs.IsGameAllowed,
		IsConfessAllowed: gs.IsConfessAllowed,
	}

	db := database.GetInstance()
	tx := db.Save(&settings)

	return tx.Error
}
