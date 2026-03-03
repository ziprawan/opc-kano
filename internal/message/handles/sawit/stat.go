package sawit

import "kano/internal/utils/messageutil"

func Stat(c *messageutil.MessageContext) error {
	partId, err := c.GetParticipantID()
	if err != nil {
		c.QuoteReply("Failed to get participant info: %s", err)
		return err
	}

	partSawit, err := GetParticipantSawit(partId)
	if err != nil {
		c.QuoteReply("Failed to get participant sawit: %s", err)
		return err
	}

	position, err := GetParticipantPosition(c.Group.ID, partId)
	if err != nil {
		c.QuoteReply("Failed to get participant sawit rank position: %s", err)
		return err
	}

	c.QuoteReply(
		"Height: *%d*\nPosition in the top: *%d*\n\nWin rate: *%.02f%%*\nAttacks: *%d*\nWins: *%d*\nAcquired height: *%d cm*\nLost height: *%d cm*",
		partSawit.Height, position, partSawit.GetWinrate()*100, partSawit.AttackTotal, partSawit.AttackWin, partSawit.AttackAcquiredHeight, partSawit.AttackLostHeight,
	)

	return nil
}
