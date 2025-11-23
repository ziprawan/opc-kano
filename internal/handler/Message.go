package handler

import (
	"context"
	"kano/internal/config"
	"kano/internal/message"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func Message(cli *whatsmeow.Client, evt *events.Message) error {
	log := config.GetLogger().Sub("Message")

	// Mark as read
	// Somehow the RawAgent is included, no need to call ToNonAD()
	sender := evt.Info.Sender
	chat := evt.Info.Chat

	log.Debugf("Marking message %s as read at chat %s with sender %s", evt.Info.ID, chat.String(), sender.String())
	err := cli.MarkRead(context.Background(), []types.MessageID{evt.Info.ID}, time.Now(), chat, sender)
	if err != nil {
		log.Warnf("Failed to mark message %s as read at chat %s: %s", evt.Info.ID, chat.String(), err.Error())
	}

	return message.Main(cli, evt)
}
