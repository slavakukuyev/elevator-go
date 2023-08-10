package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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

	var wg sync.WaitGroup
	wg.Add(1)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Start your main logic in a goroutine
	go func() {
		defer wg.Done()
		for {
			select {
			case <-signals:
				logger.Info("received termination signal.")
				return // Exit the loop when a termination signal is received
			default:
				time.Sleep(time.Second * 5)
			}
		}
	}()

	initLogger()

	manager := NewManager()

	elevator1 := NewElevator("A")
	elevator2 := NewElevator("B")

	manager.AddElevator(elevator1)
	manager.AddElevator(elevator2)

	// Request an elevator going from floor 1 to floor 9
	if err := manager.RequestElevator(1, 9); err != nil {
		logger.Error("request elevator 1,9 error", zap.Error(err))
	}

	// Request an elevator going from floor 3 to floor 5
	if err := manager.RequestElevator(3, 5); err != nil {
		logger.Error("request elevator 3,5 error", zap.Error(err))
	}

	// Request an elevator going from floor 3 to floor 5
	if err := manager.RequestElevator(6, 4); err != nil {
		logger.Error("request elevator 6,4 error", zap.Error(err))
	}

	time.Sleep(time.Second * 7)

	if err := manager.RequestElevator(1, 2); err != nil {
		logger.Error("request elevator 1,2 error", zap.Error(err))
	}

	time.Sleep(time.Second * 10)

	if err := manager.RequestElevator(7, 0); err != nil {
		logger.Error("request elevator 7,0 error", zap.Error(err))
	}

	wg.Wait() // Wait until the termination
}
