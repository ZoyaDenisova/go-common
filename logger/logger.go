package logger

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Interface -.
type Interface interface {
	Debug(msg interface{}, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg interface{}, args ...interface{})
	Fatal(msg interface{}, args ...interface{})
}

type Logger struct {
	logger *zap.Logger
}

var _ Interface = (*Logger)(nil)

func New(level string) *Logger {
	var zapLevel zapcore.Level
	switch strings.ToLower(level) {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	case "fatal":
		zapLevel = zapcore.FatalLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	encoder := zapcore.NewJSONEncoder(encoderCfg)

	logFile, err := os.OpenFile("logs/app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Sprintf("can't open log file: %v", err))
	}
	fileWriter := zapcore.AddSync(logFile)
	consoleWriter := zapcore.AddSync(os.Stdout)

	core := zapcore.NewTee(
		zapcore.NewCore(encoder, consoleWriter, zapLevel),
		zapcore.NewCore(encoder, fileWriter, zapLevel),
	)

	return &Logger{
		logger: zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)),
	}
}

func (l *Logger) Debug(message interface{}, args ...interface{}) {
	msg := stringify(message)
	l.logger.Debug(msg, toZapFields(args...)...)
}

func (l *Logger) Info(message string, args ...interface{}) {
	l.logger.Info(message, toZapFields(args...)...)
}

func (l *Logger) Warn(message string, args ...interface{}) {
	l.logger.Warn(message, toZapFields(args...)...)
}

func (l *Logger) Error(message interface{}, args ...interface{}) {
	msg := stringify(message)
	l.logger.Error(msg, toZapFields(args...)...)
}

func (l *Logger) Fatal(message interface{}, args ...interface{}) {
	msg := stringify(message)
	l.logger.Fatal(msg, toZapFields(args...)...)
}

func stringify(v interface{}) string {
	switch m := v.(type) {
	case error:
		return m.Error()
	case string:
		return m
	default:
		return fmt.Sprintf("%v", m)
	}
}

func toZapFields(args ...interface{}) []zap.Field {
	fields := make([]zap.Field, 0, len(args)/2)
	for i := 0; i+1 < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			key = fmt.Sprintf("invalid_key_%d", i)
		}
		fields = append(fields, zap.Any(key, args[i+1]))
	}
	return fields
}
