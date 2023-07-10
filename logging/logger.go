package logging

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger() (*zap.Logger, error) {
	// Log level parsing
	logLevel, err := zap.ParseAtomicLevel(viper.GetString("log_level"))
	if err != nil {
		return nil, err
	}

	zapCFG := zap.Config{
		Encoding:         "console",
		Level:            logLevel,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message",

			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalColorLevelEncoder,

			TimeKey:    "time",
			EncodeTime: zapcore.ISO8601TimeEncoder,

			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}

	logger, err := zapCFG.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}
