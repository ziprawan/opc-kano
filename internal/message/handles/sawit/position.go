package sawit

import (
	"fmt"
	"kano/internal/database/models"
)

func GetParticipantPosition(groupId, participantId uint) (uint, error) {
	founds := []models.Sawit{}
	tx := db.
		Joins("JOIN participant ON participant.id = sawit.participant_id").
		Where("participant.group_id = ?", groupId).
		Order("height DESC, updated_at ASC").
		Select("participant_id", "height").
		Find(&founds)
	if tx.Error != nil {
		return 0, fmt.Errorf("Failed to get all sawits: %s", tx.Error)
	}

	pos := int(1)
	if len(founds) == 0 {
		return uint(pos), nil
	}

	for i, f := range founds {
		if f.ParticipantId == participantId {
			pos = i + 1
			break
		}
	}

	return uint(pos), nil
}
