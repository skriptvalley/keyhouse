package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/skriptvalley/keyhouse/config"
	"github.com/skriptvalley/keyhouse/pkg/server"

	"go.uber.org/zap"
)

// App represents the core application
type App struct {
	logger *zap.Logger
	config *config.Config
	server *server.Server
}

// NewApp initializes a new App instance with routing and middleware
func NewApp(cfg *config.Config, logger *zap.Logger) *App {
	app := &App{
		logger: logger,
		config: cfg,
		server: server.NewServer(cfg, logger),
	}

	return app
}

// Run starts the application with graceful shutdown
func (a *App) Run() {
	a.server.Start()
	// Graceful shutdown on system signals
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	<-shutdownChan
	a.logger.Info("Shutting down the server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	a.server.Shutdown(ctx)
	a.logger.Info("Server gracefully stopped.")
}
