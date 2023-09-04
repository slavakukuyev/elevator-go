package main

import (
	"os"
	"os/signal"
	"syscall"

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
	initConfig()
	initLogger()

	manager := NewManager()

	elevator1 := NewElevator("A", cfg.MinFloor, cfg.MaxFloor, logger)
	elevator2 := NewElevator("B", cfg.MinFloor, cfg.MaxFloor, logger)
	elevator3 := NewElevator("C", -4, 5, logger)

	manager.AddElevator(elevator1)
	manager.AddElevator(elevator2)
	manager.AddElevator(elevator3)

	port := cfg.Port
	server := NewServer(port, manager)

	// Start the server in a separate goroutine
	go server.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	server.Shutdown()
}
