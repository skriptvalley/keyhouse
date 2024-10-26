package logger

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger initializes a zap logger with the specified log level
func NewLogger(level string) *zap.Logger {
	var zapConfig zap.Config
	if level == "debug" {
		zapConfig = zap.NewDevelopmentConfig()
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	zapLevel := zap.NewAtomicLevel()
	switch level {
	case "debug":
		zapLevel.SetLevel(zapcore.DebugLevel)
	case "info":
		zapLevel.SetLevel(zapcore.InfoLevel)
	case "warn":
		zapLevel.SetLevel(zapcore.WarnLevel)
	case "error":
		zapLevel.SetLevel(zapcore.ErrorLevel)
	default:
		log.Fatalf("Invalid log level: %s", level)
	}
	zapConfig.Level = zapLevel
	zapConfig.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		MessageKey:     "message",
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
	
	logger, err := zapConfig.Build()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	return logger
}
