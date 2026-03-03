package handles

import (
	"encoding/json"
	"fmt"
	"kano/internal/utils/messageutil"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

func Test(ctx *messageutil.MessageContext) error {
	mar, _ := json.Marshal(map[string]any{
		"display_text": "Google",
		"url":          "https://www.google.com",
	})

	interactiveMessage := &waE2E.InteractiveMessage{
		Header: &waE2E.InteractiveMessage_Header{
			Title:              proto.String("Ini judul"),
			Subtitle:           proto.String("Ini subjudul"),
			HasMediaAttachment: proto.Bool(false),
		},
		Body: &waE2E.InteractiveMessage_Body{
			Text: proto.String("Ini body"),
		},
		Footer: &waE2E.InteractiveMessage_Footer{
			Text: proto.String("Ini footer"),
		},
		InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
			NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{
				Buttons: []*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton{
					{
						Name:             proto.String("cta_url"),
						ButtonParamsJSON: proto.String(string(mar)),
					},
				},
			},
		},
	}
	fmt.Printf("%+v\n", interactiveMessage)

	_, err := ctx.SendMessage(&waE2E.Message{
		InteractiveMessage: interactiveMessage,
	})
	fmt.Println(err)

	// ctx.Client.GetClient().DangerousInternals().Message

	return nil
}
