package grouputil

import (
	"context"
	"fmt"
	"kano/internal/database"
	"kano/internal/database/models"
	"kano/internal/utils/chatutil/contactutil"

	"go.mau.fi/whatsmeow/types"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Participant struct {
	models.Participant
}

func (g Group) UpdateParticipantList(grpInfo *types.GroupInfo) error {
	if len(grpInfo.Participants) > 0 {
		log.Debugf("Got %d participant(s)", len(grpInfo.Participants))

		participants := make([]types.JID, len(grpInfo.Participants))
		roles := make([]models.ParticipantRole, len(grpInfo.Participants))
		for i, part := range grpInfo.Participants {
			participants[i] = part.JID
			if part.IsSuperAdmin {
				roles[i] = models.ParticipantRoleAdmin
			} else if part.IsAdmin {
				roles[i] = models.ParticipantRoleAdmin
			} else {
				roles[i] = models.ParticipantRoleMember
			}
		}

		ids, err := contactutil.GetIDs(participants)
		if err != nil {
			log.Errorf("Failed to get contact IDs: %s", err.Error())
			return err
		}

		partModels := make([]models.Participant, len(participants))
		i := 0
		for _, id := range ids {
			partModels[i].GroupID = g.ID
			partModels[i].ContactID = id
			partModels[i].Role = roles[i]
			i++
		}

		db := database.GetInstance()
		tx := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&partModels)
		if tx.Error != nil {
			log.Errorf("Failed to insert participants")
		}
		return tx.Error
	}

	return nil
}

func (g Group) GetParticipantByJID(userJid types.JID) (Participant, error) {
	db := database.GetInstance()

	part := Participant{}

	res := db.Table("participant").Joins("JOIN contact ON contact.id = participant.contact_id").
		Where("participant.group_id = ?", g.ID).
		Where("contact.jid = ?", userJid.String()).
		First(&part)

	if res.Error != nil {
		return part, res.Error
	}

	return part, nil
}

func (g Group) GetParticipantByContactId(contactId uint) (Participant, error) {
	if g.ID == 0 {
		return Participant{}, fmt.Errorf("Group is not initialized")
	}

	db := database.GetInstance()

	part := Participant{}
	found, err := gorm.G[models.Participant](db).
		Where(
			&models.Participant{
				GroupID:   g.ID,
				ContactID: contactId,
			},
		).
		First(context.Background())
	if err != nil {
		return part, err
	}

	part.Participant = found

	return part, nil
}
