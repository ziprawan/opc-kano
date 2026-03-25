package handles

import (
	"fmt"
	"kano/internal/utils/messageutil"

	"go.mau.fi/whatsmeow/proto/waCommon"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

func Test(ctx *messageutil.MessageContext) error {
	rectMsg := &waE2E.Message{
		ReactionMessage: &waE2E.ReactionMessage{
			Key: &waCommon.MessageKey{
				RemoteJID: proto.String(ctx.GetChat(true).String()),
				FromMe:    proto.Bool(false),
				ID:        proto.String(ctx.GetID()),
			},
			Text: proto.String("😂"),
		},
	}
	if ctx.Group != nil {
		rectMsg.ReactionMessage.Key.Participant = proto.String(ctx.GetSender().String())
	}
	fmt.Printf("%+v\n", rectMsg)
	_, err := ctx.SendMessage(rectMsg)
	if err != nil {
		ctx.QuoteReply("%s", err)
	}

	return nil
}

var TestMan = CommandMan{
	Name:     "test - test",
	Synopsis: []string{"*test* [ _any_ ]"},
	Description: []string{
		"Just a testing command that can only be used by the owner of the bot.",
	},
	SourceFilename: "test.go",
	SeeAlso:        []SeeAlso{},
}
