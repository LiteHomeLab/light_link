package storage

import (
	"database/sql"
	"fmt"
	"time"
)

// Instance represents a service instance
type Instance struct {
	ID            int64     `db:"id" json:"id"`
	ServiceName   string    `db:"service_name" json:"service_name"`
	InstanceKey   string    `db:"instance_key" json:"instance_key"`
	Language      string    `db:"language" json:"language"`
	HostIP        string    `db:"host_ip" json:"host_ip"`
	HostMAC       string    `db:"host_mac" json:"host_mac"`
	WorkingDir    string    `db:"working_dir" json:"working_dir"`
	Version       string    `db:"version" json:"version"`
	FirstSeen     time.Time `db:"first_seen" json:"first_seen"`
	LastHeartbeat time.Time `db:"last_heartbeat" json:"last_heartbeat"`
	Online        bool      `db:"online" json:"online"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

// SaveInstance saves or updates an instance record
func (d *Database) SaveInstance(inst *Instance) error {
	query := `
	INSERT INTO instances (service_name, instance_key, language, host_ip, host_mac, working_dir, version, first_seen, last_heartbeat, online)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(instance_key) DO UPDATE SET
		service_name = excluded.service_name,
		language = excluded.language,
		host_ip = excluded.host_ip,
		host_mac = excluded.host_mac,
		working_dir = excluded.working_dir,
		version = excluded.version,
		last_heartbeat = excluded.last_heartbeat,
		online = excluded.online,
		updated_at = CURRENT_TIMESTAMP
	`
	_, err := d.db.Exec(query, inst.ServiceName, inst.InstanceKey, inst.Language,
		inst.HostIP, inst.HostMAC, inst.WorkingDir, inst.Version,
		inst.FirstSeen, inst.LastHeartbeat, inst.Online)
	return err
}

// GetInstance retrieves an instance by its instance key
func (d *Database) GetInstance(instanceKey string) (*Instance, error) {
	query := `SELECT id, service_name, instance_key, language, host_ip, host_mac, working_dir, version, first_seen, last_heartbeat, online, created_at, updated_at FROM instances WHERE instance_key = ?`
	row := d.db.QueryRow(query, instanceKey)

	var inst Instance
	err := row.Scan(&inst.ID, &inst.ServiceName, &inst.InstanceKey, &inst.Language,
		&inst.HostIP, &inst.HostMAC, &inst.WorkingDir, &inst.Version,
		&inst.FirstSeen, &inst.LastHeartbeat, &inst.Online, &inst.CreatedAt, &inst.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("instance not found: %s", instanceKey)
	}
	return &inst, err
}
