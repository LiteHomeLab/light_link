package service

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/LiteHomeLab/light_link/sdk/go/backup"
	"github.com/LiteHomeLab/light_link/sdk/go/client"
	"github.com/LiteHomeLab/light_link/sdk/go/types"
)

// chunkDownloadState tracks an ongoing chunked download
type chunkDownloadState struct {
	metadata  backup.ChunkMetadata
	data      []byte
	chunkSize int
}

// BackupService manages backup storage and retrieval
type BackupService struct {
	*Service
	storagePath    string
	mu             sync.RWMutex
	uploads        map[string]*chunkUploadState  // transfer_id -> upload state
	downloads      map[string]*chunkDownloadState // transfer_id -> download state
	uploadMu       sync.RWMutex
	downloadMu     sync.RWMutex
}

// chunkUploadState tracks an ongoing chunked upload
type chunkUploadState struct {
	serviceName  string
	backupName   string
	maxVersions  int
	assembler    *backup.ChunkAssembler
	tempFile     string
	createdAt    time.Time
}

// NewBackupService creates a new backup service
func NewBackupService(name, natsURL string, tlsConfig *client.TLSConfig, storagePath string) (*BackupService, error) {
	svc, err := NewService(name, natsURL, WithServiceTLS(tlsConfig))
	if err != nil {
		return nil, err
	}

	// Ensure storage path exists
	if storagePath == "" {
		storagePath = "./backups"
	}
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		svc.Stop()
		return nil, fmt.Errorf("create storage path: %w", err)
	}

	bs := &BackupService{
		Service:     svc,
		storagePath: storagePath,
		uploads:     make(map[string]*chunkUploadState),
		downloads:   make(map[string]*chunkDownloadState),
	}

	// Register RPC handlers
	if err := bs.RegisterRPC("backup.create", bs.handleCreateBackup); err != nil {
		svc.Stop()
		return nil, err
	}
	if err := bs.RegisterRPC("backup.create_incremental", bs.handleCreateIncrementalBackup); err != nil {
		svc.Stop()
		return nil, err
	}
	if err := bs.RegisterRPC("backup.list", bs.handleListBackups); err != nil {
		svc.Stop()
		return nil, err
	}
	if err := bs.RegisterRPC("backup.get", bs.handleGetBackup); err != nil {
		svc.Stop()
		return nil, err
	}
	if err := bs.RegisterRPC("backup.delete", bs.handleDeleteBackup); err != nil {
		svc.Stop()
		return nil, err
	}
	if err := bs.RegisterRPC("backup.cleanup", bs.handleCleanup); err != nil {
		svc.Stop()
		return nil, err
	}
	if err := bs.RegisterRPC("backup.upload_init", bs.handleUploadInit); err != nil {
		svc.Stop()
		return nil, err
	}
	if err := bs.RegisterRPC("backup.upload_chunk", bs.handleUploadChunk); err != nil {
		svc.Stop()
		return nil, err
	}
	if err := bs.RegisterRPC("backup.upload_complete", bs.handleUploadComplete); err != nil {
		svc.Stop()
		return nil, err
	}
	if err := bs.RegisterRPC("backup.download_init", bs.handleDownloadInit); err != nil {
		svc.Stop()
		return nil, err
	}
	if err := bs.RegisterRPC("backup.download_chunk", bs.handleDownloadChunk); err != nil {
		svc.Stop()
		return nil, err
	}

	return bs, nil
}

// getBackupDir returns the directory for a specific backup
func (s *BackupService) getBackupDir(serviceName, backupName string) string {
	return filepath.Join(s.storagePath, fmt.Sprintf("%s.%s", serviceName, backupName))
}

// getMetadataPath returns the path to metadata file
func (s *BackupService) getMetadataPath(serviceName, backupName string) string {
	return filepath.Join(s.getBackupDir(serviceName, backupName), "metadata.json")
}

