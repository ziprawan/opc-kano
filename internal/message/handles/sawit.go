package handles

import (
	"kano/internal/message/handles/sawit"
	"kano/internal/utils/messageutil"
	"strconv"
)

func SawitHandler(c *messageutil.MessageContext) error {
	if c.Group == nil {
		c.QuoteReply("You can only use it in group chats")
		return nil
	}

	if !c.Group.GroupSettings.IsGameAllowed {
		c.Logger.Debugf("Game is not allowed in %s", c.Group.JID)
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
	case "draobredael", "bl":
		return sawit.Draobredael(c)
	case "stat", "sta", "st", "s":
		return sawit.Stat(c)
	default:
		theNum, err := strconv.ParseUint(cmd, 10, 0)
		if err != nil {
			c.QuoteReply("Invalid sawit command %s", cmd)
			return nil
		} else {
			return sawit.Attack(c, uint(theNum))
		}
	}
}
