package handler

import (
	"kano/internal/config"
	"strings"

	"go.mau.fi/whatsmeow/types/events"
)

func Receipt(ev *events.Receipt) (err error) {
	log := config.GetLogger().Sub("Receipt")
	log.Debugf("%s \"%s\" %s:%s at %s", ev.Sender, ev.Type, ev.Chat.String(), strings.Join(ev.MessageIDs, ","), ev.Timestamp.String())

	return
}
