package config

import (
	"os"
	"time"
)

type Config struct {
	ServerAddr  string
	DatabaseURL string
	RedisURL    string
	BaseURL     string // e.g. https://myqrcode.com
	CacheTTL    time.Duration
}

func Load() *Config {
	return &Config{
		ServerAddr:  getEnv("SERVER_ADDR", ":8080"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		RedisURL:    getEnv("REDIS_URL", ""),
		BaseURL:     getEnv("BASE_URL", "http://localhost:8080"),
		CacheTTL:    24 * time.Hour,
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
