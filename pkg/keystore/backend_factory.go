package keystore

import "fmt"

type KeystoreType string

type BackendKeyStore interface {
	Ping() error
	Store(storageId, key string, value []byte) error
	Retrieve(storageId, key string) ([]byte, error)
	Delete(storageId, key string) error
}

func NewKeystore(storeType, cfgPath string) (BackendKeyStore, error) {
	switch storeType {
	case "postgres":
		pgCfg, err := LoadPostgresConfig(cfgPath)
		if err != nil {
			return nil, err
		}
		return NewPostgresStore(pgCfg)
	default:
		return nil, fmt.Errorf("unknown store type %s", storeType)
	}
}
