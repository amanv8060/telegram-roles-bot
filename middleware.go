package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// RateLimiter implements a simple rate limiter
type RateLimiter struct {
	mu       sync.RWMutex
	requests map[int64][]time.Time
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[int64][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Allow checks if a request is allowed for the given user
func (rl *RateLimiter) Allow(userID int64) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Clean old requests
	if requests, exists := rl.requests[userID]; exists {
		var validRequests []time.Time
		for _, req := range requests {
			if req.After(cutoff) {
				validRequests = append(validRequests, req)
			}
		}
		rl.requests[userID] = validRequests
	}

	// Check if under limit
	if len(rl.requests[userID]) >= rl.limit {
		return false
	}

	// Add current request
	rl.requests[userID] = append(rl.requests[userID], now)
	return true
}

// SecurityMiddleware handles security checks
type SecurityMiddleware struct {
	config      *Config
	rateLimiter *RateLimiter
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(config *Config) *SecurityMiddleware {
	return &SecurityMiddleware{
		config:      config,
		rateLimiter: NewRateLimiter(config.RateLimitPerMin, time.Minute),
	}
}

// ValidateMessage performs security validation on incoming messages
func (sm *SecurityMiddleware) ValidateMessage(update tgbotapi.Update) error {
	if update.Message == nil {
		return nil
	}

	// Check if chat is allowed (if restrictions are set)
	if len(sm.config.AllowedChats) > 0 {
		chatID := update.Message.Chat.ID
		allowed := false
		for _, allowedChat := range sm.config.AllowedChats {
			if chatID == allowedChat {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("chat %d is not allowed", chatID)
		}
	}

	// Rate limiting
	userID := update.Message.From.ID
	if !sm.rateLimiter.Allow(userID) {
		return fmt.Errorf("rate limit exceeded for user %d", userID)
	}

	// Basic input validation
	if update.Message.Text != "" {
		text := strings.TrimSpace(update.Message.Text)
		if len(text) > 4000 { // Telegram message limit
			return fmt.Errorf("message too long")
		}
	}

	return nil
}

// IsAdmin checks if a user is an admin
func (sm *SecurityMiddleware) IsAdmin(username string) bool {
	return username == sm.config.AdminUsername
}

// SanitizeInput sanitizes user input
func SanitizeInput(input string) string {
	// Remove potentially dangerous characters
	input = strings.TrimSpace(input)
	input = strings.ReplaceAll(input, "\n", " ")
	input = strings.ReplaceAll(input, "\r", " ")

	// Limit length
	if len(input) > 100 {
		input = input[:100]
	}

	return input
}
