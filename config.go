package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration for the bot
type Config struct {
	TelegramToken   string
	AdminUsername   string
	DatabasePath    string
	LogLevel        string
	MaxRetries      int
	UpdateTimeout   int
	AllowedChats    []int64
	RateLimitPerMin int
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	config := &Config{
		TelegramToken:   os.Getenv("TELEGRAM_APITOKEN"),
		AdminUsername:   os.Getenv("ADMIN_USERNAME"),
		DatabasePath:    getEnvOrDefault("DATABASE_PATH", "bot.db"),
		LogLevel:        getEnvOrDefault("LOG_LEVEL", "info"),
		MaxRetries:      getEnvIntOrDefault("MAX_RETRIES", 3),
		UpdateTimeout:   getEnvIntOrDefault("UPDATE_TIMEOUT", 60),
		RateLimitPerMin: getEnvIntOrDefault("RATE_LIMIT_PER_MIN", 30),
	}

	// Parse allowed chats if provided
	if allowedChatsStr := os.Getenv("ALLOWED_CHATS"); allowedChatsStr != "" {
		chats := strings.Split(allowedChatsStr, ",")
		for _, chat := range chats {
			if chatID, err := strconv.ParseInt(strings.TrimSpace(chat), 10, 64); err == nil {
				config.AllowedChats = append(config.AllowedChats, chatID)
			}
		}
	}

	// Validate required fields
	if config.TelegramToken == "" {
		return nil, fmt.Errorf("TELEGRAM_APITOKEN is required")
	}
	if config.AdminUsername == "" {
		return nil, fmt.Errorf("ADMIN_USERNAME is required")
	}

	return config, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