// getVersionPath returns the path to a version file
func (s *BackupService) getVersionPath(serviceName, backupName string, version int) string {
	return filepath.Join(s.getBackupDir(serviceName, backupName), fmt.Sprintf("v%d.bin", version))
}

// loadMetadata loads backup metadata from disk
func (s *BackupService) loadMetadata(serviceName, backupName string) (*types.BackupMetadata, error) {
	path := s.getMetadataPath(serviceName, backupName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty metadata for new backup
			return &types.BackupMetadata{
				ServiceName:    serviceName,
				BackupName:     backupName,
				CurrentVersion: 0,
				MaxVersions:    0,
				Versions:       []types.BackupVersion{},
			}, nil
		}
		return nil, err
	}

	var metadata types.BackupMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("unmarshal metadata: %w", err)
	}

	return &metadata, nil
}

// saveMetadata saves backup metadata to disk
func (s *BackupService) saveMetadata(metadata *types.BackupMetadata) error {
	path := s.getMetadataPath(metadata.ServiceName, metadata.BackupName)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create backup dir: %w", err)
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write metadata: %w", err)
	}

	return nil
}

// handleCreateBackup handles backup creation requests
func (s *BackupService) handleCreateBackup(args map[string]interface{}) (map[string]interface{}, error) {
	serviceName, ok := args["service_name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing service_name")
	}

	backupName, ok := args["backup_name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing backup_name")
	}

	// Get max_versions parameter
	maxVersions := 0
	if mv, ok := args["max_versions"].(float64); ok {
		maxVersions = int(mv)
	}

	// Decode base64 data
	dataBase64, ok := args["data"].(string)
	if !ok {
		return nil, fmt.Errorf("missing data")
	}

	data, err := base64.StdEncoding.DecodeString(dataBase64)
	if err != nil {
		return nil, fmt.Errorf("decode data: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Load current metadata
	metadata, err := s.loadMetadata(serviceName, backupName)
	if err != nil {
		return nil, fmt.Errorf("load metadata: %w", err)
	}

	// Update max_versions if provided
	if maxVersions > 0 {
		metadata.MaxVersions = maxVersions
	}

	// Increment version
	metadata.CurrentVersion++
	version := metadata.CurrentVersion

	// Calculate checksum
	hash := sha256.Sum256(data)
	checksum := hex.EncodeToString(hash[:])

	// Create version info
	versionInfo := types.BackupVersion{
		Version:   version,
		Type:      "full",
		FileSize:  int64(len(data)),
		Checksum:  checksum,
		CreatedAt: time.Now(),
	}

	// Save data to file
	versionPath := s.getVersionPath(serviceName, backupName, version)
	versionDir := filepath.Dir(versionPath)
	if err := os.MkdirAll(versionDir, 0755); err != nil {
		return nil, fmt.Errorf("create backup dir: %w", err)
	}
	if err := os.WriteFile(versionPath, data, 0644); err != nil {
		return nil, fmt.Errorf("write version file: %w", err)
	}

	// Update metadata
	metadata.Versions = append(metadata.Versions, versionInfo)

	// Save metadata
	if err := s.saveMetadata(metadata); err != nil {
		// Clean up version file if metadata save fails
		os.Remove(versionPath)
		return nil, fmt.Errorf("save metadata: %w", err)
	}

	// Auto cleanup if max_versions is set
	cleanupCount := 0
	if metadata.MaxVersions > 0 && len(metadata.Versions) > metadata.MaxVersions {
		cleanupCount = s.cleanupOldVersionsLocked(metadata)
		// Save metadata after cleanup
		if err := s.saveMetadata(metadata); err != nil {
			// Log but don't fail the backup
			fmt.Printf("Warning: failed to save metadata after cleanup: %v\n", err)
		}
	}

	result := map[string]interface{}{
		"version":  float64(version),
		"size":     versionInfo.FileSize,
		"checksum": checksum,
	}

	if cleanupCount > 0 {
		result["cleaned"] = cleanupCount
	}

	return result, nil
}

// handleCreateIncrementalBackup handles incremental backup creation requests
func (s *BackupService) handleCreateIncrementalBackup(args map[string]interface{}) (map[string]interface{}, error) {
	serviceName, ok := args["service_name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing service_name")
	}

	backupName, ok := args["backup_name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing backup_name")
	}

	// Get max_versions parameter
	maxVersions := 0
	if mv, ok := args["max_versions"].(float64); ok {
		maxVersions = int(mv)
	}

	// Decode base64 data
	dataBase64, ok := args["data"].(string)
	if !ok {
		return nil, fmt.Errorf("missing data")
	}

	data, err := base64.StdEncoding.DecodeString(dataBase64)
	if err != nil {
		return nil, fmt.Errorf("decode data: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Load current metadata
	metadata, err := s.loadMetadata(serviceName, backupName)
	if err != nil {
		return nil, fmt.Errorf("load metadata: %w", err)
	}

	// Check if there's a previous version
	if metadata.CurrentVersion == 0 {
		return nil, fmt.Errorf("no previous backup found, create a full backup first")
	}

	// Get the latest version (or find the latest full version)
	var latestVersionData []byte
	var baseVersion int

	// Try to find the latest full version as base
	for i := len(metadata.Versions) - 1; i >= 0; i-- {
		if metadata.Versions[i].Type == "full" {
			baseVersion = metadata.Versions[i].Version
			// Read the full version file
			versionPath := s.getVersionPath(serviceName, backupName, baseVersion)
			latestVersionData, err = os.ReadFile(versionPath)
			if err != nil {
				return nil, fmt.Errorf("read base version %d: %w", baseVersion, err)
			}
			break
		}
	}

	if latestVersionData == nil {
		return nil, fmt.Errorf("no full backup found to create incremental from")
	}

	// Calculate diff
	diffOps, err := backup.BinaryDiff(latestVersionData, data)
	if err != nil {
		return nil, fmt.Errorf("calculate diff: %w", err)
	}

	// Serialize diff operations
	diffData, err := backup.SerializeDiffOps(diffOps)
	if err != nil {
		return nil, fmt.Errorf("serialize diff: %w", err)
	}

	// Increment version
	metadata.CurrentVersion++
	version := metadata.CurrentVersion

	// Calculate checksum
	hash := sha256.Sum256(diffData)
	checksum := hex.EncodeToString(hash[:])

	// Create version info
	versionInfo := types.BackupVersion{
		Version:     version,
		Type:        "incremental",
		BaseVersion: baseVersion,
		FileSize:    int64(len(diffData)),
		Checksum:    checksum,
		CreatedAt:   time.Now(),
	}

	// Save diff data to file
	versionPath := s.getVersionPath(serviceName, backupName, version)
	versionDir := filepath.Dir(versionPath)
	if err := os.MkdirAll(versionDir, 0755); err != nil {
		return nil, fmt.Errorf("create backup dir: %w", err)
	}
	if err := os.WriteFile(versionPath, diffData, 0644); err != nil {
		return nil, fmt.Errorf("write version file: %w", err)
	}

	// Update metadata
	metadata.Versions = append(metadata.Versions, versionInfo)

	// Update max_versions if provided
	if maxVersions > 0 {
		metadata.MaxVersions = maxVersions
	}

	// Save metadata
	if err := s.saveMetadata(metadata); err != nil {
		// Clean up version file if metadata save fails
		os.Remove(versionPath)
		return nil, fmt.Errorf("save metadata: %w", err)
	}

	// Auto cleanup if max_versions is set
	cleanupCount := 0
	if metadata.MaxVersions > 0 && len(metadata.Versions) > metadata.MaxVersions {
		cleanupCount = s.cleanupOldVersionsLocked(metadata)
		// Save metadata after cleanup
		if err := s.saveMetadata(metadata); err != nil {
			fmt.Printf("Warning: failed to save metadata after cleanup: %v\n", err)
		}
	}

	result := map[string]interface{}{
		"version":      float64(version),
		"size":         versionInfo.FileSize,
		"checksum":     checksum,
		"type":         "incremental",
		"base_version": float64(baseVersion),
	}

	if cleanupCount > 0 {
		result["cleaned"] = cleanupCount
	}

	return result, nil
}

// handleListBackups handles list backup versions requests
func (s *BackupService) handleListBackups(args map[string]interface{}) (map[string]interface{}, error) {
	serviceName, ok := args["service_name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing service_name")
	}

	backupName, ok := args["backup_name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing backup_name")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	metadata, err := s.loadMetadata(serviceName, backupName)
	if err != nil {
		return nil, fmt.Errorf("load metadata: %w", err)
	}

	return map[string]interface{}{
		"current_version": float64(metadata.CurrentVersion),
		"versions":        metadata.Versions,
	}, nil
}

