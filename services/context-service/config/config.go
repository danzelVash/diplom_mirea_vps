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
		HTTP: Server{Host: "0.0.0.0", Port: 8083},
		GRPC: Server{Host: "0.0.0.0", Port: 9003},
		Targets: map[string]string{
			DeviceService: "127.0.0.1:9002",
			VoiceService:  "127.0.0.1:9005",
			VisionService: "127.0.0.1:9006",
		},
	}
}
