package config

import "os"

// ...existing code...

type ServerConfig struct {
	Address string
}

type Config struct {
	Server ServerConfig
}

func Load() *Config {
	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	return &Config{Server: ServerConfig{Address: addr}}
}
