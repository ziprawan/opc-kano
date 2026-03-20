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
	case "transfer", "tf", "t":
		if len(args) < 2 {
			c.QuoteReply("Usage: *sawit transfer* <amount> <target>")
			return nil
		}
		transferAmt, err := strconv.ParseUint(args[1].Content.Data, 10, 0)
		if err != nil {
			c.QuoteReply("Invalid transfer amount")
			return nil
		}
		targetJID := args[2].Content.Data[1:]
		return sawit.Transfer(c, uint(transferAmt), targetJID)
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

var SawitMan = CommandMan{
	Name: "sawit — grow your sawit",
	Synopsis: []string{
		"*sawit* [ *g*|*grow* ]",
		"*sawit* _attack size_",
		"*sawit* *l*|*lb*|*leaderboard*",
		"*sawit* *bl*|*draobredael*",
		"*sawit* *s*|*st*|*sta*|*stat*",
		"*sawit* *t*|*tf*|*transfer* _amount_ _target_",
	},
	Description: []string{
		"Sawit is a simple game where you can grow your sawit and place bets with other players. Player data is scoped per group (different groups have separate data).",
		"[ *g*|*grow* ]" +
			"\n{SPACE}Grows your sawit. This action can only be performed once per day and resets at 00:00 UTC. The growth amount is randomly determined within the range of 2 to 20. There is a 10%% probability that the player will be shrunk, which decreases the sawit height instead.",
		"_attack size_" +
			"\n{SPACE}Creates a new bet. The attack size must not exceed the current sawit height. A player cannot initiate a bet if their sawit height is less than or equal to 0.",
		"\n{SPACE}Other players can accept the bet by reacting to the corresponding bot message. The outcome is determined with a 50%% win/loss probability. The winner gains sawit height equal to the attack size, while the loser loses the same amount. This deduction may cause a player's sawit height to become negative, depending on the attack size and their current height.",
		"*l*|*lb*|*leaderboard*" +
			"\n{SPACE}Displays the top 10 tallest sawits along with their owners. A [+] indicator is shown if a player has not grown their sawit for the current day.",
		"*bl*|*draobredael*" +
			"\n{SPACE}Displays the bottom 10 shortest sawits along with their owners. A [+] indicator is shown if a player has not grown their sawit for the current day." +
			"\n{SPACE}_Fun fact: “draobredael” is simply “leaderboard” spelled backwards._",
		"*s*|*st*|*sta*|*stat*" +
			"\n{SPACE}Displays the current status of the player's sawit, including:" +
			"\n{SPACE}- Current sawit height" +
			"\n{SPACE}- Rank in the top leaderboard" +
			"\n{SPACE}- Total bets initiated" +
			"\n{SPACE}- Total wins and losses" +
			"\n{SPACE}- Bet win rate" +
			"\n{SPACE}- Total sawit height gained and lost from bets",
		"_Note: The game is currently due for a redesign due to limited mechanics and lack of creative depth. If you are interested in contributing to a redesign, feel free to reach out._",
	},
	SourceFilename: "sawit.go",
	SeeAlso:        []SeeAlso{},
}
