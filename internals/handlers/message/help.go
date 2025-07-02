package message

import (
	"fmt"
	"strings"
)

const SPACE_INDENT = 7

func (ctx *MessageContext) HelpHandler() {
	// if ctx.Instance.ChatJID().Server != types.DefaultUserServer {
	// 	ctx.Instance.Reply("Bukanya di private chat aja yah 🙂", false)
	// 	return
	// }

	args := ctx.Parser.GetArgs()
	if len(args) > 0 {
		queryCommand := args[0].Content
		foundCommand, ok := mappedCommands[queryCommand]
		if !ok {
			ctx.Instance.Reply(fmt.Sprintf("Perintah \"%s\" tidak ditemukan!", queryCommand), false)
			return
		}

		commandMan := foundCommand.Man
		if len(commandMan.Name) == 0 {
			ctx.Instance.Reply(fmt.Sprintf("Docs entry for \"%s\" is empty", queryCommand), false)
			return
		}

		pref := ctx.Parser.GetCommand().UsedPrefix
		SPACE := strings.Repeat(" ", SPACE_INDENT)

		// NAME
		msg := fmt.Sprintf("*NAME*\n%s%s\n\n", SPACE, commandMan.Name)

		// SYNOPSIS
		msg += "*SYNOPSIS*\n"
		for _, s := range commandMan.Synopsis {
			msg += fmt.Sprintf("%s%s%s\n", SPACE, pref, s)
		}
		msg += "\n"

		// ALIASES
		if len(foundCommand.Aliases) > 0 {
			msg += "*ALIASES*\n"
			for _, a := range foundCommand.Aliases {
				msg += fmt.Sprintf("%s%s%s\n", SPACE, pref, a)
			}
			msg += "\n"
		}

		// DESCRIPTION
		msg += "*DESCRIPTION*\n"
		for _, d := range commandMan.Description {
			d = strings.ReplaceAll(d, "\n", "\n"+SPACE)
			d = strings.ReplaceAll(d, "{SPACE}", SPACE)

			msg += fmt.Sprintf("%s%s\n\n", SPACE, d)
		}

		// SOURCE CODE LINK
		msg += fmt.Sprintf("*SOURCE CODE*\n%shttps://github.com/ziprawan/opc-kano/tree/main/internals/handlers/message/%s\n\n", SPACE, commandMan.Source)

		// SEE ALSO
		if len(commandMan.SeeAlso) > 0 {
			msg += "*SEE ALSO*\n"
			for idx, a := range commandMan.SeeAlso {
				switch a.Type {
				case SeeAlsoTypeCommand:
					msg += fmt.Sprintf("%s%d. Perintah %s%s\n", SPACE, idx+1, pref, a.Content)
				case SeeAlsoTypeExternalLink:
					msg += fmt.Sprintf("%s%d. Link: %s\n", SPACE, idx+1, a.Content)
				}
			}
		}

		ctx.Instance.Reply(msg, false)
		return
	}

	msg := fmt.Sprintf("Selamat datang di Simplest Kano Help!\nPrefix perintah yang digunakan saat ini adalah:\n%s\n\nBerikut adalah list perintah yang tersedia:\n", strings.Join(ctx.Parser.Prefixes, ""))
	for cmd := range MESSAGE_HANDLERS {
		msg += "- " + cmd
		aliases := MESSAGE_HANDLERS[cmd].Aliases
		if len(aliases) > 0 {
			msg += fmt.Sprintf(" (%s)", strings.Join(aliases, ", "))
		}
		msg += "\n"
	}
	msg += "\nSource code bot: https://github.com/ziprawan/opc-kano"
	ctx.Instance.Reply(strings.TrimSpace(msg), false)
}
