package logger

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapLogger struct {
	logger *zap.Logger
}

func (l *zapLogger) Error(msg string) {
	l.logger.Sugar().Error("❌ " + msg)
}

func (l *zapLogger) Errorf(msg string, args ...interface{}) {
	l.logger.Sugar().Errorf("❌ "+msg, args...)
}

func (l *zapLogger) Warn(msg string) {
	l.logger.Sugar().Warn("⚠️ " + msg)
}

func (l *zapLogger) Warnf(msg string, args ...interface{}) {
	l.logger.Sugar().Warnf("⚠️ "+msg, args...)
}

func (l *zapLogger) Debug(msg string) {
	l.logger.Sugar().Debug("🔍 " + msg)
}

func (l *zapLogger) Debugf(msg string, args ...interface{}) {
	l.logger.Sugar().Debugf("🔍 "+msg, args...)
}

func (l *zapLogger) Info(msg string) {
	l.logger.Sugar().Info("🔵 " + msg)
}

func (l *zapLogger) Infof(msg string, args ...interface{}) {
	l.logger.Sugar().Infof("🔵 "+msg, args...)
}

func (l *zapLogger) Fatal(msg string) {
	l.logger.Sugar().Fatal("☠️ " + msg)
}

func (l *zapLogger) Fatalf(msg string, args ...interface{}) {
	l.logger.Sugar().Fatalf("☠️"+msg, args...)
}

var (
	once              sync.Once
	zapLoggerInstance *zap.Logger
)

func setupZapLogger() *zapLogger {
	if zapLoggerInstance == nil {
		once.Do(func() {
			create()
		})
	}
	return &zapLogger{logger: zapLoggerInstance}
}

func create() {
	consoleWriteSyncer := zapcore.AddSync(os.Stdout)

	var encoder zapcore.Encoder
	var level zapcore.Level
	var encoderConfig zapcore.EncoderConfig

	var loggerOptions []zap.Option
	loggerOptions = append(loggerOptions, zap.AddCaller(), zap.AddCallerSkip(1))

	encoderConfig = zap.NewDevelopmentEncoderConfig()
	level = zap.DebugLevel
	loggerOptions = append(loggerOptions, zap.AddStacktrace(zapcore.ErrorLevel))

	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderConfig.LevelKey = "level"
	encoderConfig.TimeKey = "ts"
	encoderConfig.MessageKey = "msg"
	encoderConfig.CallerKey = "caller"
	encoderConfig.StacktraceKey = "stacktrace"
	encoderConfig.FunctionKey = ""
	encoder = zapcore.NewConsoleEncoder(encoderConfig)

	consoleCore := zapcore.NewCore(encoder, consoleWriteSyncer, level)
	zapCore := zapcore.NewTee(consoleCore)

	zapLoggerInstance = zap.New(zapCore, loggerOptions...)
}