// handleGetBackup handles get backup data requests
func (s *BackupService) handleGetBackup(args map[string]interface{}) (map[string]interface{}, error) {
	serviceName, ok := args["service_name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing service_name")
	}

	backupName, ok := args["backup_name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing backup_name")
	}

	versionFloat, ok := args["version"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing version")
	}
	version := int(versionFloat)

	s.mu.RLock()
	defer s.mu.RUnlock()

	metadata, err := s.loadMetadata(serviceName, backupName)
	if err != nil {
		return nil, fmt.Errorf("load metadata: %w", err)
	}

	// Check if version exists
	if version < 1 || version > metadata.CurrentVersion {
		return nil, fmt.Errorf("version %d not found (current: %d)", version, metadata.CurrentVersion)
	}

	// Read version file
	versionPath := s.getVersionPath(serviceName, backupName, version)
	data, err := os.ReadFile(versionPath)
	if err != nil {
		return nil, fmt.Errorf("read version file: %w", err)
	}

	// Verify checksum
	hash := sha256.Sum256(data)
	checksum := hex.EncodeToString(hash[:])

	// Find version info to compare checksum
	var versionInfo *types.BackupVersion
	for i := range metadata.Versions {
		if metadata.Versions[i].Version == version {
			versionInfo = &metadata.Versions[i]
			break
		}
	}

	if versionInfo != nil && versionInfo.Checksum != checksum {
		return nil, fmt.Errorf("checksum mismatch for version %d", version)
	}

	// Return base64 encoded data
	return map[string]interface{}{
		"data":     base64.StdEncoding.EncodeToString(data),
		"size":     len(data),
		"checksum": checksum,
	}, nil
}

