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
		Role:    DefaultRole,
		HTTP:    Server{Host: "0.0.0.0", Port: 8085},
		GRPC:    Server{Host: "0.0.0.0", Port: 9005},
		Targets: map[string]string{},
	}
}
