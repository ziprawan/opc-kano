package handler

import (
	"kano/internal/config"

	"go.mau.fi/whatsmeow/types/events"
)

func PushName(ev *events.PushName) error {
	log := config.GetLogger().Sub("PushName")
	log.Debugf("%s changed push name from %q to %q", ev.JID.String(), ev.OldPushName, ev.NewPushName)

	return nil
}
