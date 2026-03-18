package handles

import (
	"fmt"
	"kano/internal/utils/messageutil"
	"strings"
)

const SPACE_INDENT = 6

// TODO: Add webpage to see the help message since the message
// display may vary depending on device's screen size

var commandListStr string = ""

func HelpHandler(c *messageutil.MessageContext) error {
	// if ctx.Instance.ChatJID().Server != types.DefaultUserServer {
	// 	ctx.Instance.Reply("Bukanya di private chat aja yah 🙂", false)
	// 	return
	// }

	args := c.Parser.Args
	if len(args) > 0 {
		queryCommand := args[0].Content.Data
		foundCommand, ok := mappedCommands[queryCommand]
		if !ok {
			c.QuoteReply("Command \"%s\" is not found!", queryCommand)
			return nil
		}

		commandMan := foundCommand.Man
		if len(commandMan.Name) == 0 {
			c.QuoteReply("Docs entry for \"%s\" is empty", queryCommand)
			return nil
		}

		pref := c.Parser.Command.UsedPrefix
		SPACE := strings.Repeat(" ", SPACE_INDENT)

		// NAME
		var msg strings.Builder
		fmt.Fprintf(&msg, "*NAME*\n%s%s\n\n", SPACE, commandMan.Name)

		// SYNOPSIS
		msg.WriteString("*SYNOPSIS*\n")
		for _, s := range commandMan.Synopsis {
			fmt.Fprintf(&msg, "%s%s%s\n", SPACE, pref, s)
		}
		msg.WriteString("\n")

		// ALIASES
		if len(foundCommand.Aliases) > 0 {
			msg.WriteString("*ALIASES*\n")
			for _, a := range foundCommand.Aliases {
				fmt.Fprintf(&msg, "%s%s%s\n", SPACE, pref, a)
			}
			msg.WriteString("\n")
		}

		// DESCRIPTION
		msg.WriteString("*DESCRIPTION*\n")
		for _, d := range commandMan.Description {
			d = strings.ReplaceAll(d, "\n", "\n"+SPACE)
			d = strings.ReplaceAll(d, "{SPACE}", SPACE)

			fmt.Fprintf(&msg, "%s%s\n\n", SPACE, d)
		}

		// SOURCE CODE LINK
		fmt.Fprintf(&msg, "*SOURCE CODE*\n%shttps://github.com/ziprawan/opc-kano/tree/rebase/internal/message/handles/%s\n\n", SPACE, commandMan.SourceFilename)

		// SEE ALSO
		if len(commandMan.SeeAlso) > 0 {
			msg.WriteString("*SEE ALSO*\n")
			for idx, a := range commandMan.SeeAlso {
				switch a.Type {
				case SeeAlsoTypeCommand:
					fmt.Fprintf(&msg, "%s%d. %s%s command\n", SPACE, idx+1, pref, a.Content)
				case SeeAlsoTypeExternalLink:
					fmt.Fprintf(&msg, "%s%d. Link: %s\n", SPACE, idx+1, a.Content)
				}
			}
		}

		c.QuoteReply("%s", msg.String())
		return nil
	} else {
		prefixes := make([]string, len(c.Parser.Prefixes))
		for i, p := range c.Parser.Prefixes {
			prefixes[i] = fmt.Sprintf("`%s`", p)
		}

		c.QuoteReply(
			"Welcome to my most simple help message!\n_Tbh, I don't know how to design a help message, try `%s %s` for more info, I guess._\nCommand prefixes currently used include:\n%s\n\nBelow is a list of available commands:\n%s\nSource code bot: https://github.com/ziprawan/opc-kano",
			c.Parser.Command.Raw.Data, c.Parser.Command.Name.Data, strings.Join(prefixes, " "), commandListStr,
		)
	}

	return nil
}

var HelpMan = CommandMan{
	Name: "help - a simple command reference manuals",
	Synopsis: []string{
		"*help* [ _command_ ]",
	},
	Description: []string{
		"Help is a very simple command reference manual, with the entire help functionality inspired by the manpage.",
		"Some terms that will be used:" +
			"\n- Prefix: one or more characters placed before a message as one of many ways to trigger a bot, usually followed by the command name." +
			"\n- Command: a word (without spaces, of course) that indicates the command to be executed by the bot if it is recognized." +
			"\n- Argument: a space-separated set of values passed to the command function to perform a more specific action, if any. If a command requires EXACTLY ONE argument and the argument contains at least one space, wrap the argument with quotation marks." +
			"\n- Named argument: an argument type that pairs a key and value with an equals sign (=). If the value contains at least one space, enclose and enclose the argument without quotation marks." +
			"\n- Space: a character that visually separates the characters mediated by it. The characters defined as spaces are SPACE (U+0020), LINE FEED (U+000A), CARRIAGE RETURN (U+000D), CHARACTER TABULATION (U+0009), LINE TABULATION (U+000B)." +
			"\n- Quotation marks: punctuation marks used in pairs to identify direct speech, quotations, or phrases, or in this context, to identify spaced arguments. The characters defined as quotation marks are as defined in the following source code: https://github.com/ziprawan/opc-kano/blob/rebase/internal/utils/word/isquote.go#L8-L17",
		"Conventional section names include *NAME*, *SYNOPSIS*, *ALIASES*, *DESCRIPTION*, *SOURCE CODE*, and *SEE ALSO*.",
		"The following conventions apply to the *SYNOPSIS* section and can be used as a guide in other sections.",
		"*bold text* - type exactly as shown." +
			"\n_italic text_ - replace with appropriate argument." +
			"\n[abc] - any or all arguments within [ ] are optional." +
			"\na|b - options delimited by | cannot be used together." +
			"\n_argument_ ... - argument is repeatable." +
			"\n[ _expression_ ] ... - entire expression within [ ] is repeatable.",
	},

	SourceFilename: "help.go",
	SeeAlso: []SeeAlso{
		{"https://www.man7.org/linux/man-pages/man1/man.1.html", SeeAlsoTypeExternalLink},
		{"man", SeeAlsoTypeCommand},
	},
}
