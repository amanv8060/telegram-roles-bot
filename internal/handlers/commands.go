// Package handlers implements command handlers for the bot.
package handlers

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"didactic-spork/internal/middleware"
	"didactic-spork/internal/models"
	"didactic-spork/internal/store"
	"didactic-spork/pkg/logger"
)

// Commands handles bot commands
type Commands struct {
	store    store.Store
	security *middleware.Security
	logger   *logger.Logger
}

// NewCommands creates a new command handler
func NewCommands(store store.Store, security *middleware.Security, logger *logger.Logger) *Commands {
	return &Commands{
		store:    store,
		security: security,
		logger:   logger,
	}
}

// Handle processes a bot command
func (c *Commands) Handle(bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	command := update.Message.Command()
	args := update.Message.CommandArguments()

	// Check admin permissions
	if models.AdminCommands[command] && !c.security.IsAdmin(update.Message.From.UserName) {
		msg.Text = models.MsgUnauthorized
		_, err := bot.Send(msg)
		return err
	}

	// Route command
	switch command {
	case models.CmdPing:
		msg.Text = c.handlePing(args)
	case models.CmdCreateRole:
		msg.Text = c.handleCreateRole(args)
	case models.CmdRemoveRole:
		msg.Text = c.handleRemoveRole(args)
	case models.CmdAddToRole:
		msg.Text = c.handleAddToRole(args)
	case models.CmdRemoveFromRole:
		msg.Text = c.handleRemoveFromRole(args)
	case models.CmdListRoles:
		msg.Text = c.handleListRoles()
	case models.CmdListMembers:
		msg.Text = c.handleListMembers(args)
	case models.CmdHelp:
		msg.Text = models.HelpMessage
	case models.CmdStatus:
		msg.Text = models.MsgBotHealthy
	default:
		msg.Text = models.MsgUnknownCommand
	}

	_, err := bot.Send(msg)
	return err
}

func (c *Commands) handlePing(args string) string {
	if args == "" {
		return models.MsgPong
	}

	// Normalize role name to lowercase
	roleName := strings.ToLower(strings.TrimSpace(args))

	users, err := c.store.GetUsersInRole(roleName)
	if err != nil {
		return fmt.Sprintf(models.PrefixError, err)
	}

	if len(users) == 0 {
		return fmt.Sprintf("No users found in role '%s'", roleName)
	}

	msgText := fmt.Sprintf(models.PrefixPing, roleName)
	for _, user := range users {
		msgText += "@" + user + " "
	}
	return msgText
}

func (c *Commands) handleCreateRole(args string) string {
	if args == "" {
		return models.MsgProvideRoleName
	}

	if err := c.store.CreateRole(args); err != nil {
		return fmt.Sprintf(models.PrefixError, err)
	}

	return fmt.Sprintf(models.PrefixSuccess, fmt.Sprintf("Role '%s' created successfully", args))
}

func (c *Commands) handleRemoveRole(args string) string {
	if args == "" {
		return models.MsgProvideRoleName
	}

	if err := c.store.RemoveRole(args); err != nil {
		return fmt.Sprintf(models.PrefixError, err)
	}

	return fmt.Sprintf(models.PrefixSuccess, fmt.Sprintf("Role '%s' removed successfully", args))
}

func (c *Commands) handleAddToRole(args string) string {
	parts := strings.Split(args, " ")
	if len(parts) != 2 {
		return models.MsgUsageAddToRole
	}

	role, user := parts[0], parts[1]
	if err := c.store.AddUserToRole(role, user); err != nil {
		return fmt.Sprintf(models.PrefixError, err)
	}

	return fmt.Sprintf(models.PrefixSuccess, fmt.Sprintf("User %s added to role '%s'", user, role))
}

func (c *Commands) handleRemoveFromRole(args string) string {
	parts := strings.Split(args, " ")
	if len(parts) != 2 {
		return models.MsgUsageRemoveFromRole
	}

	role, user := parts[0], parts[1]
	if err := c.store.RemoveUserFromRole(role, user); err != nil {
		return fmt.Sprintf(models.PrefixError, err)
	}

	return fmt.Sprintf(models.PrefixSuccess, fmt.Sprintf("User %s removed from role '%s'", user, role))
}

func (c *Commands) handleListRoles() string {
	roles, err := c.store.GetAllRoles()
	if err != nil {
		return fmt.Sprintf(models.PrefixError, err)
	}

	if len(roles) == 0 {
		return models.MsgNoRoles
	}

	return fmt.Sprintf(models.PrefixInfo, "Roles: "+strings.Join(roles, ", "))
}

func (c *Commands) handleListMembers(args string) string {
	if args == "" {
		return models.MsgProvideRoleName
	}

	// Normalize role name to lowercase
	roleName := strings.ToLower(strings.TrimSpace(args))

	users, err := c.store.GetUsersInRole(roleName)
	if err != nil {
		return fmt.Sprintf(models.PrefixError, err)
	}

	if len(users) == 0 {
		return fmt.Sprintf("No users found in role '%s'", roleName)
	}

	return fmt.Sprintf("Users in role '%s': %s", roleName, strings.Join(users, ", "))
}
