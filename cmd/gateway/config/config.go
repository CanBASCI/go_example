package config

import (
	"os"
	"strings"
)

// Config holds gateway configuration.
type Config struct {
	Port            string
	UserServiceURLs []string
	OrderServiceURL string
}

// Load reads configuration from environment.
func Load() *Config {
	return &Config{
		Port:            getEnv("PORT", "8080"),
		UserServiceURLs: getEnvSlice("USER_SERVICE_URLS", []string{"http://user-service-1:8081", "http://user-service-2:8082"}),
		OrderServiceURL: getEnv("ORDER_SERVICE_URL", "http://order-service:8091"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvSlice(key string, fallback []string) []string {
	if v := os.Getenv(key); v != "" {
		parts := strings.Split(v, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			if s := strings.TrimSpace(p); s != "" {
				out = append(out, s)
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	return fallback
}
