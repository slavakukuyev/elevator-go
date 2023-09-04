package main

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger() *zap.Logger {
	config := zap.Config{
		Encoding:    "console", // or "json"
		Level:       zap.NewAtomicLevelAt(zap.DebugLevel),
		OutputPaths: []string{"stdout"},
		EncoderConfig: zapcore.EncoderConfig{
			LevelKey:    "level",
			TimeKey:     "time",
			MessageKey:  "message",
			EncodeLevel: zapcore.LowercaseColorLevelEncoder,
			EncodeTime:  zapcore.ISO8601TimeEncoder,
		},
	}

	log, err := config.Build()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer func() {
		err := log.Sync() // Flushes buffer, if any
		if err != nil {
			fmt.Println(err)
		}
	}()
	return log
}
