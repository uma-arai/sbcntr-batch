package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/uma-arai/sbcntr-batch/internal/repository"
)

type Config struct {
	DB *repository.DBConfig
}

func Load() (*Config, error) {
	dbConfig := &repository.DBConfig{
		Host:     getEnvOrDefault("DB_HOST", "localhost"),
		Port:     getEnvAsIntOrDefault("DB_PORT", 5432),
		UserName: getEnvOrDefault("DB_USERNAME", "postgres"),
		Password: getEnvOrDefault("DB_PASSWORD", "postgres"),
		DBName:   getEnvOrDefault("DB_NAME", "echo_playground"),
		SSLMode:  getEnvOrDefault("DB_SSL_MODE", "disable"),
	}

	return &Config{
		DB: dbConfig,
	}, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	log.Printf("Environment variable %s is not set, using default value", key)
	return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}

	log.Printf("Environment variable %s is not set, using default value", key)
	return defaultValue
}

// GetDSN returns the database connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DB.Host, c.DB.Port, c.DB.UserName, c.DB.Password, c.DB.DBName, c.DB.SSLMode)
}
