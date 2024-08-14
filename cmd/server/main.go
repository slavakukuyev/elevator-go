package main

import (
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/slavakukuyev/elevator-go/internal/app/config"
)

func main() {
	config.InitConfig()
	logger := NewLogger()

	factory := &StandardElevatorFactory{}
	manager := NewManager(factory, logger)

	err := manager.AddElevator("A", cfg.MinFloor, cfg.MaxFloor, cfg.EachFloorDuration, cfg.OpenDoorDuration, logger)
	if err != nil {
		logger.Fatal("elevator %s not created", zap.String("name", "A"))
	}
	err = manager.AddElevator("B", cfg.MinFloor, cfg.MaxFloor, cfg.EachFloorDuration, cfg.OpenDoorDuration, logger)
	if err != nil {
		logger.Fatal("elevator %s not created", zap.String("name", "B"))
	}
	err = manager.AddElevator("C", -4, 5, cfg.EachFloorDuration, cfg.OpenDoorDuration, logger)
	if err != nil {
		logger.Fatal("elevator %s not created", zap.String("name", "C"))
	}

	port := cfg.Port
	server := NewServer(port, manager, logger)

	// Start the server in a separate goroutine
	go server.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	server.Shutdown()
}
