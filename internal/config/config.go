package config

import (
	"flag"
	"os"
)

type Config struct {
	ListenAddr string
	DataDir    string
	PublicURL  string
}

func Load() Config {
	var cfg Config

	flag.StringVar(&cfg.ListenAddr, "listen", getenv("CORE_LISTEN_ADDR", ":8080"), "HTTP listen address")
	flag.StringVar(&cfg.DataDir, "data-dir", getenv("CORE_DATA_DIR", "data"), "data directory")
	flag.StringVar(&cfg.PublicURL, "public-url", getenv("CORE_PUBLIC_URL", ""), "public URL advertised to edge nodes")
	flag.Parse()

	return cfg
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
