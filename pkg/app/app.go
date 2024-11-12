package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/skriptvalley/keyhouse/config"
	"github.com/skriptvalley/keyhouse/pkg/server"
	"github.com/skriptvalley/keyhouse/pkg/statemanager"

	"go.uber.org/zap"
)

// App represents the core application
type App struct {
	logger *zap.Logger
	config *config.Config
	server *server.Server
	sm     *statemanager.StateManager
}

// NewApp initializes a new App instance with routing and middleware
func NewApp(cfg *config.Config, logger *zap.Logger) *App {
	ctx := context.Background()
	var err error

	statemgr := statemanager.NewStateManager(logger, cfg.RedisHost, cfg.RedisPort, cfg.RedisPassword)
	err = statemgr.Ping(ctx)
	if err != nil {
		logger.Fatal("Failed to connect to redis", zap.String("method", "NewApp"), zap.Error(err))
	}
	err = statemgr.InitStateDBCache(ctx)
	if err != nil {
		logger.Fatal("Failed to initialize state db cache", zap.String("method", "NewApp"), zap.Error(err))
	}

	app := &App{
		logger: logger,
		config: cfg,
		server: server.NewServer(logger, cfg, statemgr),
		sm:     statemgr,
	}
	return app
}

// Run starts the application with graceful shutdown
func (a *App) Run() {
	a.VaultStateChecks()
	// API Server
	a.server.Start()

	a.gracefulShutdown()
}

func (a *App) gracefulShutdown() {
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

func (a *App) VaultStateChecks() {
	ctx := context.Background()
	down := a.sm.IsVaultDown(ctx)
	if down {
		code, err := a.sm.GetInitCode(ctx)
		if err != nil {
			a.logger.Fatal("Failed to initialize vault", zap.String("method", "VaultStateChecks"), zap.Error(err))
		}
		a.logger.Info("==> Vault not initialized, please initialize vault with the following code",
			zap.String("code", code),
			zap.String("method", "NewApp"))
	}
	ready := a.sm.IsVaultReady(ctx)
	if !ready {
		a.logger.Info("==> Vault is not ready, please unlock the vault with minimum of 3 activation keys",
			zap.String("method", "VaultStateChecks"))
	} else {
		a.logger.Info("Vault is ready", zap.String("method", "VaultStateChecks"))
	}
}