// handleDeleteBackup handles delete backup version requests
func (s *BackupService) handleDeleteBackup(args map[string]interface{}) (map[string]interface{}, error) {
	serviceName, ok := args["service_name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing service_name")
	}

	backupName, ok := args["backup_name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing backup_name")
	}

	versionFloat, ok := args["version"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing version")
	}
	version := int(versionFloat)

	s.mu.Lock()
	defer s.mu.Unlock()

	metadata, err := s.loadMetadata(serviceName, backupName)
	if err != nil {
		return nil, fmt.Errorf("load metadata: %w", err)
	}

	// Check if version exists
	if version < 1 || version > metadata.CurrentVersion {
		return nil, fmt.Errorf("version %d not found (current: %d)", version, metadata.CurrentVersion)
	}

	// Delete version file
	versionPath := s.getVersionPath(serviceName, backupName, version)
	if err := os.Remove(versionPath); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("delete version file: %w", err)
	}

	// Remove version from metadata
	var newVersions []types.BackupVersion
	for _, v := range metadata.Versions {
		if v.Version != version {
			newVersions = append(newVersions, v)
		}
	}
	metadata.Versions = newVersions

	// Save metadata
	if err := s.saveMetadata(metadata); err != nil {
		return nil, fmt.Errorf("save metadata: %w", err)
	}

	return map[string]interface{}{
		"deleted": true,
		"version": float64(version),
	}, nil
}

