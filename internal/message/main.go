package message

import (
	"kano/internal/message/handles"
	"kano/internal/utils/messageutil"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

func Main(cli *whatsmeow.Client, evt *events.Message) error {
	msgCtx := messageutil.CreateContext(cli, evt)

	if msgCtx.GetText() != "" {
		return handles.Handle(msgCtx)
	}

	return nil
}
