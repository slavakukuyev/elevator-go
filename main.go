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

const _directionUp = "up"
const _directionDown = "down"

func main() {
	initLogger()

	manager := NewManager()

	elevator1 := NewElevator("A")
	elevator2 := NewElevator("B")

	manager.AddElevator(elevator1)
	manager.AddElevator(elevator2)

	port := 8080
	server := NewServer(port, manager)

	// Start the server in a separate goroutine
	go server.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	server.Shutdown()
}
