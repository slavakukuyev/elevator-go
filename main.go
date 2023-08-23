package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func initLogger() {
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

	var err error
	logger, err = config.Build()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer logger.Sync() // Flushes buffer, if any
}

func main() {
	if err := env.Parse(&cfg); err != nil {
		panic("error on parsing env")
	}

	initLogger()

	manager := NewManager()

	elevator1 := NewElevator("A", cfg.MaxFloor, cfg.MinFloor)
	elevator2 := NewElevator("B", cfg.MaxFloor, cfg.MinFloor)

	manager.AddElevator(elevator1)
	manager.AddElevator(elevator2)

	port := cfg.Port
	server := NewServer(port, manager)

	// Start the server in a separate goroutine
	go server.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	server.Shutdown()
}
