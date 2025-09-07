package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/uma-arai/sbcntr-batch/internal/common/database"
)

type Config struct {
	DB  database.Config
	SFN struct {
		TaskToken string
	}
	EnableTracing bool
}

// LoadConfig は設定を読み込みます
func LoadConfig(taskToken string) (*Config, error) {
	cfg := &Config{
		DB: database.Config{
			Host:     getEnvOrDefault("DB_HOST", "localhost"),
			Port:     getEnvAsIntOrDefault("DB_PORT", 5432),
			UserName: getEnvOrDefault("DB_USERNAME", "sbcntrapp"),
			Password: getEnvOrDefault("DB_PASSWORD", "password"),
			DBName:   getEnvOrDefault("DB_NAME", "sbcntrapp"),
		},
		SFN: struct {
			TaskToken string
		}{
			TaskToken: taskToken,
		},
		EnableTracing: false,
	}

	// 環境変数[SBCNTR_ENABLE_TRACING]を見てトレースを有効にする。対応しているTracingはAWS_XRAYのみ。
	// 環境変数[AWS_XRAY_SDK_DISABLED]がtrueの場合は必ずトレースを無効にする。
	enableKey := os.Getenv("SBCNTR_ENABLE_TRACING")
	if !sdkDisabled() && (strings.ToLower(enableKey) == "true" || enableKey == "1") {
		os.Setenv("AWS_XRAY_SDK_DISABLED", "FALSE")
		cfg.EnableTracing = true
	} else {
		os.Setenv("AWS_XRAY_SDK_DISABLED", "TRUE")
		cfg.EnableTracing = false
	}

	return cfg, nil
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
	return defaultValue
}

// Check if SDK is disabled
func sdkDisabled() bool {
	disableKey := os.Getenv("AWS_XRAY_SDK_DISABLED")
	return strings.ToLower(disableKey) == "true"
}
