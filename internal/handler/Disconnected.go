package handler

import (
	"kano/internal/config"
)

func Disconnected() error {

	logger := config.GetLogger().Sub("Disconnected")
	logger.Warnf("Disconnected from server, reconnecting")

	return nil
}
