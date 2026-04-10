package config

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
	return Config{
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
}
