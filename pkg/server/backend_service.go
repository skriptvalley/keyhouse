package server

import (
	"context"
	"time"

	"github.com/skriptvalley/keyhouse/pkg/pb/backend"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type BackendServiceServer struct {
	backend.UnimplementedBackendServer
	appVersion string
}

// GetStatus returns the status of the service
func (s *BackendServiceServer) GetStatus(ctx context.Context, req *backend.StatusRequest) (*backend.StatusResponse, error) {
	return &backend.StatusResponse{
		Service:   "KeyHouse",
		Version:   s.appVersion,
		Status:    "healthy",
		Timestamp: timestamppb.New(time.Now()),
	}, nil
}
