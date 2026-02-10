package config

import "os"

// Config holds user-service configuration.
type Config struct {
	ServerPort string
	DB        DBConfig
	Kafka     KafkaConfig
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
			Brokers: []string{getEnv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092")},
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// DSN returns PostgreSQL connection string.
func (c *DBConfig) DSN() string {
	return "postgres://" + c.User + ":" + c.Password + "@" + c.Host + ":" + c.Port + "/" + c.Database + "?sslmode=disable"
}
