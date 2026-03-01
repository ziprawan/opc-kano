package wordle

import (
	"errors"
	"kano/internal/database"
	"kano/internal/database/models"
	"kano/internal/utils/definition"

	"gorm.io/gorm"
)

func isWordExists(word string) bool {
	db := database.GetInstance()

	found := models.Wordle{}
	tx := db.Where("word = ?", word).First(&found)

	if err := tx.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dict, _ := definition.FindDefinition(word)
			if dict != nil && len(dict.Results) != 0 {
				found.Word = word
				found.Point = calculateWordPoint(word)
				found.Lang = "en"
				found.IsWordle = false

				db.Create(&found)

				return true
			}
		}

		return false
	} else {
		return true
	}
}
