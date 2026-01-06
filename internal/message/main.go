package message

import (
	"fmt"
	"kano/internal/config"
	"kano/internal/message/handles"
	"kano/internal/utils/messageutil"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

func Main(cli *whatsmeow.Client, evt *events.Message) error {
	msgCtx := messageutil.CreateContext(cli, evt)
	if msgCtx == nil {
		return fmt.Errorf("msgCtx is nil, look for error logs related to messageutil.CreateContext")
	}

	if config.GetConfig().OwnerOnlyMode {
		if !msgCtx.IsSenderSame(config.GetConfig().OwnerJID) {
			return nil
		}
	}

	if msgCtx.GetText() != "" {
		return handles.Handle(msgCtx)
	}

	return nil
}
