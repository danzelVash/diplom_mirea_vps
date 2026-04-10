package config

type Server struct {
	Host string
	Port int
}

type Config struct {
	Role           string
	OfflineEnabled bool
	HTTP           Server
	GRPC           Server
}

func LoadDefault() Config {
	return Config{
		Role:           DefaultRole,
		OfflineEnabled: true,
		HTTP:           Server{Host: "0.0.0.0", Port: 8090},
		GRPC:           Server{Host: "0.0.0.0", Port: 9010},
	}
}

