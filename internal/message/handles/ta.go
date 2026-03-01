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