// handleCleanup handles cleanup old versions requests
func (s *BackupService) handleCleanup(args map[string]interface{}) (map[string]interface{}, error) {
	serviceName, ok := args["service_name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing service_name")
	}

	backupName, ok := args["backup_name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing backup_name")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	metadata, err := s.loadMetadata(serviceName, backupName)
	if err != nil {
		return nil, fmt.Errorf("load metadata: %w", err)
	}

	// If no max_versions set, return without doing anything
	if metadata.MaxVersions <= 0 {
		return map[string]interface{}{
			"cleaned": 0,
			"message": "no max_versions configured",
		}, nil
	}

	cleaned := s.cleanupOldVersionsLocked(metadata)

	// Save metadata after cleanup
	if err := s.saveMetadata(metadata); err != nil {
		return nil, fmt.Errorf("save metadata: %w", err)
	}

	return map[string]interface{}{
		"cleaned": cleaned,
	}, nil
}

// cleanupOldVersionsLocked removes old versions based on MaxVersions setting
// Must be called with mu.Lock() held
func (s *BackupService) cleanupOldVersionsLocked(metadata *types.BackupMetadata) int {
	if metadata.MaxVersions <= 0 {
		return 0
	}

	// Keep the most recent MaxVersions versions
	versionsToKeep := metadata.MaxVersions
	totalVersions := len(metadata.Versions)

	if totalVersions <= versionsToKeep {
		return 0
	}

	cleaned := 0

	// Delete old version files and update metadata
	for i := 0; i < totalVersions-versionsToKeep; i++ {
		version := metadata.Versions[i]
		versionPath := s.getVersionPath(metadata.ServiceName, metadata.BackupName, version.Version)

		// Delete the file
		if err := os.Remove(versionPath); err != nil && !os.IsNotExist(err) {
			// Log but continue
			fmt.Printf("Warning: failed to delete version file %d: %v\n", version.Version, err)
		} else {
			cleaned++
		}
	}

	// Keep only the last MaxVersions versions
	metadata.Versions = metadata.Versions[totalVersions-versionsToKeep:]

	return cleaned
}

// handleUploadInit initializes a chunked upload
func (s *BackupService) handleUploadInit(args map[string]interface{}) (map[string]interface{}, error) {
	serviceName, ok := args["service_name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing service_name")
	}

	backupName, ok := args["backup_name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing backup_name")
	}

	// Get max_versions parameter
	maxVersions := 0
	if mv, ok := args["max_versions"].(float64); ok {
		maxVersions = int(mv)
	}

	// Get metadata
	metadataBase64, ok := args["metadata"].(string)
	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}

	metadataBytes, err := base64.StdEncoding.DecodeString(metadataBase64)
	if err != nil {
		return nil, fmt.Errorf("decode metadata: %w", err)
	}

	metadata, err := backup.DeserializeMetadata(metadataBytes)
	if err != nil {
		return nil, fmt.Errorf("deserialize metadata: %w", err)
	}

	// Generate transfer ID
	transferID := fmt.Sprintf("%s.%s.%d", serviceName, backupName, time.Now().UnixNano())

	// Create temp file for chunks
	tempDir := filepath.Join(s.storagePath, "temp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	tempFile := filepath.Join(tempDir, transferID+".tmp")

	// Create upload state
	state := &chunkUploadState{
		serviceName: serviceName,
		backupName:  backupName,
		maxVersions: maxVersions,
		assembler:   backup.NewChunkAssembler(metadata),
		tempFile:    tempFile,
		createdAt:   time.Now(),
	}

	s.uploadMu.Lock()
	s.uploads[transferID] = state
	s.uploadMu.Unlock()

	return map[string]interface{}{
		"transfer_id": transferID,
		"total_chunks": float64(metadata.TotalChunks),
		"total_size":   float64(metadata.TotalSize),
	}, nil
}

