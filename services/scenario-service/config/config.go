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
		HTTP: Server{Host: "0.0.0.0", Port: 8084},
		GRPC: Server{Host: "0.0.0.0", Port: 9004},
		Postgres: Postgres{
			Host:     "127.0.0.1",
			Port:     5432,
			User:     "smarthome",
			Password: "smarthome",
			Database: "scenarios",
			SSLMode:  "disable",
		},
		Targets: map[string]string{
			DeviceService:       "127.0.0.1:9002",
			ContextService:      "127.0.0.1:9003",
			NotificationService: "127.0.0.1:9007",
		},
	}

	cfg.HTTP.Host = envString("SCENARIO_HTTP_HOST", cfg.HTTP.Host)
	cfg.HTTP.Port = envInt("SCENARIO_HTTP_PORT", cfg.HTTP.Port)
	cfg.GRPC.Host = envString("SCENARIO_GRPC_HOST", cfg.GRPC.Host)
	cfg.GRPC.Port = envInt("SCENARIO_GRPC_PORT", cfg.GRPC.Port)

	cfg.Postgres.Host = envString("SCENARIO_DB_HOST", cfg.Postgres.Host)
	cfg.Postgres.Port = envInt("SCENARIO_DB_PORT", cfg.Postgres.Port)
	cfg.Postgres.User = envString("SCENARIO_DB_USER", cfg.Postgres.User)
	cfg.Postgres.Password = envString("SCENARIO_DB_PASSWORD", cfg.Postgres.Password)
	cfg.Postgres.Database = envString("SCENARIO_DB_NAME", cfg.Postgres.Database)
	cfg.Postgres.SSLMode = envString("SCENARIO_DB_SSLMODE", cfg.Postgres.SSLMode)

	cfg.Targets[DeviceService] = envString("SCENARIO_DEVICE_SERVICE_ADDR", cfg.Targets[DeviceService])
	cfg.Targets[ContextService] = envString("SCENARIO_CONTEXT_SERVICE_ADDR", cfg.Targets[ContextService])
	cfg.Targets[NotificationService] = envString("SCENARIO_NOTIFICATION_SERVICE_ADDR", cfg.Targets[NotificationService])

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
