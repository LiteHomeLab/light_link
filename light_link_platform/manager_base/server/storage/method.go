package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/LiteHomeLab/light_link/sdk/go/types"
)

// MethodMetadata represents method metadata in the database
type MethodMetadata struct {
	ID          int64                  `db:"id" json:"id"`
	ServiceID   int64                  `db:"service_id" json:"service_id"`
	Name        string                 `db:"name" json:"name"`
	Description string                 `db:"description" json:"description"`
	Params      []types.ParameterMetadata `db:"params" json:"parameters"`
	Returns     []types.ReturnMetadata   `db:"returns" json:"return_info"`
	Example     *types.ExampleMetadata   `db:"example" json:"examples"`
	Tags        []string               `db:"tags" json:"tags"`
	Deprecated  bool                   `db:"deprecated" json:"deprecated"`
	CreatedAt   time.Time              `db:"created_at" json:"created_at"`
}

// SaveMethod saves or updates a method
func (d *Database) SaveMethod(serviceID int64, meta *MethodMetadata) error {
	paramsJSON, _ := json.Marshal(meta.Params)
	returnsJSON, _ := json.Marshal(meta.Returns)
	exampleJSON, _ := json.Marshal(meta.Example)
	tagsJSON, _ := json.Marshal(meta.Tags)

	query := `
	INSERT INTO methods (service_id, name, description, params, returns, example, tags, deprecated)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(service_id, name) DO UPDATE SET
		description = excluded.description,
		params = excluded.params,
		returns = excluded.returns,
		example = excluded.example,
		tags = excluded.tags,
		deprecated = excluded.deprecated
	`

	_, err := d.db.Exec(query, serviceID, meta.Name, meta.Description,
		string(paramsJSON), string(returnsJSON), string(exampleJSON),
		string(tagsJSON), meta.Deprecated)
	return err
}

// SaveMethods saves multiple methods for a service
func (d *Database) SaveMethods(serviceID int64, methods []MethodMetadata) error {
	for _, m := range methods {
		if err := d.SaveMethod(serviceID, &m); err != nil {
			return err
		}
	}
	return nil
}

// GetMethods retrieves all methods for a service
func (d *Database) GetMethods(serviceName string) ([]*MethodMetadata, error) {
	query := `
	SELECT m.id, m.service_id, m.name, m.description, m.params, m.returns,
		   m.example, m.tags, m.deprecated, m.created_at
	FROM methods m
	INNER JOIN services s ON s.id = m.service_id
	WHERE s.name = ?
	ORDER BY m.name
	`

	rows, err := d.db.Query(query, serviceName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	methods := []*MethodMetadata{}
	for rows.Next() {
		var m MethodMetadata
		var paramsJSON, returnsJSON, exampleJSON, tagsJSON string
		if err := rows.Scan(&m.ID, &m.ServiceID, &m.Name, &m.Description,
			&paramsJSON, &returnsJSON, &exampleJSON, &tagsJSON,
			&m.Deprecated, &m.CreatedAt); err != nil {
			return nil, err
		}
		if paramsJSON != "" {
			json.Unmarshal([]byte(paramsJSON), &m.Params)
		}
		if returnsJSON != "" {
			json.Unmarshal([]byte(returnsJSON), &m.Returns)
		}
		if exampleJSON != "" {
			json.Unmarshal([]byte(exampleJSON), &m.Example)
		}
		if tagsJSON != "" {
			json.Unmarshal([]byte(tagsJSON), &m.Tags)
		}
		methods = append(methods, &m)
	}

	return methods, rows.Err()
}

// GetMethod retrieves a specific method
func (d *Database) GetMethod(serviceName, methodName string) (*MethodMetadata, error) {
	query := `
	SELECT m.id, m.service_id, m.name, m.description, m.params, m.returns,
		   m.example, m.tags, m.deprecated, m.created_at
	FROM methods m
	INNER JOIN services s ON s.id = m.service_id
	WHERE s.name = ? AND m.name = ?
	`

	row := d.db.QueryRow(query, serviceName, methodName)

	var m MethodMetadata
	var paramsJSON, returnsJSON, exampleJSON, tagsJSON string
	err := row.Scan(&m.ID, &m.ServiceID, &m.Name, &m.Description,
		&paramsJSON, &returnsJSON, &exampleJSON, &tagsJSON,
		&m.Deprecated, &m.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("method not found: %s.%s", serviceName, methodName)
	}
	if err != nil {
		return nil, err
	}

	if paramsJSON != "" {
		json.Unmarshal([]byte(paramsJSON), &m.Params)
	}
	if returnsJSON != "" {
		json.Unmarshal([]byte(returnsJSON), &m.Returns)
	}
	if exampleJSON != "" {
		json.Unmarshal([]byte(exampleJSON), &m.Example)
	}
	if tagsJSON != "" {
		json.Unmarshal([]byte(tagsJSON), &m.Tags)
	}
	return &m, nil
}

// GetMethodID returns the ID of a method
func (d *Database) GetMethodID(serviceID int64, methodName string) (int64, error) {
	var id int64
	err := d.db.QueryRow("SELECT id FROM methods WHERE service_id = ? AND name = ?",
		serviceID, methodName).Scan(&id)
	return id, err
}

// DeleteMethods deletes all methods for a service
func (d *Database) DeleteMethods(serviceID int64) error {
	_, err := d.db.Exec("DELETE FROM methods WHERE service_id = ?", serviceID)
	return err
}
