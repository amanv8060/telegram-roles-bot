// Package main implements a production-ready Telegram bot for role management.
//
// The bot provides functionality for creating and managing roles within Telegram groups,
// allowing administrators to assign users to roles and ping all users in a role.
//
// Features:
//   - Role creation and management
//   - User assignment to roles
//   - Role-based pinging
//   - Admin access controls
//   - Rate limiting
//   - Health monitoring
//   - Graceful shutdown
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// run contains the main application logic
func run() error {
	// Load environment variables
	if err := loadEnvironment(); err != nil {
		return fmt.Errorf("failed to load environment: %w", err)
	}

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Initialize logger
	InitLogger(config.LogLevel)
	Logger.Info("Starting Telegram Role Bot")

	// Initialize database
	db, err := InitDB(config.DatabasePath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			Logger.WithError(closeErr).Error("Failed to close database")
		}
	}()

	// Initialize dependencies
	store := NewDBStore(db)
	security := NewSecurityMiddleware(config)
	healthChecker := NewHealthChecker(db)

	// Initialize bot
	bot, err := initializeBot(config)
	if err != nil {
		return fmt.Errorf("failed to initialize bot: %w", err)
	}

	// Create bot service
	botService := &TelegramBotService{
		bot:      bot,
		store:    store,
		security: security,
		config:   config,
	}

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	go handleShutdownSignals(cancel)

	// Start health check server
	healthPort := getEnvOrDefault("HEALTH_PORT", DefaultHealthPort)
	go StartHealthServer(healthPort, healthChecker)

	// Start bot
	Logger.Info("Bot started, listening for updates")
	if err := botService.Start(ctx); err != nil {
		return fmt.Errorf("bot service error: %w", err)
	}

	Logger.Info("Bot stopped gracefully")
	return nil
}

// loadEnvironment loads environment variables from .env file
func loadEnvironment() error {
	if err := godotenv.Load(); err != nil {
		// Don't fail if .env file doesn't exist in production
		if os.Getenv("ENV") != EnvProduction {
			fmt.Printf("Warning: Error loading .env file: %v\n", err)
		}
	}
	return nil
}

// initializeBot creates and configures the Telegram bot
func initializeBot(config *Config) (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(config.TelegramToken)
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") {
			return nil, fmt.Errorf("TELEGRAM_APITOKEN is invalid")
		}
		return nil, fmt.Errorf("failed to create bot API: %w", err)
	}

	bot.Debug = config.LogLevel == LogLevelDebug
	Logger.WithField("username", bot.Self.UserName).Info("Bot authorized successfully")

	return bot, nil
}

// handleShutdownSignals listens for shutdown signals and triggers graceful shutdown
func handleShutdownSignals(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	Logger.Info("Received shutdown signal")
	cancel()
}

// TelegramBotService implements the main bot service
type TelegramBotService struct {
	bot      *tgbotapi.BotAPI
	store    Store
	security SecurityValidator
	config   *Config
}

// Start starts the bot service
func (s *TelegramBotService) Start(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = s.config.UpdateTimeout

	updates := s.bot.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			Logger.Info("Shutdown requested, stopping bot")
			return nil
		case update := <-updates:
			if err := s.HandleUpdate(update); err != nil {
				Logger.WithError(err).Error("Failed to handle update")
			}
		}
	}
}

// Stop stops the bot service
func (s *TelegramBotService) Stop() error {
	s.bot.StopReceivingUpdates()
	return nil
}

// HandleUpdate processes incoming Telegram updates
func (s *TelegramBotService) HandleUpdate(update tgbotapi.Update) error {
	// Security validation
	if err := s.security.ValidateMessage(update); err != nil {
		Logger.WithError(err).Warn("Message validation failed")
		return err
	}

	if update.Message == nil {
		return nil
	}

	// Log message for debugging
	s.logMessage(update.Message)

	// Handle commands
	if update.Message.IsCommand() {
		return s.handleCommand(update)
	}

	// Handle role mentions
	if strings.HasPrefix(update.Message.Text, "@") {
		return s.handleRoleMention(update)
	}

	return nil
}

// logMessage logs incoming messages for debugging
func (s *TelegramBotService) logMessage(message *tgbotapi.Message) {
	Logger.WithFields(map[string]interface{}{
		"user_id":    message.From.ID,
		"username":   message.From.UserName,
		"chat_id":    message.Chat.ID,
		"message_id": message.MessageID,
		"text":       message.Text,
	}).Debug("Received message")
}

