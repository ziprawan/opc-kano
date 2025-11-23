package handler

import (
	"kano/internal/config"

	"go.mau.fi/whatsmeow/types/events"
)

func OfflineSyncPreview(evt *events.OfflineSyncPreview) error {
	logger := config.GetLogger().Sub("OfflineSyncPreview")

	logger.Debugf("Sync preview (total is %d):", evt.Total)
	logger.Debugf("App data changes: %d", evt.AppDataChanges)
	logger.Debugf("Messages: %d", evt.Messages)
	logger.Debugf("Notifications: %d", evt.Notifications)
	logger.Debugf("Receipts: %d", evt.Receipts)

	return nil
}
