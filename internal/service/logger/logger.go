package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	AccessLogger *zap.Logger
	DBLogger     *zap.Logger
)

func InitLoggers() error {
	accessConfig := zap.NewProductionConfig()
	accessConfig.OutputPaths = []string{
		"access.log",
	}
	accessConfig.EncoderConfig.TimeKey = "timestamp"
	accessConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var err error
	AccessLogger, err = accessConfig.Build()
	if err != nil {
		return err
	}

	dbConfig := zap.NewProductionConfig()
	dbConfig.OutputPaths = []string{
		"db.log",
	}
	dbConfig.EncoderConfig.TimeKey = "timestamp"
	dbConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	DBLogger, err = dbConfig.Build()
	if err != nil {
		return err
	}

	return nil
}

func SyncLoggers() error {
	err := AccessLogger.Sync()
	if err != nil {
		return err
	}
	err = DBLogger.Sync()
	if err != nil {
		return err
	}
	return nil
}
