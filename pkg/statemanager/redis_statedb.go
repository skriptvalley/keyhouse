package statemanager

import (
	"context"

	redis "github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type RedisDB struct {
	client *redis.Client
	logger    *zap.Logger
}

func NewRedisDB(logger *zap.Logger, host, port, password string) *RedisDB {
	rdb := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       STATE_DB,
	})
	return &RedisDB{
		client: rdb,
		logger: logger.With(zap.String("component", "redisdb")),
	}
}

func (rdb *RedisDB) Ping(ctx context.Context) error {
	return rdb.client.Ping(ctx).Err()
}

func (rdb  *RedisDB) GetVaultState(ctx context.Context) (string, error) {
	return rdb.client.Get(ctx, STATE_KEY).Result()
}

func (rdb *RedisDB) SetVaultState(ctx context.Context, state string) error {
	return rdb.client.Set(ctx, STATE_KEY, state, 0).Err()
}

func (rdb *RedisDB) CreateOrGetInitCode(ctx context.Context) (string, error) {
	code, err := rdb.client.Get(ctx, INIT_CODE_KEY).Result()
	if err == redis.Nil {
		code = uuid.New().String()
		err = rdb.client.Set(ctx, INIT_CODE_KEY, code, 0).Err()
		if err != nil {
			rdb.logger.Error("failed to set init code", zap.Error(err))
			return "", err
		}
	} else if err != nil {
		rdb.logger.Error("failed to get init code", zap.Error(err))
		return "", err
	}
	rdb.logger.Debug("fetched init code", zap.String("code", code))
	return code, nil
}

func (rdb *RedisDB) AddKeyholder(ctx context.Context, keyholder string) error {
	return rdb.client.Set(ctx, KEY_PREFIX+":"+keyholder, false, 0).Err()
}

func (rdb *RedisDB) GetKeyholders(ctx context.Context) (map[string]bool, error) {
	keyholders := make(map[string]bool, 5)
	keys, err := rdb.client.Keys(ctx, KEY_PREFIX+":*").Result()
	if err != nil {
		return nil, err
	}
	if len(keys) != 5 {
		rdb.logger.Error("invalid number of keyholders", zap.Int("keyholder count", len(keys)))
		return nil, err
	}
	for _, key := range keys {
		keyholders[key], err = rdb.client.Get(ctx, key).Bool()
		if err != nil {
			rdb.logger.Error("failed to parse keyholder state", zap.String("keyholder", key))
			return nil, err
		}
	}
	rdb.logger.Debug("fetched keyholders", zap.Any("keyholders", keyholders))
	return keyholders, nil
}

func (rdb *RedisDB) GetActiveKeysCount(ctx context.Context) (int, error) {
	count, err := rdb.client.Get(ctx, ACTIVE_KEY_COUNT).Int()
	if err != nil {
		rdb.logger.Error("failed to get active keys count", zap.Error(err))
		return 0, err
	}
	rdb.logger.Debug("fetched active keys count", zap.Int("count", count))
	return count, nil
}

func (rdb *RedisDB) SetActiveKeysCount(ctx context.Context, count int) error {
	return rdb.client.Set(ctx, ACTIVE_KEY_COUNT, count, 0).Err()
}

func (rdb *RedisDB) ActivateKey(ctx context.Context, keyholder string) error {
	current_state, err := rdb.client.Get(ctx, keyholder).Bool()
	if err != nil {
		rdb.logger.Error("failed to get keyholder state", zap.String("keyholder", keyholder))
		return err
	}
	// if the keyholder is already active, do nothing
	if current_state {
		rdb.logger.Info("key is already activated", zap.String("keyholder", keyholder))
		return nil
	}
	return rdb.client.Set(ctx, keyholder, true, 0).Err()
}

func (rdb *RedisDB) CleanKeyholders(ctx context.Context) error {
	keyholders, err := rdb.GetKeyholders(ctx)
	if err != nil {
		if err == redis.Nil {
			rdb.logger.Debug("no keyholders found")
			return nil
		}
		rdb.logger.Error("failed to get keyholders", zap.Error(err))
		return err
	}
	// delete all keyholders
	for keyholder := range keyholders {
		err = rdb.client.Del(ctx, keyholder).Err()
		if err != nil {
			rdb.logger.Error("failed to delete keyholder", zap.String("keyholder", keyholder))
			return err
		}
	}
	return nil
}

func (rdb *RedisDB) Close() error {
	return rdb.client.Close()
}
