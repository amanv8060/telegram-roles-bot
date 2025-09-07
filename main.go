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
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		// Don't fail if .env file doesn't exist in production
		if os.Getenv("ENV") != "production" {
			fmt.Printf("Warning: Error loading .env file: %v\n", err)
		}
	}

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		fmt.Printf("Configuration error: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	InitLogger(config.LogLevel)
	Logger.Info("Starting Telegram Role Bot")

	// Initialize database
	db, err := InitDB(config.DatabasePath)
	if err != nil {
		Logger.WithError(err).Fatal("Failed to initialize database")
	}
	defer func() {
		if err := db.Close(); err != nil {
			Logger.WithError(err).Error("Failed to close database")
		}
	}()

	// Initialize store
	store := NewDBStore(db)

	// Initialize security middleware
	security := NewSecurityMiddleware(config)

	// Initialize bot
	bot, err := tgbotapi.NewBotAPI(config.TelegramToken)
	if err != nil {
		if strings.Contains(err.Error(), "Not Found") {
			Logger.Fatal("TELEGRAM_APITOKEN is invalid. Please check your .env file.")
		}
		Logger.WithError(err).Fatal("Failed to initialize bot")
	}

	bot.Debug = config.LogLevel == "debug"
	Logger.WithField("username", bot.Self.UserName).Info("Bot authorized successfully")

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		Logger.Info("Received shutdown signal")
		cancel()
	}()

	// Start health check server
	healthChecker := NewHealthChecker(db)
	healthPort := getEnvOrDefault("HEALTH_PORT", "8080")
	go StartHealthServer(healthPort, healthChecker)

	// Start bot
	if err := runBot(ctx, bot, store, security, config); err != nil {
		Logger.WithError(err).Error("Bot stopped with error")
		os.Exit(1)
	}

	Logger.Info("Bot stopped gracefully")
}

func runBot(ctx context.Context, bot *tgbotapi.BotAPI, store *DBStore, security *SecurityMiddleware, config *Config) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = config.UpdateTimeout

	updates := bot.GetUpdatesChan(u)
	Logger.Info("Bot started, listening for updates")

	for {
		select {
		case <-ctx.Done():
			Logger.Info("Shutdown requested, stopping bot")
			return nil
		case update := <-updates:
			if err := handleUpdate(bot, store, security, update); err != nil {
				Logger.WithError(err).Error("Failed to handle update")
			}
		}
	}
}

func handleUpdate(bot *tgbotapi.BotAPI, store *DBStore, security *SecurityMiddleware, update tgbotapi.Update) error {
	// Security validation
	if err := security.ValidateMessage(update); err != nil {
		Logger.WithError(err).Warn("Message validation failed")
		return err
	}

	if update.Message == nil {
		return nil
	}

	// Log message
	Logger.WithFields(map[string]interface{}{
		"user_id":    update.Message.From.ID,
		"username":   update.Message.From.UserName,
		"chat_id":    update.Message.Chat.ID,
		"message_id": update.Message.MessageID,
		"text":       update.Message.Text,
	}).Debug("Received message")

	// Handle commands
	if update.Message.IsCommand() {
		return handleCommand(bot, store, security, update)
	}

	// Handle role mentions
	if strings.HasPrefix(update.Message.Text, "@") {
		return handleRoleMention(bot, store, update)
	}

	return nil
}

