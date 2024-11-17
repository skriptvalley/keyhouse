package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/skriptvalley/keyhouse/pkg/keystore"
	"github.com/skriptvalley/keyhouse/pkg/pb/app"
	"github.com/skriptvalley/keyhouse/pkg/statemanager"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AppServer struct {
	app.UnimplementedAppServer
	appVersion string
	be         keystore.BackendKeyStore
	sm         *statemanager.StateManager
}

// GetStatus returns the status of the service
func (s *AppServer) GetStatus(ctx context.Context, req *app.StatusRequest) (*app.StatusResponse, error) {
	status, err := s.sm.DB.GetVaultState(ctx)
	if err != nil {
		return &app.StatusResponse{
			Service:   "KeyHouse",
			Version:   s.appVersion,
			Status:    "unhealthy",
			Timestamp: timestamppb.New(time.Now()),
		}, err
	}
	return &app.StatusResponse{
		Service:   "KeyHouse",
		Version:   s.appVersion,
		Status:    status,
		Timestamp: timestamppb.New(time.Now()),
	}, nil
}

func (s *AppServer) InitKeyhouse(ctx context.Context, req *app.InitRequest) (*app.InitResponse, error) {
	var resp *app.InitResponse
	var err error
	if s.sm.IsVaultReady(ctx) {
		resp, err = &app.InitResponse{
			Status:     statemanager.VAULT_STATE_READY,
			Message:    "vault is initialized and ready",
			Keyholders: nil,
		}, nil
	} else if s.sm.IsVaultLocked(ctx) {
		active_keys, err := s.sm.DB.GetActiveKeysCount(ctx)
		if err != nil {
			return &app.InitResponse{
				Status:     statemanager.VAULT_STATE_LOCKED,
				Message:    "failed to fetch active keys count",
				Keyholders: nil,
			}, err
		}
		resp, err = &app.InitResponse{
			Status:     statemanager.VAULT_STATE_LOCKED,
			Message:    fmt.Sprintf("vault is locked. need %v keyholders to activate keyhouse", 3-active_keys),
			Keyholders: nil,
		}, nil
	} else if s.sm.IsVaultDown(ctx) {
		err = s.sm.GenerateKeys(ctx)
		if err != nil {
			return &app.InitResponse{
				Status:     "unknown",
				Message:    "failed to generate keys",
				Keyholders: nil,
			}, err
		}
		keyholdersMap, err := s.sm.DB.GetKeyholders(ctx)
		if err != nil {
			return &app.InitResponse{
				Status:     "unknown",
				Message:    "failed to fetch keyholders",
				Keyholders: nil,
			}, err
		}
		var keyholders []string
		for keyholder := range keyholdersMap {
			keyholders = append(keyholders, keyholder)
		}
		resp = &app.InitResponse{
			Status:     statemanager.VAULT_STATE_LOCKED,
			Message:    "vault is initialized. please distribute generated keys to different individuals over private channel",
			Keyholders: keyholders,
		}
	} else {
		resp, err = &app.InitResponse{
			Status:     "unknown",
			Message:    "vault is in unknown state",
			Keyholders: nil,
		}, fmt.Errorf("internal server error: %v", http.StatusInternalServerError)
	}
	return resp, err
}

func (s *AppServer) ActivateKey(ctx context.Context, req *app.ActivateKeyRequest) (*app.ActivateKeyResponse, error) {
	var resp *app.ActivateKeyResponse
	var err error
	keyholder := fmt.Sprintf("%v:%v", statemanager.KEY_PREFIX, req.GetKeyholder())
	is_ready, err := s.sm.UnlockVault(ctx, keyholder)
	if err != nil {
		return &app.ActivateKeyResponse{
			Status:  "unknown",
			Message: "failed to activate key",
		}, err
	}
	if is_ready {
		resp = &app.ActivateKeyResponse{
			Status:  statemanager.VAULT_STATE_READY,
			Message: "vault is activated",
		}
	} else {
		resp = &app.ActivateKeyResponse{
			Status:  statemanager.VAULT_STATE_LOCKED,
			Message: "key activated",
		}
	}
	return resp, err
}
