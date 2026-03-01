package handles

import (
	"kano/internal/message/handles/sawit"
	"kano/internal/utils/messageutil"
)

func SawitHandler(c *messageutil.MessageContext) error {
	if c.Group.ID == 0 {
		c.QuoteReply("You can only use it in group chats")
		return nil
	}

	args := c.Parser.Args
	if len(args) == 0 {
		return sawit.Grow(c)
	}

	cmd := args[0].Content.Data
	switch cmd {
	case "grow", "g":
		return sawit.Grow(c)
	case "leaderboard", "lb", "l":
		return sawit.Leaderboard(c)
	}

	return nil
}