func handleCommand(bot *tgbotapi.BotAPI, store *DBStore, security *SecurityMiddleware, update tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	command := update.Message.Command()
	args := update.Message.CommandArguments()

	// Admin commands
	adminCommands := map[string]bool{
		"createrole":     true,
		"removerole":     true,
		"addtorole":      true,
		"removefromrole": true,
	}

	if _, ok := adminCommands[command]; ok {
		if !security.IsAdmin(update.Message.From.UserName) {
			msg.Text = "❌ You are not authorized to use this command."
			_, err := bot.Send(msg)
			return err
		}
	}

	switch command {
	case "ping":
		if args == "" {
			msg.Text = "🏓 pong"
		} else {
			users, err := store.GetUsersInRole(args)
			if err != nil {
				msg.Text = fmt.Sprintf("❌ Error: %v", err)
			} else if len(users) == 0 {
				msg.Text = fmt.Sprintf("❌ No users found in role '%s'", args)
			} else {
				msgText := fmt.Sprintf("📢 Pinging role '%s': ", args)
				for _, user := range users {
					msgText += "@" + user + " "
				}
				msg.Text = msgText
			}
		}

	case "createrole":
		if args == "" {
			msg.Text = "❌ Please provide a role name."
		} else {
			if err := store.CreateRole(args); err != nil {
				msg.Text = fmt.Sprintf("❌ Error: %v", err)
			} else {
				msg.Text = fmt.Sprintf("✅ Role '%s' created successfully.", args)
			}
		}

	case "removerole":
		if args == "" {
			msg.Text = "❌ Please provide a role name."
		} else {
			if err := store.RemoveRole(args); err != nil {
				msg.Text = fmt.Sprintf("❌ Error: %v", err)
			} else {
				msg.Text = fmt.Sprintf("✅ Role '%s' removed successfully.", args)
			}
		}

	case "addtorole":
		parts := strings.Split(args, " ")
		if len(parts) != 2 {
			msg.Text = "❌ Usage: /addtorole <rolename> <username>"
		} else {
			role, user := parts[0], parts[1]
			if err := store.AddUserToRole(role, user); err != nil {
				msg.Text = fmt.Sprintf("❌ Error: %v", err)
			} else {
				msg.Text = fmt.Sprintf("✅ User %s added to role '%s' successfully.", user, role)
			}
		}

	case "removefromrole":
		parts := strings.Split(args, " ")
		if len(parts) != 2 {
			msg.Text = "❌ Usage: /removefromrole <rolename> <username>"
		} else {
			role, user := parts[0], parts[1]
			if err := store.RemoveUserFromRole(role, user); err != nil {
				msg.Text = fmt.Sprintf("❌ Error: %v", err)
			} else {
				msg.Text = fmt.Sprintf("✅ User %s removed from role '%s' successfully.", user, role)
			}
		}

	case "listroles":
		roles, err := store.GetAllRoles()
		if err != nil {
			msg.Text = fmt.Sprintf("❌ Error: %v", err)
		} else if len(roles) == 0 {
			msg.Text = "📋 No roles found."
		} else {
			msg.Text = "📋 Roles: " + strings.Join(roles, ", ")
		}

	case "listmembers":
		if args == "" {
			msg.Text = "❌ Please provide a role name."
		} else {
			users, err := store.GetUsersInRole(args)
			if err != nil {
				msg.Text = fmt.Sprintf("❌ Error: %v", err)
			} else if len(users) == 0 {
				msg.Text = fmt.Sprintf("📋 No users found in role '%s'.", args)
			} else {
				msg.Text = fmt.Sprintf("📋 Users in role '%s': %s", args, strings.Join(users, ", "))
			}
		}

	case "help":
		msg.Text = `🤖 **Telegram Role Bot Commands**

**General Commands:**
/ping - Test if the bot is working
/ping <rolename> - Ping all users in a role
/listroles - List all roles
/listmembers <rolename> - List members of a role
/help - Show this help message

**Admin Commands:**
/createrole <rolename> - Create a new role
/removerole <rolename> - Remove a role
/addtorole <rolename> <username> - Add a user to a role
/removefromrole <rolename> <username> - Remove a user from a role

**Role Mentions:**
@<rolename> - Ping all users in a role

**Examples:**
/ping developers
/createrole developers
/addtorole developers john_doe
@developers`

	case "status":
		msg.Text = "🟢 Bot is running and healthy!"

	default:
		msg.Text = "❌ Unknown command. Use /help to see available commands."
	}

	_, err := bot.Send(msg)
	return err
}

func handleRoleMention(bot *tgbotapi.BotAPI, store *DBStore, update tgbotapi.Update) error {
	role := strings.TrimPrefix(update.Message.Text, "@")
	role = strings.TrimSpace(role)

	users, err := store.GetUsersInRole(role)
	if err != nil {
		Logger.WithError(err).Error("Failed to get users in role")
		return err
	}

	if len(users) > 0 {
		msgText := fmt.Sprintf("📢 Pinging role @%s: ", role)
		for _, user := range users {
			msgText += "@" + user + " "
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
		_, err := bot.Send(msg)
		return err
	}

	return nil
}