// handleUploadChunk handles a chunk upload
func (s *BackupService) handleUploadChunk(args map[string]interface{}) (map[string]interface{}, error) {
	transferID, ok := args["transfer_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing transfer_id")
	}

	chunkBase64, ok := args["chunk"].(string)
	if !ok {
		return nil, fmt.Errorf("missing chunk")
	}

	chunkBytes, err := base64.StdEncoding.DecodeString(chunkBase64)
	if err != nil {
		return nil, fmt.Errorf("decode chunk: %w", err)
	}

	chunk, err := backup.DeserializeChunk(chunkBytes)
	if err != nil {
		return nil, fmt.Errorf("deserialize chunk: %w", err)
	}

	s.uploadMu.Lock()
	state, ok := s.uploads[transferID]
	s.uploadMu.Unlock()

	if !ok {
		return nil, fmt.Errorf("upload not found: %s", transferID)
	}

	if err := state.assembler.AddChunk(chunk); err != nil {
		return nil, fmt.Errorf("add chunk: %w", err)
	}

	missing := state.assembler.MissingChunks()
	return map[string]interface{}{
		"chunk_index": float64(chunk.Index),
		"received":    float64(len(missing)),
		"complete":    state.assembler.IsComplete(),
	}, nil
}

// handleUploadComplete completes a chunked upload
func (s *BackupService) handleUploadComplete(args map[string]interface{}) (map[string]interface{}, error) {
	transferID, ok := args["transfer_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing transfer_id")
	}

	s.uploadMu.Lock()
	state, ok := s.uploads[transferID]
	if ok {
		delete(s.uploads, transferID)
	}
	s.uploadMu.Unlock()

	if !ok {
		return nil, fmt.Errorf("upload not found: %s", transferID)
	}

	// Assemble all chunks
	data, err := state.assembler.Assemble()
	if err != nil {
		return nil, fmt.Errorf("assemble: %w", err)
	}

	// Clean up temp file
	os.Remove(state.tempFile)

	// Create the backup using the assembled data
	s.mu.Lock()
	defer s.mu.Unlock()

	// Load current metadata
	metadata, err := s.loadMetadata(state.serviceName, state.backupName)
	if err != nil {
		return nil, fmt.Errorf("load metadata: %w", err)
	}

	// Update max_versions if provided
	if state.maxVersions > 0 {
		metadata.MaxVersions = state.maxVersions
	}

	// Increment version
	metadata.CurrentVersion++
	version := metadata.CurrentVersion

	// Calculate checksum
	hash := sha256.Sum256(data)
	checksum := hex.EncodeToString(hash[:])

	// Create version info
	versionInfo := types.BackupVersion{
		Version:   version,
		Type:      "full",
		FileSize:  int64(len(data)),
		Checksum:  checksum,
		CreatedAt: time.Now(),
	}

	// Save data to file
	versionPath := s.getVersionPath(state.serviceName, state.backupName, version)
	versionDir := filepath.Dir(versionPath)
	if err := os.MkdirAll(versionDir, 0755); err != nil {
		return nil, fmt.Errorf("create backup dir: %w", err)
	}
	if err := os.WriteFile(versionPath, data, 0644); err != nil {
		return nil, fmt.Errorf("write version file: %w", err)
	}

	// Update metadata
	metadata.Versions = append(metadata.Versions, versionInfo)

	// Save metadata
	if err := s.saveMetadata(metadata); err != nil {
		os.Remove(versionPath)
		return nil, fmt.Errorf("save metadata: %w", err)
	}

	// Auto cleanup if max_versions is set
	cleanupCount := 0
	if metadata.MaxVersions > 0 && len(metadata.Versions) > metadata.MaxVersions {
		cleanupCount = s.cleanupOldVersionsLocked(metadata)
		if err := s.saveMetadata(metadata); err != nil {
			fmt.Printf("Warning: failed to save metadata after cleanup: %v\n", err)
		}
	}

	result := map[string]interface{}{
		"version":  float64(version),
		"size":     versionInfo.FileSize,
		"checksum": checksum,
	}

	if cleanupCount > 0 {
		result["cleaned"] = cleanupCount
	}

	return result, nil
}

