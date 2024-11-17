package keystore

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"
	_ "github.com/lib/pq"
)

var (
	postgresConfigFilePath = "/var/db/pg_config.yaml"
)

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(connCfg PostgresConfig) (*PostgresStore, error) {
	connStr := GetConnectionStringFromConfig(connCfg)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return &PostgresStore{db: db}, nil
}

func (p *PostgresStore) Ping() error {
	return p.db.Ping()
}

func (p *PostgresStore) Store(table, key string, value []byte) error {
	_, err := p.db.Exec("INSERT INTO $3 (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = $2", key, value, table)
	return err
}

func (p *PostgresStore) Retrieve(table, key string) ([]byte, error) {
	var value []byte
	err := p.db.QueryRow("SELECT value FROM $2 WHERE key = $1", key, table).Scan(&value)
	return value, err
}

func (p *PostgresStore) Delete(table, key string) error {
	_, err := p.db.Exec("DELETE FROM $2 WHERE key = $1", key, table)
	return err
}

func LoadPostgresConfig(cfgPath string) (PostgresConfig, error) {
	// Read the YAML file
	if cfgPath != "" {
		postgresConfigFilePath = cfgPath
	}
	data, err := os.ReadFile(postgresConfigFilePath)
	if err != nil {
		return PostgresConfig{}, fmt.Errorf("error reading config file: %w", err)
	} // Create a PostgresConfig struct to hold the parsed data
	var config PostgresConfig

	// Parse the YAML data into the struct
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return PostgresConfig{}, fmt.Errorf("error parsing YAML: %w", err)
	}

	return config, nil
}

func GetConnectionStringFromConfig(connCfg PostgresConfig) string {
	return "host=" + connCfg.Host + " port=" + strconv.Itoa(connCfg.Port) + " user=" + connCfg.User + " password=" + connCfg.Password + " dbname=" + connCfg.DBName + " sslmode=disable"
}
