package main

import (
	"fmt"
)

// Custom error types for better error handling

// ErrRoleNotFound indicates that a role was not found
type ErrRoleNotFound struct {
	Role string
}

func (e ErrRoleNotFound) Error() string {
	return fmt.Sprintf("role '%s' not found", e.Role)
}

// ErrRoleAlreadyExists indicates that a role already exists
type ErrRoleAlreadyExists struct {
	Role string
}

func (e ErrRoleAlreadyExists) Error() string {
	return fmt.Sprintf("role '%s' already exists", e.Role)
}

// ErrUserNotFound indicates that a user was not found
type ErrUserNotFound struct {
	User string
	Role string
}

func (e ErrUserNotFound) Error() string {
	return fmt.Sprintf("user '%s' not found in role '%s'", e.User, e.Role)
}

// ErrUnauthorized indicates that a user is not authorized
type ErrUnauthorized struct {
	Operation string
	User      string
}

func (e ErrUnauthorized) Error() string {
	return fmt.Sprintf("user '%s' is not authorized to perform operation '%s'", e.User, e.Operation)
}

// ErrRateLimited indicates that a user has exceeded rate limits
type ErrRateLimited struct {
	UserID int64
}

func (e ErrRateLimited) Error() string {
	return fmt.Sprintf("rate limit exceeded for user %d", e.UserID)
}

// ErrInvalidInput indicates invalid input was provided
type ErrInvalidInput struct {
	Field  string
	Value  string
	Reason string
}

func (e ErrInvalidInput) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("invalid %s '%s': %s", e.Field, e.Value, e.Reason)
	}
	return fmt.Sprintf("invalid %s '%s'", e.Field, e.Value)
}

// ErrDatabaseOperation indicates a database operation failed
type ErrDatabaseOperation struct {
	Operation string
	Err       error
}

func (e ErrDatabaseOperation) Error() string {
	return fmt.Sprintf("database operation '%s' failed: %v", e.Operation, e.Err)
}

func (e ErrDatabaseOperation) Unwrap() error {
	return e.Err
}
