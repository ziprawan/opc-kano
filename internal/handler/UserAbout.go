package handler

import (
	"kano/internal/config"

	"go.mau.fi/whatsmeow/types/events"
)

func UserAbout(ev *events.UserAbout) (err error) {
	log := config.GetLogger().Sub("UserAbout")
	log.Debugf("%s changed about into \"%s\" at %s", ev.JID.String(), ev.Status, ev.Timestamp.String())

	return
}
