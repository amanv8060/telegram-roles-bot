// Package store provides data storage operations.
package store

import (
	"database/sql"
	"fmt"
	"strings"

	"didactic-spork/internal/models"
	"didactic-spork/pkg/utils"
)

// Store defines the interface for data storage operations
type Store interface {
	CreateRole(role string) error
	RemoveRole(role string) error
	AddUserToRole(role, user string) error
	RemoveUserFromRole(role, user string) error
	GetUsersInRole(role string) ([]string, error)
	GetAllRoles() ([]string, error)
}

// SQLStore implements Store interface using SQL database
type SQLStore struct {
	db *sql.DB
}

// New creates a new store instance
func New(db *sql.DB) Store {
	return &SQLStore{db: db}
}

// CreateRole creates a new role
func (s *SQLStore) CreateRole(role string) error {
	role = utils.SanitizeRoleName(role)
	if role == "" {
		return models.ErrInvalidInput{Field: "role name", Value: role, Reason: "cannot be empty"}
	}

	_, err := s.db.Exec("INSERT INTO roles (name) VALUES (?)", role)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return models.ErrRoleAlreadyExists{Role: role}
		}
		return fmt.Errorf("failed to create role: %w", err)
	}

	return nil
}

// RemoveRole removes a role
func (s *SQLStore) RemoveRole(role string) error {
	role = utils.SanitizeRoleName(role)
	if role == "" {
		return models.ErrInvalidInput{Field: "role name", Value: role, Reason: "cannot be empty"}
	}

	result, err := s.db.Exec("DELETE FROM roles WHERE name = ?", role)
	if err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return models.ErrRoleNotFound{Role: role}
	}

	return nil
}

// AddUserToRole adds a user to a role
func (s *SQLStore) AddUserToRole(role, user string) error {
	role = utils.SanitizeRoleName(role)
	user = utils.SanitizeUsername(user)

	if role == "" {
		return models.ErrInvalidInput{Field: "role name", Value: role, Reason: "cannot be empty"}
	}
	if user == "" {
		return models.ErrInvalidInput{Field: "username", Value: user, Reason: "cannot be empty"}
	}

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Ensure user exists
	_, err = tx.Exec("INSERT OR IGNORE INTO users (name) VALUES (?)", user)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Check if role exists
	var roleExists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM roles WHERE name = ?)", role).Scan(&roleExists)
	if err != nil {
		return fmt.Errorf("failed to check role existence: %w", err)
	}
	if !roleExists {
		return models.ErrRoleNotFound{Role: role}
	}

	// Add user to role
	_, err = tx.Exec(`
		INSERT OR IGNORE INTO role_users (role_id, user_id)
		SELECT r.id, u.id
		FROM roles r, users u
		WHERE r.name = ? AND u.name = ?
	`, role, user)
	if err != nil {
		return fmt.Errorf("failed to add user to role: %w", err)
	}

	return tx.Commit()
}

// RemoveUserFromRole removes a user from a role
func (s *SQLStore) RemoveUserFromRole(role, user string) error {
	role = utils.SanitizeRoleName(role)
	user = utils.SanitizeUsername(user)

	if role == "" {
		return models.ErrInvalidInput{Field: "role name", Value: role, Reason: "cannot be empty"}
	}
	if user == "" {
		return models.ErrInvalidInput{Field: "username", Value: user, Reason: "cannot be empty"}
	}

	result, err := s.db.Exec(`
		DELETE FROM role_users
		WHERE role_id = (SELECT id FROM roles WHERE name = ?)
		AND user_id = (SELECT id FROM users WHERE name = ?)
	`, role, user)
	if err != nil {
		return fmt.Errorf("failed to remove user from role: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return models.ErrUserNotFound{User: user, Role: role}
	}

	return nil
}

// GetUsersInRole returns the users in a role
func (s *SQLStore) GetUsersInRole(role string) ([]string, error) {
	role = utils.SanitizeRoleName(role)
	if role == "" {
		return nil, models.ErrInvalidInput{Field: "role name", Value: role, Reason: "cannot be empty"}
	}

	rows, err := s.db.Query(`
		SELECT u.name
		FROM users u
		JOIN role_users ru ON u.id = ru.user_id
		JOIN roles r ON r.id = ru.role_id
		WHERE r.name = ?
		ORDER BY u.name
	`, role)
	if err != nil {
		return nil, fmt.Errorf("failed to get users in role: %w", err)
	}
	defer rows.Close()

	var users []string
	for rows.Next() {
		var user string
		if err := rows.Scan(&user); err != nil {
			continue // Skip invalid entries
		}
		users = append(users, user)
	}

	return users, nil
}

// GetAllRoles returns all roles
func (s *SQLStore) GetAllRoles() ([]string, error) {
	rows, err := s.db.Query("SELECT name FROM roles ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("failed to get all roles: %w", err)
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			continue // Skip invalid entries
		}
		roles = append(roles, role)
	}

	return roles, nil
}
