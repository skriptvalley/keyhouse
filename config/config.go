package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Config holds application-wide configuration
type Config struct {
	LogLevel       string
	AppVersion     string
	ShutdownGrace  time.Duration
	HTTPPort       int
	GRPCPort       int
	SwaggerEnabled bool
	SwaggerPort    int
	SwaggerDir     string
	// Store
	StoreType    string
	StoreCfgPath string
	// Redis
	RedisHost     string
	RedisPort     string
	RedisPassword string
}

func SetConfigInEnvs(cfg *Config) {
	envFilePath := "/app/.env"
	file, err := os.Create(envFilePath)
	if err != nil {
		log.Fatalf("failed to create env file: %v", err)
	}
	defer file.Close()

	writeEnv(file, "LOG_LEVEL", cfg.LogLevel)
	writeEnv(file, "APP_VERSION", cfg.AppVersion)
	writeEnv(file, "HTTP_PORT", fmt.Sprintf("%d", cfg.HTTPPort))
	writeEnv(file, "GRPC_PORT", fmt.Sprintf("%d", cfg.GRPCPort))
	writeEnv(file, "SWAGGER_ENABLED", fmt.Sprintf("%t", cfg.SwaggerEnabled))
	writeEnv(file, "SWAGGER_PORT", fmt.Sprintf("%d", cfg.SwaggerPort))
	writeEnv(file, "SWAGGER_DIR", cfg.SwaggerDir)
	writeEnv(file, "STORE_TYPE", cfg.StoreType)
	writeEnv(file, "STORE_CFG_PATH", cfg.StoreCfgPath)
	writeEnv(file, "REDIS_HOST", cfg.RedisHost)
	writeEnv(file, "REDIS_PORT", cfg.RedisPort)
	writeEnv(file, "REDIS_PASSWORD", cfg.RedisPassword)

	appendSourceCommandToRC(envFilePath)
}

func writeEnv(file *os.File, key, value string) {
	fmt.Fprintf(file, "export %s=%s\n", key, value)
	os.Setenv(key, value)
}

func appendSourceCommandToRC(envFilePath string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to determine home directory: %v", err)
	}
	bashrcPath := filepath.Join(homeDir, ".bashrc")
	profilePath := filepath.Join(homeDir, ".profile")

	sourceCommand := fmt.Sprintf("\n# Source the application's .env file\nsource %s\n", envFilePath)

	for _, rcFile := range []string{bashrcPath, profilePath} {
		fullPath := os.ExpandEnv(rcFile) // Expand the tilde (~) to the user's home directory
		file, err := os.OpenFile(fullPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			log.Printf("failed to open %s: %v", fullPath, err)
			continue
		}
		defer file.Close()

		// Append the source command if not already present
		fmt.Fprintln(file, sourceCommand)
	}
}

// LoadConfig initializes configuration using environment variables and flags
func LoadConfig() *Config {
	cfg := &Config{}

	// Command-line flags
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "log level (debug, info, warn, error)")
	flag.StringVar(&cfg.AppVersion, "app-version", "0.0.0", "application version")
	flag.DurationVar(&cfg.ShutdownGrace, "shutdown-grace", 10*time.Second, "graceful shutdown timeout")
	flag.IntVar(&cfg.HTTPPort, "http-port", 8080, "application port")
	flag.IntVar(&cfg.GRPCPort, "grpc-port", 9090, "gRPC server port")
	flag.BoolVar(&cfg.SwaggerEnabled, "swagger-enabled", false, "enable Swagger UI")
	flag.IntVar(&cfg.SwaggerPort, "swagger-port", 8081, "Swagger UI port")
	flag.StringVar(&cfg.SwaggerDir, "swagger-dir", "./docs", "Swagger docs directory")
	// Store configuration
	flag.StringVar(&cfg.StoreType, "store-type", "postgres", "database type (postgres, mysql, etc.)")
	flag.StringVar(&cfg.StoreCfgPath, "store-cfg-path", "", "store configuration file path")
	// Redis configuration
	flag.StringVar(&cfg.RedisHost, "redis-host", "localhost", "Redis host")
	flag.StringVar(&cfg.RedisPort, "redis-port", "6379", "Redis port")
	flag.StringVar(&cfg.RedisPassword, "redis-password", "admin", "Redis password")

	// Parse command-line flags
	flag.Parse()

	// Optional: Override with environment variables
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		cfg.LogLevel = envLogLevel
	}
	if envAppVersion := os.Getenv("APP_VERSION"); envAppVersion != "" {
		cfg.AppVersion = envAppVersion
	}

	SetConfigInEnvs(cfg)

	if err := cfg.Validate(); err != nil {
		log.Fatalf("config validation error: %v", err)
	}

	return cfg
}

func (cfg *Config) Validate() error {
	// Perform validation checks on the configuration

	return nil
}
