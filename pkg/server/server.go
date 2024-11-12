package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"time"

	"github.com/skriptvalley/keyhouse/config"
	"github.com/skriptvalley/keyhouse/pkg/keystore"
	"github.com/skriptvalley/keyhouse/pkg/middleware"
	"github.com/skriptvalley/keyhouse/pkg/pb/app"
	"github.com/skriptvalley/keyhouse/pkg/statemanager"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const (
	PING_RETRIES = 5
	RETRY_AFTER  = 2
)

type Server struct {
	grpcServer *grpc.Server
	httpServer *http.Server
	config     *config.Config
	logger     *zap.Logger
}

func NewServer(logger *zap.Logger, cfg *config.Config, sm *statemanager.StateManager) *Server {
	ctx := context.Background()

	beStore, err := keystore.NewKeystore(cfg.StoreType, cfg.StoreCfgPath)
	if err != nil {
		logger.Fatal("Failed to initialize keystore", zap.String("method", "NewServer"), zap.Error(err))
	}
	for i := 0; i < PING_RETRIES; i++ {
		if err = beStore.Ping(); err == nil {
			logger.Info("Connected to keystore database", zap.String("method", "NewServer"))
			break
		}
		logger.Warn("Failed to connect to keystore database, retrying", zap.String("method", "NewServer"), zap.Error(err))
		time.Sleep(RETRY_AFTER * time.Second)
	}

	// Create Services
	appServer := &AppServer{
		appVersion: cfg.AppVersion,
		sm:         sm,
		be:         beStore,
	}

	// Create gRPC server
	grpcOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			middleware.GRPCLoggingMiddleware(logger),  // Add the logging middleware
			middleware.GRPCRecoveryMiddleware(logger), // Add the recovery middleware
		),
	}
	grpcSrv := grpc.NewServer(grpcOpts...)
	// Register the service with the gRPC server
	app.RegisterAppServer(grpcSrv, appServer)

	// Create HTTP server
	mux := runtime.NewServeMux()
	httpHandler := registerMiddlewares(logger, mux)
	httpServer := &http.Server{
		Handler: httpHandler,
		Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
	}
	// Register the service with the HTTP server
	err = app.RegisterAppHandlerServer(ctx, mux, appServer)
	if err != nil {
		logger.Fatal("Could not register handler", zap.Error(err))
	}

	return &Server{
		grpcServer: grpcSrv,
		httpServer: httpServer,
		config:     cfg,
		logger:     logger.With(zap.String("component", "server")),
	}
}

func (s *Server) Start() {
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.GRPCPort))
		if err != nil {
			log.Fatalf("Failed to listen on port %d: %v", s.config.GRPCPort, err)
		}
		s.logger.Info("Starting gRPC server", zap.Int("port", s.config.GRPCPort))
		if err := s.grpcServer.Serve(lis); err != nil {
			s.logger.Fatal("Failed to start gRPC server", zap.Error(err))
		}
	}()

	go func() {
		s.logger.Info("Starting HTTP server", zap.Int("port", s.config.HTTPPort))
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	if s.config.SwaggerEnabled {
		go func() {
			s.logger.Info("Starting Swagger server", zap.Int("port", s.config.SwaggerPort))
			if err := s.startSwaggerServer().ListenAndServe(); err != nil && err != http.ErrServerClosed {
				s.logger.Fatal("Failed to start Swagger server", zap.Error(err))
			}
		}()
	}
}

func (s *Server) Shutdown(ctx context.Context) {
	s.logger.Info("Shutting down servers...")
	s.grpcServer.GracefulStop()
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("HTTP server shutdown error", zap.Error(err))
	}
}

func (s *Server) startSwaggerServer() *http.Server {
	swaggerDir, err := filepath.Abs(s.config.SwaggerDir)
	if err != nil {
		s.logger.Fatal("Could not resolve Swagger directory path", zap.Error(err))
	}

	swaggerMux := http.NewServeMux()

	// Serve Swagger UI static files
	swaggerUIPath := filepath.Join("swagger-ui") // Path to Swagger UI files
	swaggerMux.Handle("/", http.FileServer(http.Dir(swaggerUIPath)))

	// Serve Swagger JSON files
	swaggerMux.Handle("/swagger/", http.StripPrefix("/swagger/", http.FileServer(http.Dir(swaggerDir))))

	swaggerServer := &http.Server{
		Handler: swaggerMux,
		Addr:    fmt.Sprintf(":%d", s.config.SwaggerPort),
	}

	return swaggerServer
}

func registerMiddlewares(logger *zap.Logger, mux *runtime.ServeMux) http.Handler {
	var handler http.Handler = mux
	handler = middleware.HTTPLoggingMiddleware(logger)(handler)
	handler = middleware.HTTPRecoveryMiddleware(logger)(handler)

	return handler
}
