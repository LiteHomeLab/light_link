package storage

import (
	"database/sql"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system
type User struct {
	ID           int64     `db:"id"`
	Username     string    `db:"username"`
	PasswordHash string    `db:"password_hash"`
	Role         string    `db:"role"`
	CreatedAt    time.Time `db:"created_at"`
}

// CreateUser creates a new user
func (d *Database) CreateUser(username, password, role string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("generate password hash: %w", err)
	}

	query := `INSERT INTO users (username, password_hash, role) VALUES (?, ?, ?)`
	_, err = d.db.Exec(query, username, string(hash), role)
	return err
}

// GetUser retrieves a user by username
func (d *Database) GetUser(username string) (*User, error) {
	query := `SELECT id, username, password_hash, role, created_at FROM users WHERE username = ?`
	row := d.db.QueryRow(query, username)

	var u User
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found: %s", username)
	}
	return &u, err
}

// ValidateUser validates a user's credentials
func (d *Database) ValidateUser(username, password string) (*User, error) {
	u, err := d.GetUser(username)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	return u, nil
}

// InitAdminUser initializes the admin user if it doesn't exist
func (d *Database) InitAdminUser(username, password string) error {
	_, err := d.GetUser(username)
	if err == nil {
		return nil // User already exists
	}

	return d.CreateUser(username, password, "admin")
}

// ListUsers returns all users
func (d *Database) ListUsers() ([]*User, error) {
	query := `SELECT id, username, password_hash, role, created_at FROM users ORDER BY username`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, &u)
	}

	return users, rows.Err()
}

// UpdateUserRole updates a user's role
func (d *Database) UpdateUserRole(username, role string) error {
	query := `UPDATE users SET role = ? WHERE username = ?`
	_, err := d.db.Exec(query, role, username)
	return err
}

// UpdateUserPassword updates a user's password
func (d *Database) UpdateUserPassword(username, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	query := `UPDATE users SET password_hash = ? WHERE username = ?`
	_, err = d.db.Exec(query, string(hash), username)
	return err
}

// DeleteUser deletes a user
func (d *Database) DeleteUser(username string) error {
	query := `DELETE FROM users WHERE username = ?`
	_, err := d.db.Exec(query, username)
	return err
}

// UserExists checks if a user exists
func (d *Database) UserExists(username string) bool {
	var exists bool
	err := d.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", username).Scan(&exists)
	return err == nil && exists
}
