package message

import (
	"context"
	projectconfig "kano/internals/project_config"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

func (ctx MessageContext) ConfessHandler() {
	if ctx.Instance.Event.Info.Chat.Server != types.DefaultUserServer {
		return
	}

	conf := projectconfig.GetConfig()
	args := ctx.Parser.GetArgs()

	if !conf.ConfessTarget.Valid {
		return
	}

	if len(args) == 0 {
		ctx.Instance.Reply("Beri pesannya dong kak (belum support gambar lagi)", true)
		return
	}

	confessMsg := "Ada konfes dari seseorang nih!\n" + ctx.Parser.Text[args[0].Start:]

	ctx.Instance.Client.SendMessage(context.Background(), conf.ConfessTarget.JID, &waE2E.Message{
		Conversation: &confessMsg,
	})

}
