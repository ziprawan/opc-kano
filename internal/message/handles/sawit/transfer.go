package sawit

import (
	"kano/internal/utils/messageutil"

	"go.mau.fi/whatsmeow/types"
)

func Transfer(c *messageutil.MessageContext, transferAmt uint, targetJID string) error {
	participantId, err := c.GetParticipantID()

	if err != nil {
		c.QuoteReply("%s", err)
		return err
	}

	participantSawit, err := GetParticipantSawit(participantId)
	if err != nil {
		c.QuoteReply("Failed to get participant's sawit: %s", err)
		return err
	}

	if participantSawit.Height < int(transferAmt) {
		c.QuoteReply("Your transfer size is higher than your sawit height (%d > %d)", transferAmt, participantSawit.Height)
		return nil
	}

	targetPart, err := c.Group.GetParticipantByJID(types.NewJID(targetJID, types.HiddenUserServer))

	if err != nil {
		c.QuoteReply("Failed to get target participant: %s", err)
		return err
	}

	targetPartSawit, err := GetParticipantSawit(targetPart.ID)
	if err != nil {
		c.QuoteReply("Failed to get target participant's sawit: %s", err)
		return err
	}

	targetPartSawit.AddHeight(int(transferAmt))
	targetPartSawit.Save()

	participantSawit.AddHeight(-int(transferAmt))
	participantSawit.Save()

	c.QuoteReply("Successfully transferred *%d* cm from %s to %s!", transferAmt, participantSawit.GetName(), targetPartSawit.GetName())

	return nil
}
