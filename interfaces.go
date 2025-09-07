package main

import (
	"context"
	"database/sql"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Store interface defines the contract for data storage operations
type Store interface {
	CreateRole(role string) error
	RemoveRole(role string) error
	AddUserToRole(role, user string) error
	RemoveUserFromRole(role, user string) error
	GetUsersInRole(role string) ([]string, error)
	GetAllRoles() ([]string, error)
}

// SecurityValidator interface defines the contract for security validation
type SecurityValidator interface {
	ValidateMessage(update tgbotapi.Update) error
	IsAdmin(username string) bool
}

// HealthCheckerInterface defines the contract for health checking
type HealthCheckerInterface interface {
	Check(ctx context.Context) error
}

// LoggerInterface defines the contract for logging
type LoggerInterface interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	WithField(key string, value interface{}) LoggerInterface
	WithFields(fields map[string]interface{}) LoggerInterface
	WithError(err error) LoggerInterface
}

// Database interface defines the contract for database operations
type Database interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
	Begin() (*sql.Tx, error)
	Close() error
	Ping() error
}

// BotService defines the main bot service interface
type BotService interface {
	Start(ctx context.Context) error
	Stop() error
	HandleUpdate(update tgbotapi.Update) error
}

// CommandHandler defines the interface for handling bot commands
type CommandHandler interface {
	HandleCommand(bot *tgbotapi.BotAPI, store Store, security SecurityValidator, update tgbotapi.Update) error
	CanHandle(command string) bool
	IsAdminRequired() bool
}
