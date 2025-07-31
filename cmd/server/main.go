package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/slavakukuyev/elevator-go/internal/factory"
	httpPkg "github.com/slavakukuyev/elevator-go/internal/http"
	"github.com/slavakukuyev/elevator-go/internal/infra/config"
	"github.com/slavakukuyev/elevator-go/internal/infra/logging"
	"github.com/slavakukuyev/elevator-go/internal/manager"
)

func main() {
	// Initialize configuration
	cfg, err := config.InitConfig()
	if err != nil {
		slog.Error("failed to initialize configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Initialize logging
	logging.InitLogger(cfg.LogLevel)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Log environment information for debugging
	envInfo := cfg.GetEnvironmentInfo()
	slog.InfoContext(ctx, "elevator system starting up",
		slog.String("environment", cfg.Environment),
		slog.String("log_level", cfg.LogLevel),
		slog.Int("port", cfg.Port),
		slog.Bool("metrics_enabled", cfg.MetricsEnabled),
		slog.Bool("websocket_enabled", cfg.WebSocketEnabled),
		slog.Bool("circuit_breaker_enabled", cfg.CircuitBreakerEnabled),
		slog.Any("config_summary", envInfo))

	// Initialize factory and manager
	elevatorFactory := &factory.StandardElevatorFactory{}
	elevatorManager := manager.New(cfg, elevatorFactory)

	// Create default elevators if configured
	if cfg.DefaultElevatorCount > 0 {
		slog.InfoContext(ctx, "creating default elevators",
			slog.Int("count", cfg.DefaultElevatorCount),
			slog.String("prefix", cfg.NamePrefix))

		for i := 0; i < cfg.DefaultElevatorCount; i++ {
			elevatorName := fmt.Sprintf("%s-%d", cfg.NamePrefix, i+1)
			err := elevatorManager.AddElevator(ctx, cfg, elevatorName,
				cfg.MinFloor, cfg.MaxFloor,
				cfg.EachFloorDuration, cfg.OpenDoorDuration, cfg.DefaultOverloadThreshold)
			if err != nil {
				slog.ErrorContext(ctx, "failed to create default elevator",
					slog.String("name", elevatorName),
					slog.String("error", err.Error()))
			} else {
				slog.InfoContext(ctx, "default elevator created",
					slog.String("name", elevatorName))
			}
		}
	}

	// Determine the port to use
	port := cfg.Port
	if port <= 0 {
		slog.WarnContext(ctx, "invalid port in configuration, using default",
			slog.Int("configured_port", port),
			slog.Int("default_port", 6660))
		port = 6660
	}

	// Create servers
	server := httpPkg.NewServer(cfg, port, elevatorManager)
	wsServer := httpPkg.NewWebSocketServer(6661, elevatorManager, slog.With(slog.String("component", "websocket-server")))

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// Start servers with proper error handling
	var httpStarted, wsStarted bool
	serverErrCh := make(chan error, 2)

	// Start main HTTP server
	go func() {
		slog.InfoContext(ctx, "starting HTTP server",
			slog.Int("port", port),
			slog.String("environment", cfg.Environment),
			slog.Duration("read_timeout", cfg.ReadTimeout),
			slog.Duration("write_timeout", cfg.WriteTimeout),
			slog.Duration("idle_timeout", cfg.IdleTimeout))

		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			slog.ErrorContext(ctx, "HTTP server failed to start",
				slog.Int("port", port),
				slog.String("error", err.Error()))
			serverErrCh <- fmt.Errorf("HTTP server failed: %w", err)
		}
	}()

	// Start WebSocket server
	go func() {
		slog.InfoContext(ctx, "starting WebSocket server",
			slog.Int("port", 6661))

		if err := wsServer.Start(); err != nil && err != http.ErrServerClosed {
			slog.ErrorContext(ctx, "WebSocket server failed to start",
				slog.Int("port", 6661),
				slog.String("error", err.Error()))
			serverErrCh <- fmt.Errorf("WebSocket server failed: %w", err)
		}
	}()

	// Wait a moment to see if servers start successfully
	startupTimer := time.NewTimer(2 * time.Second)
	httpStarted = true // Assume success unless we get an error quickly
	wsStarted = true

	select {
	case err := <-serverErrCh:
		// Server failed to start
		startupTimer.Stop()
		slog.ErrorContext(ctx, "server startup failed", slog.String("error", err.Error()))

		// Try to gracefully shutdown any servers that might have started
		shutdownServers(server, wsServer, cfg, httpStarted, wsStarted)
		elevatorManager.Shutdown()
		os.Exit(1)

	case <-startupTimer.C:
		// Servers started successfully
		slog.InfoContext(ctx, "all servers started successfully")

	case sig := <-quit:
		// Got shutdown signal during startup
		startupTimer.Stop()
		slog.InfoContext(ctx, "received shutdown signal during startup",
			slog.String("signal", sig.String()))
		shutdownServers(server, wsServer, cfg, httpStarted, wsStarted)
		elevatorManager.Shutdown()
		return
	}

	// Wait for shutdown signal
	sig := <-quit
	slog.InfoContext(ctx, "received shutdown signal",
		slog.String("signal", sig.String()),
		slog.Duration("shutdown_timeout", cfg.ShutdownTimeout))

	// Cancel context to signal all operations to stop
	cancel()

	// Shutdown servers gracefully
	shutdownServers(server, wsServer, cfg, httpStarted, wsStarted)

	// Shutdown the manager
	slog.InfoContext(ctx, "shutting down elevator manager")
	elevatorManager.Shutdown()
	slog.InfoContext(ctx, "elevator manager shutdown completed")

	// Wait for a short grace period before final exit
	select {
	case <-time.After(cfg.ShutdownGrace):
		slog.InfoContext(ctx, "graceful shutdown completed",
			slog.Duration("grace_period", cfg.ShutdownGrace))
	}
}

// shutdownServers gracefully shuts down both HTTP and WebSocket servers
func shutdownServers(server *httpPkg.Server, wsServer *httpPkg.WebSocketServer, cfg *config.Config, httpStarted, wsStarted bool) {
	slog.Info("shutting down servers gracefully")

	// Shutdown main HTTP server
	if httpStarted {
		if err := server.Shutdown(); err != nil {
			slog.Error("HTTP server shutdown failed", slog.String("error", err.Error()))
		} else {
			slog.Info("HTTP server shutdown completed")
		}
	}

	// Shutdown WebSocket server
	if wsStarted {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := wsServer.Shutdown(shutdownCtx); err != nil {
			slog.Error("WebSocket server shutdown failed", slog.String("error", err.Error()))
		} else {
			slog.Info("WebSocket server shutdown completed")
		}
	}
}
