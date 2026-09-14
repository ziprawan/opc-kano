package handler

import (
	"encoding/json"
	"fmt"
	"kano/internal/config"
	"os"
	"time"
)

func LogEvent(evt any) {
	logger := config.GetLogger().Sub("LogEvent")
	logger.Debugf("saving event: %T", evt)
	data, err := json.MarshalIndent(evt, "", "  ")
	if err != nil {
		logger.Warnf("failed to marshal event: %s", err)
		return
	}
	err = os.WriteFile(fmt.Sprintf("json/%d_%T.json", time.Now().UnixMilli(), evt), data, 0644)
	if err != nil {
		logger.Warnf("failed to write log event: %s", err)
		return
	}
}
