package main

import (
	"os"
	"os/signal"
	"syscall"
)

func main() {
	initConfig()
	logger := NewLogger()

	factory := &StandardElevatorFactory{}
	manager := NewManager(factory, logger)

	manager.AddElevator("A", cfg.MinFloor, cfg.MaxFloor, cfg.EachFloorDuration, cfg.OpenDoorDuration, logger)
	manager.AddElevator("B", cfg.MinFloor, cfg.MaxFloor, cfg.EachFloorDuration, cfg.OpenDoorDuration, logger)
	manager.AddElevator("C", -4, 5, cfg.EachFloorDuration, cfg.OpenDoorDuration, logger)

	port := cfg.Port
	server := NewServer(port, manager, logger)

	// Start the server in a separate goroutine
	go server.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	server.Shutdown()
}
