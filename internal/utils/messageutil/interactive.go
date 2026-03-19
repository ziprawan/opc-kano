package messageutil

import (
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
)

type MessageButton interface {
	GetNativeFlowButton() *waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton
}

func (c *MessageContext) SendButton(buttons []MessageButton) (whatsmeow.SendResponse, error) {
	interactiveMsg := &waE2E.InteractiveMessage_NativeFlowMessage_{
		NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{
			Buttons: make([]*waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton, len(buttons)),
		},
	}

	for i, btn := range buttons {
		interactiveMsg.NativeFlowMessage.Buttons[i] = btn.GetNativeFlowButton()
	}

	return c.SendMessage(&waE2E.Message{
		InteractiveMessage: &waE2E.InteractiveMessage{
			InteractiveMessage: interactiveMsg,
		},
	})
}
