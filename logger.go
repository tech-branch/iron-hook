package ironhook

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newLoggerAtLevel(levelStr string) (*zap.Logger, error) {

	var logLevel zapcore.Level
	var err error

	if levelStr != "" {
		logLevel, err = zapcore.ParseLevel(levelStr)
		if err != nil {
			logLevel = zapcore.InfoLevel
		}
	} else {
		logLevel = zapcore.InfoLevel
	}

	logConf := zap.NewProductionConfig()
	logConf.Level = zap.NewAtomicLevelAt(logLevel)

	logger, err := logConf.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}
