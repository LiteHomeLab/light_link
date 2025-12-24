package types

import "time"

// ServiceMetadata 服务元数据
type ServiceMetadata struct {
	Name         string           `json:"name"`
	Version      string           `json:"version"`
	Description  string           `json:"description"`
	Author       string           `json:"author"`
	Tags         []string         `json:"tags"`
	Methods      []MethodMetadata `json:"methods"`
	RegisteredAt time.Time        `json:"registeredAt"`
	LastSeen     time.Time        `json:"lastSeen"`
}

// MethodMetadata 方法元数据
type MethodMetadata struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Params      []ParameterMetadata  `json:"params"`
	Returns     []ReturnMetadata     `json:"returns"`
	Example     *ExampleMetadata     `json:"example,omitempty"`
	Tags        []string             `json:"tags"`
	Deprecated  bool                 `json:"deprecated"`
}

// ParameterMetadata 参数元数据
type ParameterMetadata struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // string, number, boolean, array, object
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Default     any    `json:"default,omitempty"`
}

// ReturnMetadata 返回值元数据
type ReturnMetadata struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// ExampleMetadata 示例元数据
type ExampleMetadata struct {
	Input       map[string]any `json:"input"`
	Output      map[string]any `json:"output"`
	Description string         `json:"description"`
}

// RegisterMessage 注册消息
type RegisterMessage struct {
	Service   string          `json:"service"`
	Version   string          `json:"version"`
	Metadata  ServiceMetadata `json:"metadata"`
	Timestamp time.Time       `json:"timestamp"`
}

// HeartbeatMessage 心跳消息
type HeartbeatMessage struct {
	Service   string    `json:"service"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

// ServiceEvent 服务事件
type ServiceEvent struct {
	Type      string    `json:"type"`      // online, offline, registered, updated
	Service   string    `json:"service"`
	Method    string    `json:"method,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Data      any       `json:"data,omitempty"`
}

// ServiceStatus 服务状态
type ServiceStatus struct {
	Name     string    `json:"name"`
	Online   bool      `json:"online"`
	LastSeen time.Time `json:"lastSeen"`
	Version  string    `json:"version"`
}
