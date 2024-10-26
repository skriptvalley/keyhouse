package server

import (
	"context"

	"github.com/skriptvalley/keyhouse/pkg/pb/backend"
)

type BackendServiceServer struct {
	backend.UnimplementedBackendServer
	appVersion string
}

// GetStatus returns the status of the service
func (s *BackendServiceServer) GetStatus(ctx context.Context, req *backend.StatusRequest) (*backend.StatusResponse, error) {
	return &backend.StatusResponse{
		Service: "KeyHouse",
		Version: s.appVersion,
		Status:  "healthy",
	}, nil
}
