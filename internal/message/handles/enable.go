package handles

import (
	"kano/internal/database/models"
	"kano/internal/utils/messageutil"
	"slices"
	"strings"
)

func EnableHandler(c *messageutil.MessageContext) error {
	if c.Group == nil {
		return nil
	}
	part, err := c.Group.GetParticipantByContactId(c.Contact.ID)
	if err != nil {
		c.QuoteReply("Failed to get participant info: %s", err)
		return err
	}
	role := part.Role
	if role != models.ParticipantRoleAdmin && role != models.ParticipantRoleSuperadmin {
		c.QuoteReply("Your role in this group is not admin or superadmin. (%d=%s)", part.ID, role)
		return nil
	}

	isEnable := c.Parser.Command.Name.Data == "enable"
	allowedArgs := []string{"game", "confess"}

	args := c.Parser.Args
	if len(args) == 0 {
		// Show the current configuration
		c.QuoteReply(
			"Current configuration:\n\n"+
				"[game] Is Game Allowed? %t\n"+
				"[confess] Is Confess Allowed? %t",
			c.Group.GroupSettings.IsGameAllowed,
			c.Group.GroupSettings.IsConfessAllowed,
		)
	} else {
		p := strings.ToLower(args[0].Content.Data)
		if !slices.Contains(allowedArgs, p) {
			c.QuoteReply("Invalid argument. Accepted arguments are: %s.", strings.Join(allowedArgs, "; "))
		} else {
			switch p {
			case "game":
				c.Group.GroupSettings.IsGameAllowed = isEnable
			case "confess":
				c.Group.GroupSettings.IsConfessAllowed = isEnable
			default:
				c.QuoteReply("Unhandled argument %q. Please report it to the developer!", p)
				return nil
			}

			err := c.Group.GroupSettings.Save()
			if err != nil {
				c.QuoteReply("Failed to save group settings: %s", err)
				return err
			}

			c.QuoteReply(
				"Configuration saved! Current configuration:\n\n"+
					"[game] Is Game Allowed? %t\n"+
					"[confess] Is Confess Allowed? %t",
				c.Group.GroupSettings.IsGameAllowed,
				c.Group.GroupSettings.IsConfessAllowed,
			)
		}
	}

	return nil
}

var EnableMan = CommandMan{
	Name: "enable - enable feature",
	Synopsis: []string{
		"*enable* [ _config_name_ ]",
		"*disable* [ _config_name_ ]",
	},
	Description: []string{
		"Configure your group to enable some feature. Use `disable` to do the opposite.",
		"_config_name_" +
			"\n{SPACE}Currently only accepts `game` and `confess`. `game` to enable game in the group and `confess` to allow user make a confess to this group.",
	},
	SourceFilename: "enable.go",
	SeeAlso: []SeeAlso{
		{"game", SeeAlsoTypeCommand},
		{"confess", SeeAlsoTypeCommand},
	},
}
