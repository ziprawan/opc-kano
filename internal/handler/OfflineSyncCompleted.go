package handler

import (
	"kano/internal/config"

	"go.mau.fi/whatsmeow/types/events"
)

func OfflineSyncCompleted(evt *events.OfflineSyncCompleted) error {
	config.GetLogger().Sub("OfflineSyncCompleted").Infof("Synced %d message(s)", evt.Count)

	return nil
}
