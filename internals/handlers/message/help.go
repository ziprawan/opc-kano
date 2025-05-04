package message

import (
	"fmt"
	"strings"

	"go.mau.fi/whatsmeow/types"
)

func (ctx *MessageContext) HelpHandler() {
	if ctx.Instance.ChatJID().Server != types.DefaultUserServer {
		ctx.Instance.Reply("Bukanya di private chat aja yah 🙂", true)
		return
	}

	msg := fmt.Sprintf("Selamat datang di Simplest Kano Help!\nPrefix perintah yang digunakan saat ini adalah:\n%s\n\nBerikut adalah list perintah yang tersedia:\n", strings.Join(ctx.Parser.Prefixes, ""))
	for cmd := range MESSAGE_HANDLERS {
		msg += "- " + cmd + "\n"
	}
	msg += "\nSource code bot: https://github.com/ziprawan/opc-kano"
	ctx.Instance.Reply(strings.TrimSpace(msg), true)
}
