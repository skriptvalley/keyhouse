package statemanager

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	STATE_DB         = 0
	STATE_KEY        = "state"
	INIT_CODE_KEY    = "init_code"
	ACTIVE_KEY_COUNT = "active_keys_count"
	KEY_PREFIX       = "key"
)

const (
	VAULT_STATE_DOWN   = "down"
	VAULT_STATE_LOCKED = "locked"
	VAULT_STATE_READY  = "ready"
)

type IStateDB interface {
	Ping(ctx context.Context) error
	GetVaultState(ctx context.Context) (string, error)
	SetVaultState(ctx context.Context, state string) error
	CreateOrGetInitCode(ctx context.Context) (string, error)
	AddKeyholder(ctx context.Context, keyholder string) error
	GetKeyholders(ctx context.Context) (map[string]bool, error)
	GetActiveKeysCount(ctx context.Context) (int, error)
	SetActiveKeysCount(ctx context.Context, count int) error
	ActivateKey(ctx context.Context, keyholder string) error
	CleanKeyholders(ctx context.Context) error
}

type StateManager struct {
	DB     IStateDB
	logger *zap.Logger
}

func NewStateManager(logger *zap.Logger, host, port, password string) *StateManager {
	rdb := NewRedisDB(logger, host, port, password)
	return &StateManager{
		DB:     rdb,
		logger: logger.With(zap.String("component", "statemanager")),
	}
}

func (sm *StateManager) Ping(ctx context.Context) error {
	return sm.DB.Ping(ctx)
}

func (sm *StateManager) InitStateDBCache(ctx context.Context) error {
	state, err := sm.DB.GetVaultState(ctx)
	if err == redis.Nil {
		err = sm.DB.SetVaultState(ctx, VAULT_STATE_DOWN)
		if err != nil {
			sm.logger.Error("error setting initial vault state", zap.Error(err))
			return err
		}
		sm.logger.Info(fmt.Sprintf("vault state set to %s", VAULT_STATE_DOWN))
		err = sm.DB.SetActiveKeysCount(ctx, 0)
		if err != nil {
			sm.logger.Error("error creating active keys count", zap.Error(err))
			return err
		}
	} else if err != nil {
		sm.logger.Error("error getting vault state", zap.Error(err))
		return err
	} else {
		sm.logger.Info("vault state cache already initialized", zap.String("state", state))
	}
	return nil
}

func (sm *StateManager) SetVaultState(ctx context.Context, state string) error {
	return sm.DB.SetVaultState(ctx, state)
}

func (sm *StateManager) IsVaultReady(ctx context.Context) bool {
	state, err := sm.DB.GetVaultState(ctx)
	if err != nil {
		sm.logger.Fatal("error getting vault state", zap.Error(err))
		return false
	}
	return state == VAULT_STATE_READY
}

func (sm *StateManager) IsVaultLocked(ctx context.Context) bool {
	state, err := sm.DB.GetVaultState(ctx)
	if err != nil {
		sm.logger.Error("error getting vault state", zap.Error(err))
		return false
	}
	return state == VAULT_STATE_LOCKED
}

func (sm *StateManager) IsVaultDown(ctx context.Context) bool {
	state, err := sm.DB.GetVaultState(ctx)
	if err != nil {
		sm.logger.Error("error getting vault state", zap.Error(err))
		return false
	}
	return state == VAULT_STATE_DOWN
}

func (sm *StateManager) GetInitCode(ctx context.Context) (string, error) {
	return sm.DB.CreateOrGetInitCode(ctx)
}

func (sm *StateManager) GenerateKeys(ctx context.Context) error {
	err := sm.DB.CleanKeyholders(ctx)
	if err != nil {
		sm.logger.Error("error in cleaning existing keyholders")
		return err
	}
	for i := 0; i < 5; i++ {
		id := uuid.New().String()
		err = sm.DB.AddKeyholder(ctx, id)
		if err != nil {
			sm.logger.Error("error adding keyholder")
			return err
		}
	}
	sm.SetVaultState(ctx, VAULT_STATE_LOCKED)
	sm.logger.Info("new keys generated")
	return nil
}

func (sm *StateManager) UnlockVault(ctx context.Context, keyholder string) (bool, error) {
	is_ready := sm.IsVaultReady(ctx)
	keyholders, err := sm.DB.GetKeyholders(ctx)
	if err != nil {
		sm.logger.Error("error getting keyholders")
		return is_ready, err
	}
	if _, ok := keyholders[keyholder]; !ok {
		sm.logger.Info("keyholder id not found")
		return is_ready, fmt.Errorf("invalid_key")
	}
	err = sm.DB.ActivateKey(ctx, keyholder)
	if err != nil {
		sm.logger.Error("error activating key")
		return false, err
	}
	sm.logger.Info("keyholder activated")
	active_keys, err := sm.DB.GetActiveKeysCount(ctx)
	if err != nil {
		sm.logger.Error("error getting active keys count")
		return false, err
	}
	active_keys += 1
	err = sm.DB.SetActiveKeysCount(ctx, active_keys)
	if err != nil {
		sm.logger.Error("error updating active keys count", zap.Int("new count", active_keys))
		return false, err
	}
	if !is_ready && active_keys >= 3 {
		err = sm.DB.SetVaultState(ctx, VAULT_STATE_READY)
		if err != nil {
			sm.logger.Error("error setting vault state")
			return false, err
		}
		sm.logger.Info("vault unlocked", zap.Int("active_keys", active_keys))
		is_ready = true
	}
	return is_ready, nil
}
