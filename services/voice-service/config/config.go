package config

import (
	"os"
	"strconv"
)

type Server struct {
	Host string
	Port int
}

type Config struct {
	Role    string
	HTTP    Server
	GRPC    Server
	Targets map[string]string
}

func LoadDefault() Config {
	cfg := Config{
		Role: DefaultRole,
		HTTP: Server{Host: "0.0.0.0", Port: 8085},
		GRPC: Server{Host: "0.0.0.0", Port: 9005},
		Targets: map[string]string{
			ScenarioService:         "127.0.0.1:9004",
			VoiceRecognitionService: "127.0.0.1:9010",
		},
	}

	cfg.HTTP.Host = envString("VOICE_HTTP_HOST", cfg.HTTP.Host)
	cfg.HTTP.Port = envInt("VOICE_HTTP_PORT", cfg.HTTP.Port)
	cfg.GRPC.Host = envString("VOICE_GRPC_HOST", cfg.GRPC.Host)
	cfg.GRPC.Port = envInt("VOICE_GRPC_PORT", cfg.GRPC.Port)

	cfg.Targets[ScenarioService] = envString("VOICE_SCENARIO_SERVICE_ADDR", cfg.Targets[ScenarioService])
	cfg.Targets[VoiceRecognitionService] = envString("VOICE_RECOGNITION_SERVICE_ADDR", cfg.Targets[VoiceRecognitionService])

	return cfg
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
