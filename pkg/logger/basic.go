package logger

import (
	"fmt"
	"log"
	"sync"
)

type basicLogger struct {
	logger *log.Logger
}

func (l *basicLogger) Error(msg string) {
	l.logger.Println("❌ " + msg)
}

func (l *basicLogger) Errorf(msg string, args ...interface{}) {
	message := fmt.Sprintf(msg, args...)
	l.logger.Println("❌ " + message)
}

func (l *basicLogger) Warn(msg string) {
	l.logger.Println("⚠️ " + msg)
}

func (l *basicLogger) Warnf(msg string, args ...interface{}) {
	message := fmt.Sprintf(msg, args...)
	l.logger.Println("⚠️ " + message)
}

func (l *basicLogger) Debug(msg string) {
	l.logger.Println("🔍 " + msg)
}

func (l *basicLogger) Debugf(msg string, args ...interface{}) {
	message := fmt.Sprintf(msg, args...)
	l.logger.Println("🔍 " + message)
}

func (l *basicLogger) Info(msg string) {
	l.logger.Println("🔵 " + msg)
}

func (l *basicLogger) Infof(msg string, args ...interface{}) {
	message := fmt.Sprintf(msg, args...)
	l.logger.Println("🔵 " + message)
}

func (l *basicLogger) Fatal(msg string) {
	l.logger.Fatal("☠️ " + msg)
}

func (l *basicLogger) Fatalf(msg string, args ...interface{}) {
	message := fmt.Sprintf(msg, args...)
	l.logger.Fatal("☠️ " + message)
}

var (
	onceBasic           sync.Once
	basicLoggerInstance *basicLogger
)

func setupBasicLogger() *basicLogger {
	if basicLoggerInstance == nil {
		onceBasic.Do(func() {
			basicLoggerInstance = &basicLogger{logger: log.New(log.Writer(), log.Prefix(), log.Flags())}
		})
	}
	return basicLoggerInstance
}
