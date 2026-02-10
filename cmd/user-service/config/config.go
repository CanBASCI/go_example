package config

import (
	"net/url"
	"os"
	"strings"
)

// Config holds user-service configuration.
type Config struct {
	ServerPort      string
	DB              DBConfig
	Kafka           KafkaConfig
	OrderServiceURL string
}

// DBConfig holds PostgreSQL configuration.
type DBConfig struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
}

// KafkaConfig holds Kafka broker configuration.
type KafkaConfig struct {
	Brokers []string
}

// Load reads configuration from environment.
func Load() *Config {
	return &Config{
		ServerPort: getEnv("SERVER_PORT", "8081"),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			Database: getEnv("DB_NAME", "user_db"),
			User:     getEnv("DB_USER", "user"),
			Password: getEnv("DB_PASSWORD", "password"),
		},
		Kafka: KafkaConfig{
			Brokers: getEnvSlice("KAFKA_BOOTSTRAP_SERVERS", []string{"localhost:9092"}),
		},
		OrderServiceURL: getEnv("ORDER_SERVICE_URL", "http://localhost:8091"),
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

// DSN returns PostgreSQL connection string (password URL-escaped).
func (c *DBConfig) DSN() string {
	user := url.UserPassword(c.User, c.Password)
	u := &url.URL{
		Scheme:   "postgres",
		User:     user,
		Host:     c.Host + ":" + c.Port,
		Path:     "/" + c.Database,
		RawQuery: "sslmode=disable",
	}
	return u.String()
}
