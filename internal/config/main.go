package config

import (
	"errors"
	"kano/internal/logger"
	"kano/internal/utils/parser"
)

const logLevel = logger.LogLevelDebug

var log *logger.Logger
var parserObj *parser.Parser
var configObj *Config

func Init() {
	log = logger.Init("Kano", logLevel)
	parserObj = parser.Init([]string{"/", "!", ".", "/ ", "! ", ". "})
	configObj = InitConfig()
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
		panic(errors.New("config is not initialized! Call LoadConfig() first"))
	}
	return configObj
}
