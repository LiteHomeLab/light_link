package client

import (
	"encoding/base64"
	"fmt"

	"github.com/LiteHomeLab/light_link/sdk/go/backup"
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

// CreateBackupWithMaxVersions creates a full backup with version retention policy
func (c *Client) CreateBackupWithMaxVersions(serviceName, backupName string, data []byte, maxVersions int) (int, error) {
	args := map[string]interface{}{
		"service_name": serviceName,
		"backup_name":  backupName,
		"data":         base64.StdEncoding.EncodeToString(data),
		"max_versions": float64(maxVersions),
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

// CreateIncrementalBackup creates an incremental backup
func (c *Client) CreateIncrementalBackup(serviceName, backupName string, data []byte) (int, error) {
	args := map[string]interface{}{
		"service_name": serviceName,
		"backup_name":  backupName,
		"data":         base64.StdEncoding.EncodeToString(data),
	}

	result, err := c.Call("backup-agent", "backup.create_incremental", args)
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
		if val, ok := vMap["type"].(string); ok {
			version.Type = val
		}
		if val, ok := vMap["base_version"].(float64); ok {
			version.BaseVersion = int(val)
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

// CleanupOldVersions cleans up old versions based on retention policy
func (c *Client) CleanupOldVersions(serviceName, backupName string) error {
	args := map[string]interface{}{
		"service_name": serviceName,
		"backup_name":  backupName,
	}

	_, err := c.Call("backup-agent", "backup.cleanup", args)
	return err
}

// ChunkedUploadHandle represents an ongoing chunked upload
type ChunkedUploadHandle struct {
	client      *Client
	transferID  string
	totalChunks int
	Splitter    *ChunkSplitter
}

// ChunkSplitter wraps the backup.ChunkSplitter for client use
type ChunkSplitter struct {
	splitter *backup.ChunkSplitter
}

// NewChunkSplitter creates a new chunk splitter
func NewChunkSplitter(data []byte) *ChunkSplitter {
	return &ChunkSplitter{
		splitter: backup.NewChunkSplitter(data),
	}
}

// SetFileID sets the file ID for the transfer
func (cs *ChunkSplitter) SetFileID(id string) {
	cs.splitter.SetFileID(id)
}

// SetChunkSize sets custom chunk size
func (cs *ChunkSplitter) SetChunkSize(size int) {
	cs.splitter.SetChunkSize(size)
}

// Metadata returns the chunk metadata
func (cs *ChunkSplitter) Metadata() backup.ChunkMetadata {
	return cs.splitter.Metadata()
}

// SplitAll splits all data into chunks
func (cs *ChunkSplitter) SplitAll() ([]backup.Chunk, error) {
	return cs.splitter.SplitAll()
}

// UploadChunked starts a chunked upload
func (c *Client) UploadChunked(serviceName, backupName string, data []byte) (*ChunkedUploadHandle, error) {
	splitter := NewChunkSplitter(data)
	return c.UploadChunkedWithSplitter(serviceName, backupName, splitter)
}

// UploadChunkedWithSplitter starts a chunked upload with a pre-configured splitter
func (c *Client) UploadChunkedWithSplitter(serviceName, backupName string, splitter *ChunkSplitter) (*ChunkedUploadHandle, error) {
	metadata := splitter.Metadata()
	metadataBytes, err := backup.SerializeMetadata(metadata)
	if err != nil {
		return nil, fmt.Errorf("serialize metadata: %w", err)
	}

	args := map[string]interface{}{
		"service_name": serviceName,
		"backup_name":  backupName,
		"metadata":     base64.StdEncoding.EncodeToString(metadataBytes),
	}

	result, err := c.Call("backup-agent", "backup.upload_init", args)
	if err != nil {
		return nil, err
	}

	transferID, ok := result["transfer_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid transfer_id response")
	}

	totalChunksFloat, ok := result["total_chunks"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid total_chunks response")
	}

	return &ChunkedUploadHandle{
		client:      c,
		transferID:  transferID,
		totalChunks: int(totalChunksFloat),
		Splitter:    splitter,
	}, nil
}

// UploadNextChunk uploads the next chunk
func (h *ChunkedUploadHandle) UploadNextChunk(chunk backup.Chunk) error {
	chunkBytes, err := backup.SerializeChunk(chunk)
	if err != nil {
		return fmt.Errorf("serialize chunk: %w", err)
	}

	args := map[string]interface{}{
		"transfer_id": h.transferID,
		"chunk":       base64.StdEncoding.EncodeToString(chunkBytes),
	}

	_, err = h.client.Call("backup-agent", "backup.upload_chunk", args)
	return err
}

// UploadAll uploads all chunks
func (h *ChunkedUploadHandle) UploadAll() error {
	chunks, err := h.Splitter.SplitAll()
	if err != nil {
		return fmt.Errorf("split chunks: %w", err)
	}

	for _, chunk := range chunks {
		if err := h.UploadNextChunk(chunk); err != nil {
			return err
		}
	}

	return nil
}

// Complete completes the chunked upload
func (h *ChunkedUploadHandle) Complete() (int, error) {
	args := map[string]interface{}{
		"transfer_id": h.transferID,
	}

	result, err := h.client.Call("backup-agent", "backup.upload_complete", args)
	if err != nil {
		return 0, err
	}

	versionFloat, ok := result["version"].(float64)
	if !ok {
		return 0, fmt.Errorf("invalid version response")
	}

	return int(versionFloat), nil
}

// UploadChunkedComplete uploads data in chunks and completes the transfer
func (c *Client) UploadChunkedComplete(serviceName, backupName string, data []byte) (int, error) {
	handle, err := c.UploadChunked(serviceName, backupName, data)
	if err != nil {
		return 0, err
	}

	if err := handle.UploadAll(); err != nil {
		return 0, err
	}

	return handle.Complete()
}

// ChunkedDownloadHandle represents an ongoing chunked download
type ChunkedDownloadHandle struct {
	client      *Client
	transferID  string
	TotalChunks int
	metadata    backup.ChunkMetadata
}

// DownloadChunked starts a chunked download
func (c *Client) DownloadChunked(serviceName, backupName string, version int) (*ChunkedDownloadHandle, error) {
	return c.DownloadChunkedWithSize(serviceName, backupName, version, backup.ChunkSize)
}

// DownloadChunkedWithSize starts a chunked download with custom chunk size
func (c *Client) DownloadChunkedWithSize(serviceName, backupName string, version int, chunkSize int) (*ChunkedDownloadHandle, error) {
	args := map[string]interface{}{
		"service_name": serviceName,
		"backup_name":  backupName,
		"version":      float64(version),
		"chunk_size":   float64(chunkSize),
	}

	result, err := c.Call("backup-agent", "backup.download_init", args)
	if err != nil {
		return nil, err
	}

	transferID, ok := result["transfer_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid transfer_id response")
	}

	totalChunksFloat, ok := result["total_chunks"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid total_chunks response")
	}

	metadataBase64, ok := result["metadata"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid metadata response")
	}

	metadataBytes, err := base64.StdEncoding.DecodeString(metadataBase64)
	if err != nil {
		return nil, fmt.Errorf("decode metadata: %w", err)
	}

	metadata, err := backup.DeserializeMetadata(metadataBytes)
	if err != nil {
		return nil, fmt.Errorf("deserialize metadata: %w", err)
	}

	return &ChunkedDownloadHandle{
		client:      c,
		transferID:  transferID,
		TotalChunks: int(totalChunksFloat),
		metadata:    metadata,
	}, nil
}

// DownloadChunk downloads a specific chunk
func (h *ChunkedDownloadHandle) DownloadChunk(chunkIndex int) (*backup.Chunk, error) {
	args := map[string]interface{}{
		"transfer_id":  h.transferID,
		"chunk_index":  float64(chunkIndex),
	}

	result, err := h.client.Call("backup-agent", "backup.download_chunk", args)
	if err != nil {
		return nil, err
	}

	chunkBase64, ok := result["chunk"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid chunk response")
	}

	chunkBytes, err := base64.StdEncoding.DecodeString(chunkBase64)
	if err != nil {
		return nil, fmt.Errorf("decode chunk: %w", err)
	}

	chunk, err := backup.DeserializeChunk(chunkBytes)
	if err != nil {
		return nil, fmt.Errorf("deserialize chunk: %w", err)
	}

	return &chunk, nil
}

// DownloadAll downloads all chunks and assembles the data
func (h *ChunkedDownloadHandle) DownloadAll() ([]byte, error) {
	assembler := backup.NewChunkAssembler(h.metadata)

	for i := 0; i < int(h.metadata.TotalChunks); i++ {
		chunk, err := h.DownloadChunk(i)
		if err != nil {
			return nil, fmt.Errorf("download chunk %d: %w", i, err)
		}

		if err := assembler.AddChunk(*chunk); err != nil {
			return nil, fmt.Errorf("add chunk %d: %w", i, err)
		}
	}

	return assembler.Assemble()
}

// DownloadChunkedComplete downloads a backup in chunks and returns the data
func (c *Client) DownloadChunkedComplete(serviceName, backupName string, version int) ([]byte, error) {
	handle, err := c.DownloadChunked(serviceName, backupName, version)
	if err != nil {
		return nil, err
	}

	return handle.DownloadAll()
}