// handleCommand processes bot commands
func (s *TelegramBotService) handleCommand(update tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	command := update.Message.Command()
	args := update.Message.CommandArguments()

	// Check admin permissions for admin commands
	if AdminCommands[command] && !s.security.IsAdmin(update.Message.From.UserName) {
		msg.Text = MsgUnauthorized
		_, err := s.bot.Send(msg)
		return err
	}

	// Route command to appropriate handler
	switch command {
	case CmdPing:
		msg.Text = s.handlePingCommand(args)
	case CmdCreateRole:
		msg.Text = s.handleCreateRoleCommand(args)
	case CmdRemoveRole:
		msg.Text = s.handleRemoveRoleCommand(args)
	case CmdAddToRole:
		msg.Text = s.handleAddToRoleCommand(args)
	case CmdRemoveFromRole:
		msg.Text = s.handleRemoveFromRoleCommand(args)
	case CmdListRoles:
		msg.Text = s.handleListRolesCommand()
	case CmdListMembers:
		msg.Text = s.handleListMembersCommand(args)
	case CmdHelp:
		msg.Text = HelpMessage
	case CmdStatus:
		msg.Text = MsgBotHealthy
	default:
		msg.Text = MsgUnknownCommand
	}

	_, err := s.bot.Send(msg)
	return err
}

// handlePingCommand handles the ping command
func (s *TelegramBotService) handlePingCommand(args string) string {
	if args == "" {
		return MsgPong
	}

	users, err := s.store.GetUsersInRole(args)
	if err != nil {
		return fmt.Sprintf(PrefixError, err)
	}

	if len(users) == 0 {
		return fmt.Sprintf("❌ No users found in role '%s'", args)
	}

	msgText := fmt.Sprintf(PrefixPing, args)
	for _, user := range users {
		msgText += "@" + user + " "
	}
	return msgText
}

// handleCreateRoleCommand handles the createrole command
func (s *TelegramBotService) handleCreateRoleCommand(args string) string {
	if args == "" {
		return MsgProvideRoleName
	}

	if err := s.store.CreateRole(args); err != nil {
		return fmt.Sprintf(PrefixError, err)
	}

	return fmt.Sprintf(PrefixSuccess, fmt.Sprintf("Role '%s' created successfully", args))
}

// handleRemoveRoleCommand handles the removerole command
func (s *TelegramBotService) handleRemoveRoleCommand(args string) string {
	if args == "" {
		return MsgProvideRoleName
	}

	if err := s.store.RemoveRole(args); err != nil {
		return fmt.Sprintf(PrefixError, err)
	}

	return fmt.Sprintf(PrefixSuccess, fmt.Sprintf("Role '%s' removed successfully", args))
}

// handleAddToRoleCommand handles the addtorole command
func (s *TelegramBotService) handleAddToRoleCommand(args string) string {
	parts := strings.Split(args, " ")
	if len(parts) != 2 {
		return MsgUsageAddToRole
	}

	role, user := parts[0], parts[1]
	if err := s.store.AddUserToRole(role, user); err != nil {
		return fmt.Sprintf(PrefixError, err)
	}

	return fmt.Sprintf(PrefixSuccess, fmt.Sprintf("User %s added to role '%s'", user, role))
}

// handleRemoveFromRoleCommand handles the removefromrole command
func (s *TelegramBotService) handleRemoveFromRoleCommand(args string) string {
	parts := strings.Split(args, " ")
	if len(parts) != 2 {
		return MsgUsageRemoveFromRole
	}

	role, user := parts[0], parts[1]
	if err := s.store.RemoveUserFromRole(role, user); err != nil {
		return fmt.Sprintf(PrefixError, err)
	}

	return fmt.Sprintf(PrefixSuccess, fmt.Sprintf("User %s removed from role '%s'", user, role))
}

// handleListRolesCommand handles the listroles command
func (s *TelegramBotService) handleListRolesCommand() string {
	roles, err := s.store.GetAllRoles()
	if err != nil {
		return fmt.Sprintf(PrefixError, err)
	}

	if len(roles) == 0 {
		return MsgNoRoles
	}

	return fmt.Sprintf(PrefixInfo, "Roles: "+strings.Join(roles, ", "))
}

// handleListMembersCommand handles the listmembers command
func (s *TelegramBotService) handleListMembersCommand(args string) string {
	if args == "" {
		return MsgProvideRoleName
	}

	users, err := s.store.GetUsersInRole(args)
	if err != nil {
		return fmt.Sprintf(PrefixError, err)
	}

	if len(users) == 0 {
		return fmt.Sprintf("📋 No users found in role '%s'", args)
	}

	return fmt.Sprintf("📋 Users in role '%s': %s", args, strings.Join(users, ", "))
}

// handleRoleMention processes role mentions like @rolename
func (s *TelegramBotService) handleRoleMention(update tgbotapi.Update) error {
	role := strings.TrimPrefix(update.Message.Text, "@")
	role = strings.TrimSpace(role)

	users, err := s.store.GetUsersInRole(role)
	if err != nil {
		Logger.WithError(err).Error("Failed to get users in role")
		return err
	}

	if len(users) > 0 {
		msgText := fmt.Sprintf(PrefixPing, role)
		for _, user := range users {
			msgText += "@" + user + " "
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
		_, err := s.bot.Send(msg)
		return err
	}

	return nil
}
