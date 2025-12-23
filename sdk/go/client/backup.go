package client

import (
	"encoding/base64"

	"github.com/LiteHomeLab/light_link/sdk/go/types"
)

// CreateBackup creates a full backup
func (c *Client) CreateBackup(serviceName, backupName string, data []byte) (int, error) {
	args := map[string]interface{}{
		"service_name": serviceName,
		"backup_name":  backupName,
		"data":         base64.StdEncoding.EncodeToString(data),
	}

	result, err := c.Call("backup-agent", "backup.create", args)
	if err != nil {
		return 0, err
	}

	versionFloat, ok := result["version"].(float64)
	if !ok {
		return 0, nil
	}

	return int(versionFloat), nil
}

// ListBackups lists all backup versions
func (c *Client) ListBackups(serviceName, backupName string) ([]types.BackupVersion, error) {
	args := map[string]interface{}{
		"service_name": serviceName,
		"backup_name":  backupName,
	}

	result, err := c.Call("backup-agent", "backup.list", args)
	if err != nil {
		return nil, err
	}

	versionsInterface, ok := result["versions"].([]interface{})
	if !ok {
		return []types.BackupVersion{}, nil
	}

	var versions []types.BackupVersion
	for _, v := range versionsInterface {
		vMap, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		version := types.BackupVersion{}
		if val, ok := vMap["version"].(float64); ok {
			version.Version = int(val)
		}
		if val, ok := vMap["file_size"].(float64); ok {
			version.FileSize = int64(val)
		}
		if val, ok := vMap["checksum"].(string); ok {
			version.Checksum = val
		}

		versions = append(versions, version)
	}

	return versions, nil
}

// GetBackup retrieves backup data for a specific version
func (c *Client) GetBackup(serviceName, backupName string, version int) ([]byte, error) {
	args := map[string]interface{}{
		"service_name": serviceName,
		"backup_name":  backupName,
		"version":      float64(version),
	}

	result, err := c.Call("backup-agent", "backup.get", args)
	if err != nil {
		return nil, err
	}

	dataBase64, ok := result["data"].(string)
	if !ok {
		return nil, nil
	}

	return base64.StdEncoding.DecodeString(dataBase64)
}

// DeleteBackup deletes a specific backup version
func (c *Client) DeleteBackup(serviceName, backupName string, version int) error {
	args := map[string]interface{}{
		"service_name": serviceName,
		"backup_name":  backupName,
		"version":      float64(version),
	}

	_, err := c.Call("backup-agent", "backup.delete", args)
	return err
}
