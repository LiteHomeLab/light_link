package types

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
