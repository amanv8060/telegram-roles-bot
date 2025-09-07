package main

import (
	"database/sql"
	"fmt"
	"strings"
)

// DBStore manages roles and users in a SQLite database.
type DBStore struct {
	db *sql.DB
}

// NewDBStore creates a new DBStore.
func NewDBStore(db *sql.DB) *DBStore {
	return &DBStore{db: db}
}

// CreateRole creates a new role.
func (s *DBStore) CreateRole(role string) error {
	role = SanitizeInput(role)
	if role == "" {
		return fmt.Errorf("role name cannot be empty")
	}

	_, err := s.db.Exec("INSERT INTO roles (name) VALUES (?)", role)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return fmt.Errorf("role '%s' already exists", role)
		}
		return fmt.Errorf("failed to create role: %w", err)
	}

	Logger.WithField("role", role).Info("Role created successfully")
	return nil
}

// RemoveRole removes a role.
func (s *DBStore) RemoveRole(role string) error {
	role = SanitizeInput(role)
	if role == "" {
		return fmt.Errorf("role name cannot be empty")
	}

	result, err := s.db.Exec("DELETE FROM roles WHERE name = ?", role)
	if err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("role '%s' not found", role)
	}

	Logger.WithField("role", role).Info("Role removed successfully")
	return nil
}

// AddUserToRole adds a user to a role.
func (s *DBStore) AddUserToRole(role, user string) error {
	role = SanitizeInput(role)
	user = SanitizeInput(user)

	if role == "" {
		return fmt.Errorf("role name cannot be empty")
	}
	if user == "" {
		return fmt.Errorf("username cannot be empty")
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
		return fmt.Errorf("role '%s' does not exist", role)
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

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	Logger.WithFields(map[string]interface{}{
		"role": role,
		"user": user,
	}).Info("User added to role successfully")
	return nil
}

// GetUsersInRole returns the users in a role.
func (s *DBStore) GetUsersInRole(role string) ([]string, error) {
	role = SanitizeInput(role)
	if role == "" {
		return nil, fmt.Errorf("role name cannot be empty")
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
			Logger.WithError(err).Error("Failed to scan user")
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

// GetAllRoles returns all roles.
func (s *DBStore) GetAllRoles() ([]string, error) {
	rows, err := s.db.Query("SELECT name FROM roles ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("failed to get all roles: %w", err)
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			Logger.WithError(err).Error("Failed to scan role")
			continue
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// RemoveUserFromRole removes a user from a role.
func (s *DBStore) RemoveUserFromRole(role, user string) error {
	role = SanitizeInput(role)
	user = SanitizeInput(user)

	if role == "" {
		return fmt.Errorf("role name cannot be empty")
	}
	if user == "" {
		return fmt.Errorf("username cannot be empty")
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
		return fmt.Errorf("user '%s' not found in role '%s'", user, role)
	}

	Logger.WithFields(map[string]interface{}{
		"role": role,
		"user": user,
	}).Info("User removed from role successfully")
	return nil
}
