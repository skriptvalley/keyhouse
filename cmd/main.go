package main

import (
	"github.com/skriptvalley/keyhouse/config"
	"github.com/skriptvalley/keyhouse/pkg/app"
	"github.com/skriptvalley/keyhouse/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	cfg := config.LoadConfig()

	log := logger.NewLogger(cfg.LogLevel)
	log.Info("Application starting", zap.Any("config", cfg))

	defer recoverMain()

	application := app.NewApp(cfg, log)
	application.Run()
}

func recoverMain() {
	if r := recover(); r != nil {
		logger.NewLogger("debug").Fatal("Application crashed with error", zap.Any("error", r))
	}
}
