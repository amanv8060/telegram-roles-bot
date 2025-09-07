// Package models defines data models and custom errors.
package models

import "fmt"

// Custom error types for better error handling

type ErrRoleNotFound struct {
	Role string
}

func (e ErrRoleNotFound) Error() string {
	return fmt.Sprintf("role '%s' not found", e.Role)
}

type ErrRoleAlreadyExists struct {
	Role string
}

func (e ErrRoleAlreadyExists) Error() string {
	return fmt.Sprintf("role '%s' already exists", e.Role)
}

type ErrUserNotFound struct {
	User string
	Role string
}

func (e ErrUserNotFound) Error() string {
	return fmt.Sprintf("user '%s' not found in role '%s'", e.User, e.Role)
}

type ErrUnauthorized struct {
	Operation string
	User      string
}

func (e ErrUnauthorized) Error() string {
	return fmt.Sprintf("user '%s' is not authorized to perform operation '%s'", e.User, e.Operation)
}

type ErrRateLimited struct {
	UserID int64
}

func (e ErrRateLimited) Error() string {
	return fmt.Sprintf("rate limit exceeded for user %d", e.UserID)
}

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
