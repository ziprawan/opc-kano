package handler

import (
	"context"
	"kano/internal/config"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
)

var connectedCount = 0

func Connected(cli *whatsmeow.Client) error {
	connectedCount++

	logger := config.GetLogger().Sub("Connected")

	// logger.Debugf("Setting \"Force Active Delivery Receipts\" to true")
	// cli.SetForceActiveDeliveryReceipts(true)
	logger.Debugf("Marking client presence as online")
	err := cli.SendPresence(context.Background(), types.PresenceAvailable)
	if err != nil {
		logger.Warnf("Failed to set client presence as online: %s", err.Error())
	}

	logger.Infof("Connected #%d", connectedCount)

	return nil
}
