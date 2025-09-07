// Package bot implements the main bot service and logic.
package bot

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"didactic-spork/internal/config"
	"didactic-spork/internal/handlers"
	"didactic-spork/internal/middleware"
	"didactic-spork/internal/store"
	"didactic-spork/pkg/logger"
)

// Service represents the main bot service
type Service struct {
	bot      *tgbotapi.BotAPI
	store    store.Store
	security *middleware.Security
	handlers *handlers.Commands
	config   *config.Config
	logger   *logger.Logger
}

// New creates a new bot service
func New(cfg *config.Config, db *sql.DB, log *logger.Logger) (*Service, error) {
	// Initialize Telegram bot
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") {
			return nil, fmt.Errorf("invalid TELEGRAM_APITOKEN")
		}
		return nil, fmt.Errorf("failed to create bot API: %w", err)
	}

	bot.Debug = cfg.LogLevel == "debug"
	log.WithField("username", bot.Self.UserName).Info("Bot authorized successfully")

	// Initialize dependencies
	roleStore := store.New(db)
	security := middleware.NewSecurity(cfg)
	commandHandlers := handlers.NewCommands(roleStore, security, log)

	// Start health check server
	go startHealthServer(cfg.HealthPort, db, log)

	return &Service{
		bot:      bot,
		store:    roleStore,
		security: security,
		handlers: commandHandlers,
		config:   cfg,
		logger:   log,
	}, nil
}

// Start starts the bot service
func (s *Service) Start(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = s.config.UpdateTimeout

	updates := s.bot.GetUpdatesChan(u)
	s.logger.Info("Bot started, listening for updates")

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Shutdown requested, stopping bot")
			return nil
		case update := <-updates:
			if err := s.handleUpdate(update); err != nil {
				s.logger.WithError(err).Error("Failed to handle update")
			}
		}
	}
}

// handleUpdate processes incoming Telegram updates
func (s *Service) handleUpdate(update tgbotapi.Update) error {
	// Security validation
	if err := s.security.ValidateMessage(update); err != nil {
		s.logger.WithError(err).Warn("Message validation failed")
		return err
	}

	if update.Message == nil {
		return nil
	}

	// Log message for debugging
	s.logMessage(update.Message)

	// Handle commands
	if update.Message.IsCommand() {
		return s.handlers.Handle(s.bot, update)
	}

	// Handle role mentions
	if strings.HasPrefix(update.Message.Text, "@") {
		return s.handleRoleMention(update)
	}

	return nil
}

// logMessage logs incoming messages for debugging
func (s *Service) logMessage(message *tgbotapi.Message) {
	s.logger.WithFields(map[string]interface{}{
		"user_id":    message.From.ID,
		"username":   message.From.UserName,
		"chat_id":    message.Chat.ID,
		"message_id": message.MessageID,
		"text":       message.Text,
	}).Debug("Received message")
}

// handleRoleMention processes role mentions like @rolename
func (s *Service) handleRoleMention(update tgbotapi.Update) error {
	role := strings.TrimPrefix(update.Message.Text, "@")
	role = strings.TrimSpace(role)
	role = strings.ToLower(role) // Normalize to lowercase

	users, err := s.store.GetUsersInRole(role)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get users in role")
		return err
	}

	if len(users) > 0 {
		msgText := fmt.Sprintf("Pinging role @%s: ", role)
		for _, user := range users {
			msgText += "@" + user + " "
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
		_, err := s.bot.Send(msg)
		return err
	}

	return nil
}

// startHealthServer starts the health check HTTP server
func startHealthServer(port string, db *sql.DB, log *logger.Logger) {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			log.WithError(err).Error("Health check failed")
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprint(w, "UNHEALTHY")
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "HEALTHY")
	})

	log.WithField("port", port).Info("Starting health check server")
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.WithError(err).Error("Health check server failed")
	}
}
