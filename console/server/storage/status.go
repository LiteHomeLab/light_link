package storage

import (
	"database/sql"
	"time"
)

// ServiceStatus represents the status of a service
type ServiceStatus struct {
	ID         int64     `db:"id"`
	ServiceID  int64     `db:"service_id"`
	ServiceName string   `db:"service_name"`
	Online     bool      `db:"online"`
	LastSeen   time.Time `db:"last_seen"`
	Version    string    `db:"version"`
	UpdatedAt  time.Time `db:"updated_at"`
}

// UpdateServiceStatus updates the status of a service
func (d *Database) UpdateServiceStatus(serviceName string, online bool, version string) error {
	// Get service ID
	serviceID, err := d.GetServiceID(serviceName)
	if err != nil {
		return err
	}

	now := time.Now()

	// Update status
	query := `
	INSERT INTO service_status (service_id, online, last_seen, version, updated_at)
	VALUES (?, ?, ?, ?, ?)
	ON CONFLICT(service_id) DO UPDATE SET
		online = excluded.online,
		last_seen = excluded.last_seen,
		version = excluded.version,
		updated_at = excluded.updated_at
	`
	_, err = d.db.Exec(query, serviceID, online, now, version, now)
	if err != nil {
		return err
	}

	// Record history
	_, err = d.db.Exec("INSERT INTO service_status_history (service_id, online, timestamp) VALUES (?, ?, ?)",
		serviceID, online, now)
	return err
}

// GetServiceStatus retrieves the status of a service
func (d *Database) GetServiceStatus(serviceName string) (*ServiceStatus, error) {
	query := `
	SELECT ss.id, ss.service_id, s.name as service_name, ss.online, ss.last_seen, ss.version, ss.updated_at
	FROM service_status ss
	INNER JOIN services s ON s.id = ss.service_id
	WHERE s.name = ?
	`

	row := d.db.QueryRow(query, serviceName)

	var s ServiceStatus
	err := row.Scan(&s.ID, &s.ServiceID, &s.ServiceName, &s.Online,
		&s.LastSeen, &s.Version, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil // Service not found, not an error
	}
	return &s, err
}

// ListServiceStatus retrieves all service statuses
func (d *Database) ListServiceStatus() ([]*ServiceStatus, error) {
	query := `
	SELECT ss.id, ss.service_id, s.name as service_name, ss.online, ss.last_seen, ss.version, ss.updated_at
	FROM service_status ss
	INNER JOIN services s ON s.id = ss.service_id
	ORDER BY s.name
	`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statuses []*ServiceStatus
	for rows.Next() {
		var s ServiceStatus
		if err := rows.Scan(&s.ID, &s.ServiceID, &s.ServiceName, &s.Online,
			&s.LastSeen, &s.Version, &s.UpdatedAt); err != nil {
			return nil, err
		}
		statuses = append(statuses, &s)
	}

	return statuses, rows.Err()
}

// GetOnlineServices returns a list of online services
func (d *Database) GetOnlineServices() ([]string, error) {
	query := `
	SELECT s.name
	FROM service_status ss
	INNER JOIN services s ON s.id = ss.service_id
	WHERE ss.online = 1
	`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		services = append(services, name)
	}

	return services, rows.Err()
}

// GetOfflineServices returns a list of offline services
func (d *Database) GetOfflineServices() ([]string, error) {
	query := `
	SELECT s.name
	FROM service_status ss
	INNER JOIN services s ON s.id = ss.service_id
	WHERE ss.online = 0
	`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		services = append(services, name)
	}

	return services, rows.Err()
}

// GetServiceStatusHistory retrieves status history for a service
func (d *Database) GetServiceStatusHistory(serviceName string, limit int) ([]map[string]interface{}, error) {
	query := `
	SELECT ssh.online, ssh.timestamp
	FROM service_status_history ssh
	INNER JOIN services s ON s.id = ssh.service_id
	WHERE s.name = ?
	ORDER BY ssh.timestamp DESC
	LIMIT ?
	`

	rows, err := d.db.Query(query, serviceName, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []map[string]interface{}
	for rows.Next() {
		var online bool
		var timestamp time.Time
		if err := rows.Scan(&online, &timestamp); err != nil {
			return nil, err
		}
		history = append(history, map[string]interface{}{
			"online":    online,
			"timestamp": timestamp,
		})
	}

	return history, rows.Err()
}

// CleanupStatusHistory removes old status history records
func (d *Database) CleanupStatusHistory(olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	_, err := d.db.Exec("DELETE FROM service_status_history WHERE timestamp < ?", cutoff)
	return err
}
