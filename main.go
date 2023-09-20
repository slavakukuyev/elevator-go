package main

import (
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	initConfig()
	logger := NewLogger()

	manager := NewManager(logger)

	//default elevators
	//main
	elevator1, err := NewElevator("A", cfg.MinFloor, cfg.MaxFloor, cfg.EachFloorDuration, cfg.OpenDoorDuration, logger)
	if err != nil {
		logger.Fatal("elevator %s not created", zap.String("name", "A"))
	}
	elevator2, err := NewElevator("B", cfg.MinFloor, cfg.MaxFloor, cfg.EachFloorDuration, cfg.OpenDoorDuration, logger)
	if err != nil {
		logger.Fatal("elevator %s not created", zap.String("name", "B"))
	}
	//parking
	elevator3, err := NewElevator("C", -4, 5, cfg.EachFloorDuration, cfg.OpenDoorDuration, logger)
	if err != nil {
		logger.Fatal("elevator %s not created", zap.String("name", "C"))
	}

	manager.AddElevator(elevator1)
	manager.AddElevator(elevator2)
	manager.AddElevator(elevator3)

	port := cfg.Port
	server := NewServer(port, manager, logger)

	// Start the server in a separate goroutine
	go server.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	server.Shutdown()
}
