package wordle

import (
	"fmt"
	"kano/internal/database"
	"kano/internal/database/models"
	"math/rand"
)

func randomSelectWordle() (models.Wordle, error) {
	db := database.GetInstance()
	var ids []uint
	tx := db.Model(&models.Wordle{}).Select("id").Find(&ids)
	if err := tx.Error; err != nil {
		return models.Wordle{}, fmt.Errorf("randomizer: Failed to get words: %s", err)
	}

	selectedIdx := rand.Intn(len(ids))
	selectedWordleId := ids[selectedIdx]

	wordle := models.Wordle{Id: selectedWordleId}
	tx = db.First(&wordle)
	if err := tx.Error; err != nil {
		return wordle, fmt.Errorf("randomizer: Failed to get word by id: %s", err)
	}

	return wordle, nil
}
