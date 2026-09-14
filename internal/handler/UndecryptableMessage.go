package handler

import (
	"context"
	"kano/internal/config"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func UndecryptableMessage(cli *whatsmeow.Client, ev *events.UndecryptableMessage) (err error) {
	log := config.GetLogger().Sub("UndecryptableMessage")

	if !ev.IsUnavailable {
		log.Debugf("Failed to decrypt message %s:%s: %s", ev.Info.Chat.String(), ev.Info.ID, ev.DecryptFailMode)
		return
	}

	switch ev.UnavailableType {
	case events.UnavailableTypeViewOnce:
		log.Debugf("Message %s:%s is a view once message, marking as read", ev.Info.Chat.String(), ev.Info.ID)
		err := cli.MarkRead(context.Background(), []types.MessageID{ev.Info.ID}, time.Now(), ev.Info.Chat, ev.Info.Sender)
		if err != nil {
			log.Warnf("Failed to mark message %s:%s as read: %s", ev.Info.Chat.String(), ev.Info.ID, err.Error())
		}
	case events.UnavailableTypeUnknown:
		log.Debugf("Message %s:%s is intended to be unavailable for no reason", ev.Info.Chat.String(), ev.Info.ID)
	}

	return
}
