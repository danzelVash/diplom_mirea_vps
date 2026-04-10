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
		HTTP: Server{Host: "0.0.0.0", Port: 8080},
		GRPC: Server{Host: "0.0.0.0", Port: 9000},
		Targets: map[string]string{
			EdgeBridgeService:   "127.0.0.1:9001",
			DeviceService:       "127.0.0.1:9002",
			ContextService:      "127.0.0.1:9003",
			ScenarioService:     "127.0.0.1:9004",
			VoiceService:        "127.0.0.1:9005",
			VisionService:       "127.0.0.1:9006",
			NotificationService: "127.0.0.1:9007",
		},
	}

	cfg.HTTP.Host = envString("API_GATEWAY_HTTP_HOST", cfg.HTTP.Host)
	cfg.HTTP.Port = envInt("API_GATEWAY_HTTP_PORT", cfg.HTTP.Port)
	cfg.GRPC.Host = envString("API_GATEWAY_GRPC_HOST", cfg.GRPC.Host)
	cfg.GRPC.Port = envInt("API_GATEWAY_GRPC_PORT", cfg.GRPC.Port)

	cfg.Targets[EdgeBridgeService] = envString("API_GATEWAY_EDGE_BRIDGE_ADDR", cfg.Targets[EdgeBridgeService])
	cfg.Targets[DeviceService] = envString("API_GATEWAY_DEVICE_ADDR", cfg.Targets[DeviceService])
	cfg.Targets[ContextService] = envString("API_GATEWAY_CONTEXT_ADDR", cfg.Targets[ContextService])
	cfg.Targets[ScenarioService] = envString("API_GATEWAY_SCENARIO_ADDR", cfg.Targets[ScenarioService])
	cfg.Targets[VoiceService] = envString("API_GATEWAY_VOICE_ADDR", cfg.Targets[VoiceService])
	cfg.Targets[VisionService] = envString("API_GATEWAY_VISION_ADDR", cfg.Targets[VisionService])
	cfg.Targets[NotificationService] = envString("API_GATEWAY_NOTIFICATION_ADDR", cfg.Targets[NotificationService])

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
