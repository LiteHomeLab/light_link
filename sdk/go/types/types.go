package types

import (
	"fmt"
	"os"
	"path/filepath"
)

// RPC 请求
type RPCRequest struct {
    ID     string                 `json:"id"`
    Method string                 `json:"method"`
    Args   map[string]interface{} `json:"args"`
}

// RPC 响应
type RPCResponse struct {
    ID      string                 `json:"id"`
    Success bool                   `json:"success"`
    Result  map[string]interface{} `json:"result,omitempty"`
    Error   string                 `json:"error,omitempty"`
}

// 消息
type Message struct {
    Subject string                 `json:"subject"`
    Data    map[string]interface{} `json:"data"`
}

// 状态条目
type StateEntry struct {
    Key       string                 `json:"key"`
    Value     map[string]interface{} `json:"value"`
    Revision  uint64                 `json:"revision"`
    Timestamp int64                  `json:"timestamp"`
}

// 文件元数据
type FileMetadata struct {
    FileID   string `json:"file_id"`
    FileName string `json:"file_name"`
    FileSize int64  `json:"file_size"`
    ChunkNum int    `json:"chunk_num"`
    From     string `json:"from"`
    To       string `json:"to"`
}

// 配置
type Config struct {
    NATSURL     string     `json:"nats_url"`
    ServiceName string     `json:"service_name"`
    TLS         *TLSConfig `json:"tls,omitempty"`
}

type TLSConfig struct {
    CaFile     string `json:"ca_file"`
    CertFile   string `json:"cert_file"`
    KeyFile    string `json:"key_file"`
    ServerName string `json:"server_name,omitempty"`
}

// ========================================
// Certificate Auto-Discovery
// ========================================

// 证书发现相关常量
const (
	// DefaultClientCertDir 默认客户端证书目录
	DefaultClientCertDir = "./client"
	// DefaultServerCertDir 默认服务器证书目录
	DefaultServerCertDir = "./nats-server"
	// DefaultServerName 默认服务器名称
	DefaultServerName = "nats-server"
)

// CertDiscoveryResult 证书发现结果
type CertDiscoveryResult struct {
	CaFile     string
	CertFile   string
	KeyFile    string
	ServerName string
	Found      bool
}

// DiscoverClientCerts 在默认位置自动发现客户端证书
// 搜索顺序: ./client -> ../client -> ../../client
func DiscoverClientCerts() (*CertDiscoveryResult, error) {
	searchPaths := []string{
		DefaultClientCertDir,
		"../client",
		"../../client",
	}

	for _, basePath := range searchPaths {
		result := checkCertDirectory(basePath, "client")
		if result.Found {
			return result, nil
		}
	}

	return nil, fmt.Errorf("client certificates not found in search paths: %v", searchPaths)
}

// DiscoverServerCerts 在默认位置自动发现服务器证书
// 搜索顺序: ./nats-server -> ../nats-server -> ../../nats-server
func DiscoverServerCerts() (*CertDiscoveryResult, error) {
	searchPaths := []string{
		DefaultServerCertDir,
		"../nats-server",
		"../../nats-server",
	}

	for _, basePath := range searchPaths {
		result := checkCertDirectory(basePath, "server")
		if result.Found {
			return result, nil
		}
	}

	return nil, fmt.Errorf("server certificates not found in search paths: %v", searchPaths)
}

// checkCertDirectory 检查目录中的证书文件是否存在
func checkCertDirectory(dir, certType string) *CertDiscoveryResult {
	var certFile, keyFile string

	if certType == "client" {
		certFile = filepath.Join(dir, "client.crt")
		keyFile = filepath.Join(dir, "client.key")
	} else {
		certFile = filepath.Join(dir, "server.crt")
		keyFile = filepath.Join(dir, "server.key")
	}

	caFile := filepath.Join(dir, "ca.crt")

	// 检查所有文件是否存在
	if fileExists(caFile) && fileExists(certFile) && fileExists(keyFile) {
		return &CertDiscoveryResult{
			CaFile:     caFile,
			CertFile:   certFile,
			KeyFile:    keyFile,
			ServerName: DefaultServerName,
			Found:      true,
		}
	}

	return &CertDiscoveryResult{Found: false}
}

// fileExists 检查文件是否存在
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// CertDiscoveryResultToTLSConfig 将发现结果转换为 TLSConfig
func CertDiscoveryResultToTLSConfig(result *CertDiscoveryResult) *TLSConfig {
	return &TLSConfig{
		CaFile:     result.CaFile,
		CertFile:   result.CertFile,
		KeyFile:    result.KeyFile,
		ServerName: result.ServerName,
	}
}
