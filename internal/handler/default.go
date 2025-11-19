package handler

import "kano/internal/config"

func defaultHandler(ev any) error {
	log := config.GetLogger().Sub("LogEvent")
	log.Warnf("Unhandled event %T", ev)
	return nil
}
