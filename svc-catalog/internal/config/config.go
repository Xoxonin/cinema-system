package config

import "os"

type Config struct {
	HTTPAddress string
}

func Load() Config {
	addr := os.Getenv("CATALOG_HTTP_ADDR")
	if addr == "" {
		addr = ":8081"
	}

	return Config{HTTPAddress: addr}
}
