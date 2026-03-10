package handles

import (
	"kano/internal/utils/messageutil"
	"slices"
	"strings"
)

func GameHandler(c *messageutil.MessageContext) error {
	enables := []string{"true", "on", "yes", "1"}
	disables := []string{"false", "off", "no", "0"}

	if c.Group == nil {
		c.QuoteReply("You can only set game allowance in groups.")
		return nil
	}

	args := c.Parser.Args
	if len(args) == 0 {
		c.QuoteReply("Is game allowed in this group? %t", c.Group.Settings.IsGameAllowed)
		return nil
	}

	inp := strings.ToLower(args[0].Content.Data)

	isEnabled := slices.Contains(enables, inp)
	isDisabled := slices.Contains(disables, inp)

	if !isEnabled && !isDisabled {
		c.QuoteReply("Input is not valid.\nUse %s to enable.\nUse %s to disable.", strings.Join(enables, "/"), strings.Join(disables, "/"))
		return nil
	}

	if isEnabled && isDisabled {
		c.QuoteReply("Internal error: isEnabled and isDisabled are both true.")
		return nil
	}

	e := ""
	if isEnabled {
		c.Group.GroupSettings.IsGameAllowed = true
		e = "enabled"
	}
	if isDisabled {
		c.Group.GroupSettings.IsGameAllowed = false
		e = "disabled"
	}

	err := c.Group.GroupSettings.Save()
	if err != nil {
		c.QuoteReply("Internal error.\nDebug: %s", err)
		return err
	}
	c.QuoteReply("Game is %s in this group from now.", e)

	return nil
}
