package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/slavakukuyev/elevator-go/internal/factory"
	"github.com/slavakukuyev/elevator-go/internal/http"
	"github.com/slavakukuyev/elevator-go/internal/infra/config"
	"github.com/slavakukuyev/elevator-go/internal/manager"
)

func main() {
	cfg := config.InitConfig()

	factory := &factory.StandardElevatorFactory{}
	manager := manager.New(cfg, factory)

	err := manager.AddElevator(cfg, "A", cfg.MinFloor, cfg.MaxFloor, cfg.EachFloorDuration, cfg.OpenDoorDuration)
	if err != nil {
		slog.Error("elevator %s not created", slog.String("name", "A"))
		os.Exit(1)
	}
	err = manager.AddElevator(cfg, "B", cfg.MinFloor, cfg.MaxFloor, cfg.EachFloorDuration, cfg.OpenDoorDuration)
	if err != nil {
		slog.Error("elevator %s not created", slog.String("name", "B"))
		os.Exit(1)
	}
	err = manager.AddElevator(cfg, "C", -4, 5, cfg.EachFloorDuration, cfg.OpenDoorDuration)
	if err != nil {
		slog.Error("elevator %s not created", slog.String("name", "C"))
		os.Exit(1)
	}

	port := cfg.Port
	server := http.NewServer(cfg, port, manager)

	// Start the server in a separate goroutine
	go server.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	server.Shutdown()
}
