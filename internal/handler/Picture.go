package handler

import (
	"kano/internal/config"

	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func Picture(ev *events.Picture) (err error) {
	log := config.GetLogger().Sub("Picture")

	if ev.JID.Server != types.DefaultUserServer && ev.JID.Server != types.HiddenUserServer && ev.JID.Server != types.MessengerServer {
		if ev.Remove {
			log.Debugf("%s removed group %s profile", ev.Author.String(), ev.JID.String())
		} else {
			log.Debugf("%s changed group %s profile (ID: %s)", ev.Author.String(), ev.JID.String(), ev.PictureID)
		}
	}

	if ev.Remove {
		log.Debugf("%s removed their profile", ev.JID.String())
	} else {
		log.Debugf("%s changed their profile (ID: %s)", ev.JID.String(), ev.PictureID)
	}

	return
}
