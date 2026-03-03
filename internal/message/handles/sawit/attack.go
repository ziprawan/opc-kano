package sawit

import (
	"fmt"
	"kano/internal/database/models"
	"kano/internal/utils/messageutil"

	"go.mau.fi/whatsmeow/proto/waE2E"
)

func Attack(c *messageutil.MessageContext, attackValue uint) error {
	partId, err := c.GetParticipantID()
	if err != nil {
		c.QuoteReply("%s", err)
		return err
	}

	partSawit, err := GetParticipantSawit(partId)
	if err != nil {
		c.QuoteReply("Failed to get participant's sawit: %s", err)
		return err
	}

	if partSawit.Height < int(attackValue) {
		c.QuoteReply("Your attack size is higher than your sawit height (%d > %d)", attackValue, partSawit.Height)
		return nil
	}

	resp, err := c.QuoteReply("%s challenged the chat with *%d* cm!\nReact with any emoji to accept the challenge.", partSawit.GetName(), attackValue)
	if err != nil {
		return err
	}

	toInsert := models.SawitAttack{
		ParticipantId: partId,
		GroupId:       c.Group.ID,
		MessageId:     resp.ID,
		AttackSize:    attackValue,
	}
	tx := db.Create(&toInsert)
	if tx.Error != nil {
		msg := fmt.Sprintf("Failed to save sawit attack info: %s", tx.Error)
		c.EditMessageWithID(resp.ID, &waE2E.Message{Conversation: &msg})

		return err
	}

	return nil
}
