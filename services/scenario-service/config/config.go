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
		HTTP: Server{Host: "0.0.0.0", Port: 8084},
		GRPC: Server{Host: "0.0.0.0", Port: 9004},
		Targets: map[string]string{
			DeviceService:       "127.0.0.1:9002",
			ContextService:      "127.0.0.1:9003",
			NotificationService: "127.0.0.1:9007",
		},
	}
}
