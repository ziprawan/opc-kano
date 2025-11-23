package handler

import (
	"kano/internal/config"

	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func ChatPresence(ev *events.ChatPresence) error {
	log := config.GetLogger().Sub("ChatPresence")

	stateType := "[unknown state]"
	switch ev.State {
	case types.ChatPresencePaused:
		stateType = "paused"
	case types.ChatPresenceComposing:
		switch ev.Media {
		case types.ChatPresenceMediaText:
			stateType = "typing"
		case types.ChatPresenceMediaAudio:
			stateType = "recording audio"
		}
	}

	log.Debugf("%s: %s is %s", ev.Chat.String(), ev.Sender.String(), stateType)
	return nil
}
