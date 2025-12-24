package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// ServiceMetadata represents service metadata in the database
type ServiceMetadata struct {
	ID           int64     `db:"id"`
	Name         string    `db:"name"`
	Version      string    `db:"version"`
	Description  string    `db:"description"`
	Author       string    `db:"author"`
	Tags         []string  `db:"tags"`
	RegisteredAt time.Time `db:"registered_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// SaveService saves or updates a service
func (d *Database) SaveService(meta *ServiceMetadata) error {
	tagsJSON, _ := json.Marshal(meta.Tags)

	query := `
	INSERT INTO services (name, version, description, author, tags, registered_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(name) DO UPDATE SET
		version = excluded.version,
		description = excluded.description,
		author = excluded.author,
		tags = excluded.tags,
		updated_at = excluded.updated_at
	`

	now := time.Now()
	if meta.RegisteredAt.IsZero() {
		meta.RegisteredAt = now
	}
	meta.UpdatedAt = now

	_, err := d.db.Exec(query, meta.Name, meta.Version, meta.Description,
		meta.Author, string(tagsJSON), meta.RegisteredAt, now)
	return err
}

// GetService retrieves a service by name
func (d *Database) GetService(name string) (*ServiceMetadata, error) {
	query := `SELECT id, name, version, description, author, tags, registered_at, updated_at
			  FROM services WHERE name = ?`

	row := d.db.QueryRow(query, name)

	var s ServiceMetadata
	var tagsJSON string
	err := row.Scan(&s.ID, &s.Name, &s.Version, &s.Description,
		&s.Author, &tagsJSON, &s.RegisteredAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("service not found: %s", name)
	}
	if err != nil {
		return nil, err
	}

	if tagsJSON != "" {
		json.Unmarshal([]byte(tagsJSON), &s.Tags)
	}
	return &s, nil
}

// ListServices returns all services
func (d *Database) ListServices() ([]*ServiceMetadata, error) {
	query := `SELECT id, name, version, description, author, tags, registered_at, updated_at
			  FROM services ORDER BY registered_at DESC`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []*ServiceMetadata
	for rows.Next() {
		var s ServiceMetadata
		var tagsJSON string
		if err := rows.Scan(&s.ID, &s.Name, &s.Version, &s.Description,
			&s.Author, &tagsJSON, &s.RegisteredAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		if tagsJSON != "" {
			json.Unmarshal([]byte(tagsJSON), &s.Tags)
		}
		services = append(services, &s)
	}

	return services, rows.Err()
}

// DeleteService deletes a service by name
func (d *Database) DeleteService(name string) error {
	_, err := d.db.Exec("DELETE FROM services WHERE name = ?", name)
	return err
}

// GetServiceID returns the ID of a service by name
func (d *Database) GetServiceID(name string) (int64, error) {
	var id int64
	err := d.db.QueryRow("SELECT id FROM services WHERE name = ?", name).Scan(&id)
	return id, err
}

// ServiceExists checks if a service exists
func (d *Database) ServiceExists(name string) bool {
	var exists bool
	err := d.db.QueryRow("SELECT EXISTS(SELECT 1 FROM services WHERE name = ?)", name).Scan(&exists)
	return err == nil && exists
}