// handleDownloadInit initializes a chunked download
func (s *BackupService) handleDownloadInit(args map[string]interface{}) (map[string]interface{}, error) {
	serviceName, ok := args["service_name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing service_name")
	}

	backupName, ok := args["backup_name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing backup_name")
	}

	versionFloat, ok := args["version"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing version")
	}
	version := int(versionFloat)

	// Get optional chunk_size parameter
	chunkSize := backup.ChunkSize
	if cs, ok := args["chunk_size"].(float64); ok {
		chunkSize = int(cs)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	metadata, err := s.loadMetadata(serviceName, backupName)
	if err != nil {
		return nil, fmt.Errorf("load metadata: %w", err)
	}

	// Check if version exists
	if version < 1 || version > metadata.CurrentVersion {
		return nil, fmt.Errorf("version %d not found (current: %d)", version, metadata.CurrentVersion)
	}

	// Read version file
	versionPath := s.getVersionPath(serviceName, backupName, version)
	data, err := os.ReadFile(versionPath)
	if err != nil {
		return nil, fmt.Errorf("read version file: %w", err)
	}

	// Generate transfer ID
	transferID := fmt.Sprintf("%s.%s.v%d.%d", serviceName, backupName, version, time.Now().UnixNano())

	// Create chunk metadata
	chunkMetadata := backup.ChunkMetadata{
		TotalChunks: uint32((len(data) + chunkSize - 1) / chunkSize),
		TotalSize:   uint64(len(data)),
		FileID:      transferID,
		Checksum:    backup.CalculateChecksum(data),
	}

	// Store download state in memory with cached data and requested chunk size
	downloadState := &chunkDownloadState{
		metadata:  chunkMetadata,
		data:      data,
		chunkSize: chunkSize,
	}

	s.downloadMu.Lock()
	s.downloads[transferID] = downloadState
	s.downloadMu.Unlock()

	// Serialize metadata for response
	metadataBytes, err := backup.SerializeMetadata(chunkMetadata)
	if err != nil {
		return nil, fmt.Errorf("serialize metadata: %w", err)
	}

	return map[string]interface{}{
		"transfer_id":  transferID,
		"total_chunks": float64(chunkMetadata.TotalChunks),
		"total_size":   float64(chunkMetadata.TotalSize),
		"chunk_size":   float64(chunkSize),
		"metadata":     base64.StdEncoding.EncodeToString(metadataBytes),
	}, nil
}

// handleDownloadChunk handles a chunk download request
func (s *BackupService) handleDownloadChunk(args map[string]interface{}) (map[string]interface{}, error) {
	transferID, ok := args["transfer_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing transfer_id")
	}

	chunkIndexFloat, ok := args["chunk_index"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing chunk_index")
	}
	chunkIndex := uint32(chunkIndexFloat)

	s.downloadMu.Lock()
	state, ok := s.downloads[transferID]
	s.downloadMu.Unlock()

	if !ok {
		return nil, fmt.Errorf("download not found: %s", transferID)
	}

	data := state.data
	chunkSize := state.chunkSize

	// Extract the requested chunk
	start := int(chunkIndex) * chunkSize
	end := start + chunkSize
	if end > len(data) {
		end = len(data)
	}

	if start >= len(data) {
		return nil, fmt.Errorf("chunk index out of range")
	}

	chunkData := data[start:end]
	chunk := backup.Chunk{
		Index:    chunkIndex,
		Data:     chunkData,
		Checksum: backup.CalculateChecksum(chunkData),
		Size:     uint32(len(chunkData)),
	}

	// Serialize chunk
	chunkBytes, err := backup.SerializeChunk(chunk)
	if err != nil {
		return nil, fmt.Errorf("serialize chunk: %w", err)
	}

	return map[string]interface{}{
		"chunk": base64.StdEncoding.EncodeToString(chunkBytes),
		"index": float64(chunk.Index),
		"size":  float64(chunk.Size),
	}, nil
}
