package config

import (
	"flag"
	"os"
	"time"
)

// Config holds application-wide configuration
type Config struct {
	LogLevel       string
	AppVersion     string
	HTTPPort       int
	GRPCPort       int
	SwaggerEnabled bool
	SwaggerPort    int
	SwaggerDir     string
	ShutdownGrace  time.Duration
}

// LoadConfig initializes configuration using environment variables and flags
func LoadConfig() *Config {
	cfg := &Config{}

	// Command-line flags
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "log level (debug, info, warn, error)")
	flag.StringVar(&cfg.AppVersion, "app-version", "0.0.0", "application version")
	flag.IntVar(&cfg.HTTPPort, "http-port", 8080, "application port")
	flag.IntVar(&cfg.GRPCPort, "grpc-port", 9090, "gRPC server port")
	flag.BoolVar(&cfg.SwaggerEnabled, "swagger-enabled", false, "enable Swagger UI")
	flag.IntVar(&cfg.SwaggerPort, "swagger-port", 8081, "Swagger UI port")
	flag.StringVar(&cfg.SwaggerDir, "swagger-dir", "./docs", "Swagger docs directory")
	flag.DurationVar(&cfg.ShutdownGrace, "shutdown-grace", 10*time.Second, "graceful shutdown timeout")

	// Parse command-line flags
	flag.Parse()

	// Optional: Override with environment variables
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		cfg.LogLevel = envLogLevel
	}
	if envAppVersion := os.Getenv("APP_VERSION"); envAppVersion != "" {
		cfg.AppVersion = envAppVersion
	}

	return cfg
}
