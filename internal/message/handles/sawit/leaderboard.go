package sawit

import (
	"fmt"
	"kano/internal/database/models"
	"kano/internal/utils/messageutil"
	"strings"
	"time"
)

func Leaderboard(c *messageutil.MessageContext) error {
	founds := []models.Sawit{}
	tx := db.
		Preload("Participant.Contact").
		Joins("JOIN participant ON participant.id = sawit.participant_id").
		Where("participant.group_id = ?", c.Group.ID).
		Order("sawit.height DESC, sawit.updated_at ASC").
		Select("participant_id", "height", "last_grow_date").
		Limit(10).
		Find(&founds)
	if err := tx.Error; err != nil {
		c.QuoteReply("Failed to get sawits: %s", err)
		return err
	}

	if len(founds) == 0 {
		c.QuoteReply("Nobody grows sawit here :(\nNo wowo impressed.")
		return nil
	}

	now := time.Now().UTC()
	nowDateStr := now.Format("02-01-2006")

	var msg strings.Builder
	msg.WriteString("Top of the tallest sawits:\n\n")
	for i, f := range founds {
		name := f.Participant.Contact.CustomName
		if name == "" {
			name = f.Participant.Contact.PushName
		}
		if name == "" {
			name = fmt.Sprintf("[Unknown User: %s]", f.Participant.Contact.JID.User)
		}
		stat := ""
		if nowDateStr != f.LastGrowDate {
			stat = " [+]"
		}
		fmt.Fprintf(&msg, "%d | *%s* — *%d* cm%s\n", i+1, name, f.Height, stat)
	}
	msg.WriteString("\n_[+] means a grower hasn't grown his sawit today yet._")

	c.QuoteReply("%s", msg.String())
	return nil
}

func Draobredael(c *messageutil.MessageContext) error {
	founds := []models.Sawit{}
	tx := db.
		Preload("Participant.Contact").
		Joins("JOIN participant ON participant.id = sawit.participant_id").
		Where("participant.group_id = ?", c.Group.ID).
		Order("sawit.height DESC, sawit.updated_at DESC").
		Select("participant_id", "height", "last_grow_date").
		Limit(10).
		Find(&founds)
	if err := tx.Error; err != nil {
		c.QuoteReply("Failed to get sawits: %s", err)
		return err
	}

	if len(founds) == 0 {
		c.QuoteReply("Nobody grows sawit here :(\nNo wowo impressed.")
		return nil
	}

	now := time.Now().UTC()
	nowDateStr := now.Format("02-01-2006")

	var msg strings.Builder
	msg.WriteString("Top of the shortest sawits:\n\n")
	for i, f := range founds {
		name := f.Participant.Contact.CustomName
		if name == "" {
			name = f.Participant.Contact.PushName
		}
		if name == "" {
			name = fmt.Sprintf("[Unknown User: %s]", f.Participant.Contact.JID.User)
		}
		stat := ""
		if nowDateStr != f.LastGrowDate {
			stat = " [+]"
		}
		fmt.Fprintf(&msg, "%d | *%s* — *%d* cm%s\n", i+1, name, f.Height, stat)
	}
	msg.WriteString("\n_[+] means a grower hasn't grown his sawit today yet._")

	c.QuoteReply("%s", msg.String())
	return nil
}
