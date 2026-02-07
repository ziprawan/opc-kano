package config

import (
	"kano/internal/logger"
	"kano/internal/utils/parser"
	"time"
)

const logLevel = logger.LogLevelDebug

var log *logger.Logger
var parserObj *parser.Parser
var configObj *Config
var Jakarta *time.Location

func Init() {
	if log == nil {
		log = logger.Init("Kano", logLevel)
	}
	if parserObj == nil {
		parserObj = parser.Init([]string{"/", "!", ".", "/ ", "! ", ". "})
	}
	if configObj == nil {
		configObj = InitConfig()
	}

	var err error
	Jakarta, err = time.LoadLocation("Asia/Jakarta")
	if err != nil {
		panic(err)
	}
}

func GetLogger() *logger.Logger {
	if log == nil {
		Init()
	}
	return log
}

func GetParser() *parser.Parser {
	if parserObj == nil {
		Init()
	}
	return parserObj
}

func GetConfig() *Config {
	if configObj == nil {
		Init()
	}
	return configObj
}
