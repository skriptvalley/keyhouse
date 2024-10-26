package middleware

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LoggingMiddleware logs each incoming gRPC request
func GRPCLoggingMiddleware(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		start := time.Now()
		// Call the handler to complete the request
		resp, err = handler(ctx, req)

		// Log the method information and execution time
		logger.Info("gRPC Request received",
			zap.String("method", info.FullMethod),
			zap.Duration("duration", time.Since(start)),
			zap.Error(err),
		)

		return resp, err
	}
}

// RecoveryMiddleware catches and logs panics in gRPC handlers
func GRPCRecoveryMiddleware(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = status.Errorf(codes.Internal, "Internal Server Error")
				logger.Error("Panic recovered in gRPC handler", zap.Any("recover", r), zap.Error(err))
			}
		}()
		return handler(ctx, req)
	}
}
