package handles

import (
	"fmt"
	"kano/internal/utils/messageutil"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

func Test(ctx *messageutil.MessageContext) error {
	s, err := ctx.QuoteReply("Hai")
	if err != nil {
		return nil
	}
	fmt.Printf("%+v\n", s)

	ctx.SendMessage(&waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String("reply"),
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:    proto.String(s.ID),
				Participant: proto.String(s.Sender.ToNonAD().String()),
				QuotedMessage: &waE2E.Message{
					Conversation: proto.String("This is placeholder message, if you are seeing this, maybe the replied message is too old."),
				},
				QuotedType: waE2E.ContextInfo_EXPLICIT.Enum(),
			},
		},
	})

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
