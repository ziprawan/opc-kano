package handler

import (
	"kano/internal/config"

	"go.mau.fi/whatsmeow/types/events"
)

func KeepAliveTimeout(ev *events.KeepAliveTimeout) (err error) {
	log := config.GetLogger().Sub("KeepAliveTimeout")
	log.Debugf("Timeout")

	return
}
