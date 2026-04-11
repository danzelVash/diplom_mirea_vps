package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
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
	Role       string
	HTTP       Server
	GRPC       Server
	Postgres   Postgres
	RetryAfter time.Duration
	Targets    map[string]string
}

func LoadDefault() Config {
	cfg := Config{
		Role: DefaultRole,
		HTTP: Server{Host: "0.0.0.0", Port: 8081},
		GRPC: Server{Host: "0.0.0.0", Port: 9001},
		Postgres: Postgres{
			Host:     "127.0.0.1",
			Port:     5432,
			User:     "smarthome",
			Password: "smarthome",
			Database: "edgebridge",
			SSLMode:  "disable",
		},
		RetryAfter: 30 * time.Second,
		Targets: map[string]string{
			DeviceService:   "127.0.0.1:9002",
			ScenarioService: "127.0.0.1:9004",
			VoiceService:    "127.0.0.1:9005",
		},
	}

	cfg.HTTP.Host = envString("EDGE_BRIDGE_HTTP_HOST", cfg.HTTP.Host)
	cfg.HTTP.Port = envInt("EDGE_BRIDGE_HTTP_PORT", cfg.HTTP.Port)
	cfg.GRPC.Host = envString("EDGE_BRIDGE_GRPC_HOST", cfg.GRPC.Host)
	cfg.GRPC.Port = envInt("EDGE_BRIDGE_GRPC_PORT", cfg.GRPC.Port)
	cfg.Postgres.Host = envString("EDGE_BRIDGE_DB_HOST", cfg.Postgres.Host)
	cfg.Postgres.Port = envInt("EDGE_BRIDGE_DB_PORT", cfg.Postgres.Port)
	cfg.Postgres.User = envString("EDGE_BRIDGE_DB_USER", cfg.Postgres.User)
	cfg.Postgres.Password = envString("EDGE_BRIDGE_DB_PASSWORD", cfg.Postgres.Password)
	cfg.Postgres.Database = envString("EDGE_BRIDGE_DB_NAME", cfg.Postgres.Database)
	cfg.Postgres.SSLMode = envString("EDGE_BRIDGE_DB_SSLMODE", cfg.Postgres.SSLMode)
	cfg.RetryAfter = envDuration("EDGE_BRIDGE_COMMAND_RETRY_AFTER", cfg.RetryAfter)

	cfg.Targets[DeviceService] = envString("EDGE_BRIDGE_DEVICE_SERVICE_ADDR", cfg.Targets[DeviceService])
	cfg.Targets[ScenarioService] = envString("EDGE_BRIDGE_SCENARIO_SERVICE_ADDR", cfg.Targets[ScenarioService])
	cfg.Targets[VoiceService] = envString("EDGE_BRIDGE_VOICE_SERVICE_ADDR", cfg.Targets[VoiceService])

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

func envDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}
