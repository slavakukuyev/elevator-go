package main

import (
	"os"
	"os/signal"
	"syscall"
)

func main() {
	initConfig()
	logger := NewLogger()

	manager := NewManager(logger)

	elevator1 := NewElevator("A", cfg.MinFloor, cfg.MaxFloor, cfg.EachFloorDuration, cfg.OpenDoorDuration, logger)
	elevator2 := NewElevator("B", cfg.MinFloor, cfg.MaxFloor, cfg.EachFloorDuration, cfg.OpenDoorDuration, logger)
	elevator3 := NewElevator("C", -4, 5, cfg.EachFloorDuration, cfg.OpenDoorDuration, logger)

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
