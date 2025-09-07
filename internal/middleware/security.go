// Package middleware provides security and rate limiting middleware.
package middleware

import (
	"fmt"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"didactic-spork/internal/config"
	"didactic-spork/internal/models"
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

// Security handles security validation
type Security struct {
	config      *config.Config
	rateLimiter *RateLimiter
}

// NewSecurity creates a new security middleware
func NewSecurity(cfg *config.Config) *Security {
	return &Security{
		config:      cfg,
		rateLimiter: NewRateLimiter(cfg.RateLimitPerMin, time.Minute),
	}
}

// ValidateMessage performs security validation on incoming messages
func (s *Security) ValidateMessage(update tgbotapi.Update) error {
	if update.Message == nil {
		return nil
	}

	// Check if chat is allowed
	if len(s.config.AllowedChats) > 0 {
		chatID := update.Message.Chat.ID
		if !s.isChatAllowed(chatID) {
			return fmt.Errorf("chat %d is not allowed", chatID)
		}
	}

	// Rate limiting
	userID := update.Message.From.ID
	if !s.rateLimiter.Allow(userID) {
		return models.ErrRateLimited{UserID: userID}
	}

	// Basic input validation
	if update.Message.Text != "" {
		text := strings.TrimSpace(update.Message.Text)
		const telegramMessageLimit = 4000
		if len(text) > telegramMessageLimit {
			return models.ErrInvalidInput{Field: "message", Value: "text", Reason: "message too long"}
		}
	}

	return nil
}

// isChatAllowed checks if a chat ID is in the allowed chats list
func (s *Security) isChatAllowed(chatID int64) bool {
	for _, allowedChat := range s.config.AllowedChats {
		if chatID == allowedChat {
			return true
		}
	}
	return false
}

// IsAdmin checks if a user is an admin
func (s *Security) IsAdmin(username string) bool {
	return username == s.config.AdminUsername
}
