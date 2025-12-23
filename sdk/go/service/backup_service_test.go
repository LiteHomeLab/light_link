package service

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBackupService_CreateAndList(t *testing.T) {
	// Create temp storage dir
	tempDir, err := os.MkdirTemp("", "backup-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create service
	svc, err := NewBackupService("test-backup-agent", "nats://localhost:4222", nil, tempDir)
	if err != nil {
		t.Skip("Need running NATS server:", err)
	}
	defer svc.Stop()

	if err := svc.Start(); err != nil {
		t.Fatal(err)
	}

	testData := []byte("test backup data v1")

	// Test create backup
	result, err := svc.handleCreateBackup(map[string]interface{}{
		"service_name": "test-service",
		"backup_name":  "test-db",
		"data":         base64.StdEncoding.EncodeToString(testData),
	})
	if err != nil {
		t.Fatal("create backup failed:", err)
	}

	versionFloat, ok := result["version"].(float64)
	if !ok || int(versionFloat) != 1 {
		t.Error("expected version 1, got", result["version"])
	}

	// Test list backups
	listResult, err := svc.handleListBackups(map[string]interface{}{
		"service_name": "test-service",
		"backup_name":  "test-db",
	})
	if err != nil {
		t.Fatal("list backups failed:", err)
	}

	currentVersion, ok := listResult["current_version"].(float64)
	if !ok || int(currentVersion) != 1 {
		t.Error("expected current_version 1, got", listResult["current_version"])
	}
}

func TestBackupService_GetAndDelete(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "backup-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	svc, err := NewBackupService("test-backup-agent", "nats://localhost:4222", nil, tempDir)
	if err != nil {
		t.Skip("Need running NATS server:", err)
	}
	defer svc.Stop()

	if err := svc.Start(); err != nil {
		t.Fatal(err)
	}

	testData := []byte("test backup data")

	// Create backup
	_, err = svc.handleCreateBackup(map[string]interface{}{
		"service_name": "test-service",
		"backup_name":  "test-db",
		"data":         base64.StdEncoding.EncodeToString(testData),
	})
	if err != nil {
		t.Fatal("create backup failed:", err)
	}

	// Test get backup
	getResult, err := svc.handleGetBackup(map[string]interface{}{
		"service_name": "test-service",
		"backup_name":  "test-db",
		"version":      float64(1),
	})
	if err != nil {
		t.Fatal("get backup failed:", err)
	}

	dataBase64, ok := getResult["data"].(string)
	if !ok {
		t.Fatal("expected data in result")
	}

	retrievedData, err := base64.StdEncoding.DecodeString(dataBase64)
	if err != nil {
		t.Fatal("decode data failed:", err)
	}

	if string(retrievedData) != string(testData) {
		t.Error("data mismatch, got", string(retrievedData))
	}

	// Test delete backup
	deleteResult, err := svc.handleDeleteBackup(map[string]interface{}{
		"service_name": "test-service",
		"backup_name":  "test-db",
		"version":      float64(1),
	})
	if err != nil {
		t.Fatal("delete backup failed:", err)
	}

	deleted, ok := deleteResult["deleted"].(bool)
	if !ok || !deleted {
		t.Error("expected deleted=true")
	}

	// Verify deleted
	_, err = svc.handleGetBackup(map[string]interface{}{
		"service_name": "test-service",
		"backup_name":  "test-db",
		"version":      float64(1),
	})
	if err == nil {
		t.Error("expected error for deleted version")
	}
}

func TestBackupService_MultipleVersions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "backup-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	svc, err := NewBackupService("test-backup-agent", "nats://localhost:4222", nil, tempDir)
	if err != nil {
		t.Skip("Need running NATS server:", err)
	}
	defer svc.Stop()

	if err := svc.Start(); err != nil {
		t.Fatal(err)
	}

	serviceName := "test-service"
	backupName := "test-db"

	// Create 3 versions
	for i := 1; i <= 3; i++ {
		testData := []byte("test backup data v" + string(rune('0'+i)))
		_, err = svc.handleCreateBackup(map[string]interface{}{
			"service_name": serviceName,
			"backup_name":  backupName,
			"data":         base64.StdEncoding.EncodeToString(testData),
		})
		if err != nil {
			t.Fatal("create backup v", i, "failed:", err)
		}
	}

	// List and verify
	listResult, err := svc.handleListBackups(map[string]interface{}{
		"service_name": serviceName,
		"backup_name":  backupName,
	})
	if err != nil {
		t.Fatal("list backups failed:", err)
	}

	currentVersion, ok := listResult["current_version"].(float64)
	if !ok || int(currentVersion) != 3 {
		t.Error("expected current_version 3, got", currentVersion)
	}

	// Verify each version exists
	for i := 1; i <= 3; i++ {
		_, err := svc.handleGetBackup(map[string]interface{}{
			"service_name": serviceName,
			"backup_name":  backupName,
			"version":      float64(i),
		})
		if err != nil {
			t.Error("get version", i, "failed:", err)
		}
	}
}

