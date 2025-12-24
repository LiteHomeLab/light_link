package storage

import (
	"time"
)

// ServiceEvent represents a service event
type ServiceEvent struct {
	ID        int64     `db:"id" json:"id"`
	Type      string    `db:"type" json:"type"`
	Service   string    `db:"service" json:"service"`
	Method    string    `db:"method" json:"method"`
	Data      string    `db:"data" json:"data"`
	Timestamp time.Time `db:"timestamp" json:"timestamp"`
}

// SaveEvent saves an event
func (d *Database) SaveEvent(event *ServiceEvent) error {
	query := `
	INSERT INTO events (type, service, method, data, timestamp)
	VALUES (?, ?, ?, ?, ?)
	`
	_, err := d.db.Exec(query, event.Type, event.Service, event.Method,
		event.Data, event.Timestamp)
	return err
}

// ListEvents retrieves events with pagination
func (d *Database) ListEvents(limit, offset int) ([]*ServiceEvent, error) {
	query := `
	SELECT id, type, service, method, data, timestamp
	FROM events
	ORDER BY timestamp DESC
	LIMIT ? OFFSET ?
	`

	rows, err := d.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []*ServiceEvent{}
	for rows.Next() {
		var e ServiceEvent
		if err := rows.Scan(&e.ID, &e.Type, &e.Service, &e.Method,
			&e.Data, &e.Timestamp); err != nil {
			return nil, err
		}
		events = append(events, &e)
	}

	return events, rows.Err()
}

// ListEventsByType retrieves events filtered by type
func (d *Database) ListEventsByType(eventType string, limit, offset int) ([]*ServiceEvent, error) {
	query := `
	SELECT id, type, service, method, data, timestamp
	FROM events
	WHERE type = ?
	ORDER BY timestamp DESC
	LIMIT ? OFFSET ?
	`

	rows, err := d.db.Query(query, eventType, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []*ServiceEvent{}
	for rows.Next() {
		var e ServiceEvent
		if err := rows.Scan(&e.ID, &e.Type, &e.Service, &e.Method,
			&e.Data, &e.Timestamp); err != nil {
			return nil, err
		}
		events = append(events, &e)
	}

	return events, rows.Err()
}

// ListEventsByService retrieves events for a specific service
func (d *Database) ListEventsByService(serviceName string, limit, offset int) ([]*ServiceEvent, error) {
	query := `
	SELECT id, type, service, method, data, timestamp
	FROM events
	WHERE service = ?
	ORDER BY timestamp DESC
	LIMIT ? OFFSET ?
	`

	rows, err := d.db.Query(query, serviceName, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []*ServiceEvent{}
	for rows.Next() {
		var e ServiceEvent
		if err := rows.Scan(&e.ID, &e.Type, &e.Service, &e.Method,
			&e.Data, &e.Timestamp); err != nil {
			return nil, err
		}
		events = append(events, &e)
	}

	return events, rows.Err()
}

// GetRecentEvents retrieves recent events across all types
func (d *Database) GetRecentEvents(limit int) ([]*ServiceEvent, error) {
	return d.ListEvents(limit, 0)
}

// CleanupOldEvents removes events older than the specified duration
func (d *Database) CleanupOldEvents(olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	_, err := d.db.Exec("DELETE FROM events WHERE timestamp < ?", cutoff)
	return err
}

// GetEventCount returns the total number of events
func (d *Database) GetEventCount() (int, error) {
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM events").Scan(&count)
	return count, err
}

// GetEventCountByType returns the count of events by type
func (d *Database) GetEventCountByType() (map[string]int, error) {
	query := `
	SELECT type, COUNT(*) as count
	FROM events
	GROUP BY type
	ORDER BY count DESC
	`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var eventType string
		var count int
		if err := rows.Scan(&eventType, &count); err != nil {
			return nil, err
		}
		result[eventType] = count
	}

	return result, rows.Err()
}
