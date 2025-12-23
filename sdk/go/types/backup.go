package types

import "time"

// BackupCreateRequest - 创建备份请求
type BackupCreateRequest struct {
	ServiceName string `json:"service_name"`
	BackupName  string `json:"backup_name"`
	Data        []byte `json:"data"` // Base64 编码
	MaxVersions int    `json:"max_versions"` // 最大版本数，0表示不限制
}

// BackupVersion - 备份版本信息
type BackupVersion struct {
	Version    int       `json:"version"`
	Type       string    `json:"type"` // "full" or "incremental"
	BaseVersion int      `json:"base_version,omitempty"` // For incremental: base version
	FileSize   int64     `json:"file_size"`
	Checksum   string    `json:"checksum"` // SHA256 hex
	CreatedAt  time.Time `json:"created_at"`
}

// BackupMetadata - 备份元数据
type BackupMetadata struct {
	ServiceName    string          `json:"service_name"`
	BackupName     string          `json:"backup_name"`
	CurrentVersion int             `json:"current_version"`
	MaxVersions    int             `json:"max_versions"` // 最大版本数，0表示不限制
	Versions       []BackupVersion `json:"versions"`
}
