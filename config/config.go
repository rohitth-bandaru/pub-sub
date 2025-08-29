package config

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

var (
	config *Config
	once   sync.Once
)

// Config holds all configuration for the pub-sub system
type Config struct {
	// Server configuration
	Port string
	Host string

	// Topic configuration
	MaxMessagesPerTopic int

	// WebSocket configuration
	ReadBufferSize  int
	WriteBufferSize int

	// Rate limiting (messages per second per topic)
	MaxPublishRate int

	// Logging configuration
	LogLevel  string
	LogFormat string
}

// LoadConfig loads configuration from environment variables with sensible defaults
// This function is thread-safe and loads config only once
func LoadConfig() *Config {
	once.Do(func() {
		// Load .env file if it exists
		if err := godotenv.Load(); err != nil {
			// .env file not found, use system environment variables
			logrus.Info("No .env file found, using system environment variables")
		} else {
			logrus.Info("Loaded configuration from .env file")
		}

		config = &Config{
			Port:                getEnv("PORT", "8080"),
			Host:                getEnv("HOST", "0.0.0.0"),
			MaxMessagesPerTopic: getEnvAsInt("MAX_MESSAGES_PER_TOPIC", 1000),
			ReadBufferSize:      getEnvAsInt("WS_READ_BUFFER_SIZE", 1024),
			WriteBufferSize:     getEnvAsInt("WS_WRITE_BUFFER_SIZE", 1024),
			MaxPublishRate:      getEnvAsInt("MAX_PUBLISH_RATE", 100),
			LogLevel:            getEnv("LOG_LEVEL", "info"),
			LogFormat:           getEnv("LOG_FORMAT", "text"),
		}

		logrus.Infof("Configuration loaded: Port=%s, Host=%s, MaxMessagesPerTopic=%d, MaxPublishRate=%d",
			config.Port, config.Host, config.MaxMessagesPerTopic, config.MaxPublishRate)
	})

	return config
}

// GetConfig returns the current configuration instance
// This function is thread-safe
func GetConfig() *Config {
	if config == nil {
		return LoadConfig()
	}
	return config
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		logrus.Warnf("Invalid value for %s: %s, using default: %d", key, value, defaultValue)
	}
	return defaultValue
}

// ValidateConfig validates the configuration and returns any errors
func (c *Config) ValidateConfig() error {
	if c.MaxMessagesPerTopic <= 0 {
		return fmt.Errorf("MAX_MESSAGES_PER_TOPIC must be positive, got: %d", c.MaxMessagesPerTopic)
	}

	if c.MaxPublishRate <= 0 {
		return fmt.Errorf("MAX_PUBLISH_RATE must be positive, got: %d", c.MaxPublishRate)
	}

	if c.ReadBufferSize <= 0 {
		return fmt.Errorf("WS_READ_BUFFER_SIZE must be positive, got: %d", c.ReadBufferSize)
	}

	if c.WriteBufferSize <= 0 {
		return fmt.Errorf("WS_WRITE_BUFFER_SIZE must be positive, got: %d", c.WriteBufferSize)
	}

	return nil
}

// String returns a string representation of the configuration
func (c *Config) String() string {
	return fmt.Sprintf("Config{Port: %s, Host: %s, MaxMessagesPerTopic: %d, MaxPublishRate: %d, ReadBufferSize: %d, WriteBufferSize: %d, LogLevel: %s, LogFormat: %s}",
		c.Port, c.Host, c.MaxMessagesPerTopic, c.MaxPublishRate, c.ReadBufferSize, c.WriteBufferSize, c.LogLevel, c.LogFormat)
}
