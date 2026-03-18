package handles

import (
	"kano/internal/database/models"
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

	role, err := c.Group.GetParticipantRole(c.Contact.ID)
	if err != nil {
		c.QuoteReply("Failed to get participant role: %s", err)
		return err
	}
	if role != models.ParticipantRoleAdmin && role != models.ParticipantRoleSuperadmin {
		c.QuoteReply("Your role in this group is not admin or superadmin.")
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

	err = c.Group.GroupSettings.Save()
	if err != nil {
		c.QuoteReply("Internal error.\nDebug: %s", err)
		return err
	}
	c.QuoteReply("Game is %s in this group from now.", e)

	return nil
}

var GameMan = CommandMan{
	Name: "game - enable/disable game in group chat",
	Synopsis: []string{
		"*game* [ _set_ ]",
	},
	Description: []string{
		"Enable or disable game in the group chat. This command can only be executed by admin or higher.",
		"When no arguments are given, the bot will return the current game settings whether they are enabled or not.",
		"_set_" +
			"\n{SPACE}Value to set the game's permission within the group. Use *on*/*true*/*yes*/*1* to enable, *off*/*false*/*no*/*0* to disable." +
			"\n{SPACE}If the _set_ argument value is invalid, the bot will return an error message.",
	},
	SourceFilename: "game.go",
	SeeAlso: []SeeAlso{
		{"wordle", SeeAlsoTypeCommand},
		{"sawit", SeeAlsoTypeCommand},
	},
}
