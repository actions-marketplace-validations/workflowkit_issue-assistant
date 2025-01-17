package logger

import (
	"log"
)

var Log Logger

type LoggerType string

const (
	// BasicLogger is the basic logger
	BasicLogger LoggerType = "basic"

	// ZapLogger is the zap logger
	ZapLogger LoggerType = "zap"
)

type Logger interface {
	Error(msg string)
	Errorf(msg string, args ...interface{})

	Warn(msg string)
	Warnf(msg string, args ...interface{})

	Debug(msg string)
	Debugf(msg string, args ...interface{})

	Info(msg string)
	Infof(msg string, args ...interface{})

	Fatal(msg string)
	Fatalf(msg string, args ...interface{})
}

func init() {
	Log = &basicLogger{logger: log.New(log.Writer(), log.Prefix(), log.Flags())}
}

// SetLogger is initialize log library
func SetLogger(loggerInstance LoggerType) {
	if loggerInstance == ZapLogger {
		Log = setupZapLogger()
	} else if loggerInstance == BasicLogger {
		Log = setupBasicLogger()
	} else {
		Log = setupBasicLogger()
	}
}
