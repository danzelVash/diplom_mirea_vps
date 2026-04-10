package config

import (
	"fmt"
	"os"
	"strconv"
)

type Server struct {
	Host string
	Port int
}

type Postgres struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

type Config struct {
	Role     string
	HTTP     Server
	GRPC     Server
	Postgres Postgres
	Targets  map[string]string
}

func LoadDefault() Config {
	cfg := Config{
		Role: DefaultRole,
		HTTP: Server{Host: "0.0.0.0", Port: 8082},
		GRPC: Server{Host: "0.0.0.0", Port: 9002},
		Postgres: Postgres{
			Host:     "127.0.0.1",
			Port:     5432,
			User:     "smarthome",
			Password: "smarthome",
			Database: "smarthome",
			SSLMode:  "disable",
		},
		Targets: map[string]string{},
	}

	cfg.HTTP.Host = envString("DEVICE_HTTP_HOST", cfg.HTTP.Host)
	cfg.HTTP.Port = envInt("DEVICE_HTTP_PORT", cfg.HTTP.Port)
	cfg.GRPC.Host = envString("DEVICE_GRPC_HOST", cfg.GRPC.Host)
	cfg.GRPC.Port = envInt("DEVICE_GRPC_PORT", cfg.GRPC.Port)

	cfg.Postgres.Host = envString("DEVICE_DB_HOST", cfg.Postgres.Host)
	cfg.Postgres.Port = envInt("DEVICE_DB_PORT", cfg.Postgres.Port)
	cfg.Postgres.User = envString("DEVICE_DB_USER", cfg.Postgres.User)
	cfg.Postgres.Password = envString("DEVICE_DB_PASSWORD", cfg.Postgres.Password)
	cfg.Postgres.Database = envString("DEVICE_DB_NAME", cfg.Postgres.Database)
	cfg.Postgres.SSLMode = envString("DEVICE_DB_SSLMODE", cfg.Postgres.SSLMode)

	return cfg
}

func (c Config) PostgresDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Postgres.User,
		c.Postgres.Password,
		c.Postgres.Host,
		c.Postgres.Port,
		c.Postgres.Database,
		c.Postgres.SSLMode,
	)
}

func envString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