func TestBackupService_MetadataPersistence(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "backup-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	svc, err := NewBackupService("test-backup-agent", "nats://localhost:4222", nil, tempDir)
	if err != nil {
		t.Skip("Need running NATS server:", err)
	}

	if err := svc.Start(); err != nil {
		t.Fatal(err)
	}

	serviceName := "test-service"
	backupName := "test-db"

	// Create backup
	testData := []byte("test backup data")
	_, err = svc.handleCreateBackup(map[string]interface{}{
		"service_name": serviceName,
		"backup_name":  backupName,
		"data":         base64.StdEncoding.EncodeToString(testData),
	})
	if err != nil {
		t.Fatal("create backup failed:", err)
	}

	svc.Stop()

	// Verify metadata file exists
	metadataPath := filepath.Join(tempDir, serviceName+"."+backupName, "metadata.json")
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		t.Error("metadata file not created")
	}

	// Load and verify metadata
	metadata, err := svc.loadMetadata(serviceName, backupName)
	if err != nil {
		t.Fatal("load metadata failed:", err)
	}

	if metadata.ServiceName != serviceName {
		t.Error("service_name mismatch")
	}
	if metadata.BackupName != backupName {
		t.Error("backup_name mismatch")
	}
	if metadata.CurrentVersion != 1 {
		t.Error("expected current_version 1, got", metadata.CurrentVersion)
	}
	if len(metadata.Versions) != 1 {
		t.Error("expected 1 version, got", len(metadata.Versions))
	}

	version := metadata.Versions[0]
	if version.Version != 1 {
		t.Error("expected version 1")
	}
	if version.FileSize != int64(len(testData)) {
		t.Error("file_size mismatch")
	}
	if version.Checksum == "" {
		t.Error("checksum empty")
	}
	if version.CreatedAt.IsZero() {
		t.Error("created_at is zero")
	}

	// Verify version file exists
	versionPath := filepath.Join(tempDir, serviceName+"."+backupName, "v1.bin")
	if _, err := os.Stat(versionPath); os.IsNotExist(err) {
		t.Error("version file not created")
	}
}

func TestBackupService_ErrorCases(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "backup-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	svc, err := NewBackupService("test-backup-agent", "nats://localhost:4222", nil, tempDir)
	if err != nil {
		t.Skip("Need running NATS server:", err)
	}
	defer svc.Stop()

	if err := svc.Start(); err != nil {
		t.Fatal(err)
	}

	// Test missing service_name
	_, err = svc.handleCreateBackup(map[string]interface{}{
		"backup_name": "test-db",
		"data":        base64.StdEncoding.EncodeToString([]byte("test")),
	})
	if err == nil {
		t.Error("expected error for missing service_name")
	}

	// Test missing backup_name
	_, err = svc.handleCreateBackup(map[string]interface{}{
		"service_name": "test-service",
		"data":         base64.StdEncoding.EncodeToString([]byte("test")),
	})
	if err == nil {
		t.Error("expected error for missing backup_name")
	}

	// Test invalid base64 data
	_, err = svc.handleCreateBackup(map[string]interface{}{
		"service_name": "test-service",
		"backup_name":  "test-db",
		"data":         "invalid base64!!!",
	})
	if err == nil {
		t.Error("expected error for invalid base64")
	}

	// Test get non-existent version
	_, err = svc.handleGetBackup(map[string]interface{}{
		"service_name": "test-service",
		"backup_name":  "test-db",
		"version":      float64(999),
	})
	if err == nil {
		t.Error("expected error for non-existent version")
	}
}

func TestBackupService_VersionListOrder(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "backup-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	svc, err := NewBackupService("test-backup-agent", "nats://localhost:4222", nil, tempDir)
	if err != nil {
		t.Skip("Need running NATS server:", err)
	}
	defer svc.Stop()

	if err := svc.Start(); err != nil {
		t.Fatal(err)
	}

	serviceName := "test-service"
	backupName := "test-db"

	// Create 5 versions with slight delay to ensure different timestamps
	for i := 1; i <= 5; i++ {
		testData := []byte("test backup data v" + string(rune('0'+i)))
		_, err = svc.handleCreateBackup(map[string]interface{}{
			"service_name": serviceName,
			"backup_name":  backupName,
			"data":         base64.StdEncoding.EncodeToString(testData),
		})
		if err != nil {
			t.Fatal("create backup v", i, "failed:", err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	// List and verify order
	listResult, err := svc.handleListBackups(map[string]interface{}{
		"service_name": serviceName,
		"backup_name":  backupName,
	})
	if err != nil {
		t.Fatal("list backups failed:", err)
	}

	versionsInterface, ok := listResult["versions"].([]interface{})
	if !ok {
		t.Fatal("expected versions array")
	}

	if len(versionsInterface) != 5 {
		t.Error("expected 5 versions, got", len(versionsInterface))
	}

	// Verify versions are in ascending order
	for i, v := range versionsInterface {
		vMap, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		versionFloat, ok := vMap["version"].(float64)
		if !ok || int(versionFloat) != i+1 {
			t.Error("version", i, "has incorrect number:", vMap["version"])
		}
	}
}
