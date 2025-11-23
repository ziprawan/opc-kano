package logger

import (
	"fmt"
	"os"
	"time"
)

type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

type Logger struct {
	file     *os.File
	name     string
	minLevel LogLevel
}

var levelToString = map[LogLevel]string{
	LogLevelDebug: "DEBUG",
	LogLevelError: "ERROR",
	LogLevelInfo:  "INFO ",
	LogLevelWarn:  "WARN ",
}

func (l *Logger) outputf(level LogLevel, msg string, args ...any) {
	if level < l.minLevel {
		return
	}
	str, ok := levelToString[level]
	if !ok {
		str = "UKWN"
	}
	logStr := fmt.Sprintf("%s [%s %s] %s\n", time.Now().Format("15:04:05.000"), str, l.name, fmt.Sprintf(msg, args...))
	if l.file == nil {
		fmt.Print(logStr)
	} else {
		l.file.WriteString(logStr)
	}
}

func (l *Logger) Debugf(msg string, args ...any) { l.outputf(LogLevelDebug, msg, args...) }
func (l *Logger) Infof(msg string, args ...any)  { l.outputf(LogLevelInfo, msg, args...) }
func (l *Logger) Warnf(msg string, args ...any)  { l.outputf(LogLevelWarn, msg, args...) }
func (l *Logger) Errorf(msg string, args ...any) { l.outputf(LogLevelError, msg, args...) }
func (l *Logger) Sub(name string) *Logger {
	return &Logger{file: l.file, name: fmt.Sprintf("%s/%s", l.name, name), minLevel: l.minLevel}
}
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

func Init(name string, minLevel LogLevel) *Logger {
	filename := fmt.Sprintf("logs/%d.log", time.Now().UnixMilli())
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("unable to init log file, using stdout instead")
		file = nil
	}

	return &Logger{file: file, name: name, minLevel: minLevel}
}
