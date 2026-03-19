package handles

import (
	"kano/internal/utils/messageutil"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

func Ta(ctx *messageutil.MessageContext) error {
	ctx.SendMessage(&waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String("@all"),
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:       proto.String(ctx.GetID()),
				Participant:    proto.String(ctx.GetSender().String()),
				QuotedMessage:  ctx.GetCleanMessage(),
				NonJIDMentions: proto.Uint32(1),
			},
		},
	})
	return nil
}

var TaMan = CommandMan{
	Name:     "ta - tag all",
	Synopsis: []string{"*ta*"},
	Description: []string{
		"This command was originally created for testing purposes. It mentions all members in the group using WhatsApp’s built-in “mention everyone” feature.",
		"Note that some accounts and devices may not support this feature, so it is possible that certain members will not receive the mention when this command is used.",
	},
	SourceFilename: "ta.go",
	SeeAlso:        []SeeAlso{},
}
