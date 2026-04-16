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
	HTTP                 Server
	GRPC                 Server
	Postgres             Postgres
	CommandRetryAfter    time.Duration
	VoiceRecognitionAddr string
}

func LoadDefault() Config {
	cfg := Config{
		HTTP: Server{Host: "0.0.0.0", Port: 8080},
		GRPC: Server{Host: "0.0.0.0", Port: 9000},
		Postgres: Postgres{
			Host:     "127.0.0.1",
			Port:     5432,
			User:     "smarthome",
			Password: "smarthome",
			Database: "smarthome",
			SSLMode:  "disable",
		},
		CommandRetryAfter:    30 * time.Second,
		VoiceRecognitionAddr: "127.0.0.1:9010",
	}

	cfg.HTTP.Host = envString("PLATFORM_HTTP_HOST", cfg.HTTP.Host)
	cfg.HTTP.Port = envInt("PLATFORM_HTTP_PORT", cfg.HTTP.Port)
	cfg.GRPC.Host = envString("PLATFORM_GRPC_HOST", cfg.GRPC.Host)
	cfg.GRPC.Port = envInt("PLATFORM_GRPC_PORT", cfg.GRPC.Port)
	cfg.Postgres.Host = envString("PLATFORM_DB_HOST", cfg.Postgres.Host)
	cfg.Postgres.Port = envInt("PLATFORM_DB_PORT", cfg.Postgres.Port)
	cfg.Postgres.User = envString("PLATFORM_DB_USER", cfg.Postgres.User)
	cfg.Postgres.Password = envString("PLATFORM_DB_PASSWORD", cfg.Postgres.Password)
	cfg.Postgres.Database = envString("PLATFORM_DB_NAME", cfg.Postgres.Database)
	cfg.Postgres.SSLMode = envString("PLATFORM_DB_SSLMODE", cfg.Postgres.SSLMode)
	cfg.CommandRetryAfter = envDuration("PLATFORM_COMMAND_RETRY_AFTER", cfg.CommandRetryAfter)
	cfg.VoiceRecognitionAddr = envString("PLATFORM_VOICE_RECOGNITION_ADDR", cfg.VoiceRecognitionAddr)

	return cfg
}

func (c Config) HTTPAddr() string {
	return fmt.Sprintf("%s:%d", c.HTTP.Host, c.HTTP.Port)
}

func (c Config) GRPCAddr() string {
	return fmt.Sprintf("%s:%d", c.GRPC.Host, c.GRPC.Port)
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
