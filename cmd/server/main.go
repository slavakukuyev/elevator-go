package main

import (
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/slavakukuyev/elevator-go/internal/elevator"
	"github.com/slavakukuyev/elevator-go/internal/http"
	"github.com/slavakukuyev/elevator-go/internal/infra/config"
	"github.com/slavakukuyev/elevator-go/internal/infra/logger"
	"github.com/slavakukuyev/elevator-go/internal/manager"
)

func main() {
	cfg := config.InitConfig()
	logger := logger.NewLogger()

	factory := &elevator.StandardElevatorFactory{}
	manager := manager.NewManager(cfg, factory, logger)

	err := manager.AddElevator(cfg, "A", cfg.MinFloor, cfg.MaxFloor, cfg.EachFloorDuration, cfg.OpenDoorDuration, logger)
	if err != nil {
		logger.Fatal("elevator %s not created", zap.String("name", "A"))
	}
	err = manager.AddElevator(cfg, "B", cfg.MinFloor, cfg.MaxFloor, cfg.EachFloorDuration, cfg.OpenDoorDuration, logger)
	if err != nil {
		logger.Fatal("elevator %s not created", zap.String("name", "B"))
	}
	err = manager.AddElevator(cfg, "C", -4, 5, cfg.EachFloorDuration, cfg.OpenDoorDuration, logger)
	if err != nil {
		logger.Fatal("elevator %s not created", zap.String("name", "C"))
	}

	port := cfg.Port
	server := http.NewServer(cfg, port, manager, logger)

	// Start the server in a separate goroutine
	go server.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	server.Shutdown()
}
