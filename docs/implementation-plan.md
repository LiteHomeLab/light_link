# LightLink 服务管理平台实施计划

## 概述

本文档提供了服务管理平台的详细实施计划，分为 6 个阶段，每个阶段包含具体的任务列表、验收标准和依赖关系。

---

## Phase 1: 元数据系统和 NATS 消息发送

### 目标
实现 SDK 端的元数据定义、注册和通过 NATS 发送注册/心跳消息的能力。

### 任务清单

#### 1.1 元数据类型定义
**文件**: `sdk/go/types/metadata.go`

```go
package types

import "time"

// ServiceMetadata 服务元数据
type ServiceMetadata struct {
    Name        string           `json:"name"`
    Version     string           `json:"version"`
    Description string           `json:"description"`
    Author      string           `json:"author"`
    Tags        []string         `json:"tags"`
    Methods     []MethodMetadata `json:"methods"`
    RegisteredAt time.Time       `json:"registeredAt"`
    LastSeen    time.Time        `json:"lastSeen"`
}

// MethodMetadata 方法元数据
type MethodMetadata struct {
    Name        string               `json:"name"`
    Description string               `json:"description"`
    Params      []ParameterMetadata  `json:"params"`
    Returns     []ReturnMetadata     `json:"returns"`
    Example     *ExampleMetadata     `json:"example,omitempty"`
    Tags        []string             `json:"tags"`
    Deprecated bool                 `json:"deprecated"`
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
    Service   string           `json:"service"`
    Version   string           `json:"version"`
    Metadata  ServiceMetadata  `json:"metadata"`
    Timestamp time.Time        `json:"timestamp"`
}

// HeartbeatMessage 心跳消息
type HeartbeatMessage struct {
    Service   string    `json:"service"`
    Version   string    `json:"version"`
    Timestamp time.Time `json:"timestamp"`
}
```

**验收标准**:
- [ ] 所有类型定义完整
- [ ] JSON 标签正确
- [ ] 通过 `go build` 编译

---

#### 1.2 服务元数据注册和发送
**文件**: `sdk/go/service/metadata.go`

```go
package service

import (
    "encoding/json"
    "fmt"
    "github.com/LiteHomeLab/light_link/sdk/go/types"
    "time"
)

// RegisterMetadata 注册服务元数据并发送到 NATS
func (s *Service) RegisterMetadata(metadata *types.ServiceMetadata) error {
    s.metadata = metadata

    // 发送注册消息
    msg := types.RegisterMessage{
        Service:   s.name,
        Version:   metadata.Version,
        Metadata:  *metadata,
        Timestamp: time.Now(),
    }

    data, err := json.Marshal(msg)
    if err != nil {
        return fmt.Errorf("marshal metadata: %w", err)
    }

    subject := fmt.Sprintf("$LL.register.%s", s.name)
    if err := s.nc.Publish(subject, data); err != nil {
        return fmt.Errorf("publish metadata: %w", err)
    }

    return nil
}

// GetMetadata 获取服务元数据
func (s *Service) GetMetadata() *types.ServiceMetadata {
    return s.metadata
}
```

**验收标准**:
- [ ] `RegisterMetadata` 方法正确发送消息到 `$LL.register.<service>`
- [ ] 元数据存储在 Service 实例中
- [ ] 单元测试覆盖正常和错误情况

---

#### 1.3 方法注册增强
**文件**: `sdk/go/service/service.go`

```go
// 在 Service 结构体中添加
type Service struct {
    // 现有字段...
    metadata    *types.ServiceMetadata
    methodsMeta map[string]*types.MethodMetadata // 方法元数据映射
}

// RegisterMethod 注册方法及元数据
func (s *Service) RegisterMethod(
    name string,
    handler func(map[string]interface{}) (interface{}, error),
    metadata *types.MethodMetadata,
) error {
    // 存储方法元数据
    if s.methodsMeta == nil {
        s.methodsMeta = make(map[string]*types.MethodMetadata)
    }
    s.methodsMeta[name] = metadata

    // 调用原有的 RegisterRPC
    return s.RegisterRPC(name, handler)
}

// GetMethodMetadata 获取方法元数据
func (s *Service) GetMethodMetadata(name string) (*types.MethodMetadata, bool) {
    if s.methodsMeta == nil {
        return nil, false
    }
    meta, ok := s.methodsMeta[name]
    return meta, ok
}
```

**验收标准**:
- [ ] 方法元数据正确存储
- [ ] `GetMethodMetadata` 可以查询元数据
- [ ] 不影响现有的 `RegisterRPC` 方法

---

#### 1.4 心跳发送机制
**文件**: `sdk/go/service/heartbeat.go`

```go
package service

import (
    "encoding/json"
    "fmt"
    "github.com/LiteHomeLab/light_link/sdk/go/types"
    "time"
)

const (
    DefaultHeartbeatInterval = 30 * time.Second
)

// startHeartbeat 启动心跳发送
func (s *Service) startHeartbeat() error {
    subject := fmt.Sprintf("$LL.heartbeat.%s", s.name)

    // 创建定时器
    ticker := time.NewTicker(DefaultHeartbeatInterval)

    go func() {
        for range ticker.C {
            s.sendHeartbeat(subject)
        }
    }()

    // 发送首次心跳
    return s.sendHeartbeat(subject)
}

// sendHeartbeat 发送单次心跳
func (s *Service) sendHeartbeat(subject string) error {
    version := "unknown"
    if s.metadata != nil {
        version = s.metadata.Version
    }

    msg := types.HeartbeatMessage{
        Service:   s.name,
        Version:   version,
        Timestamp: time.Now(),
    }

    data, err := json.Marshal(msg)
    if err != nil {
        return fmt.Errorf("marshal heartbeat: %w", err)
    }

    return s.nc.Publish(subject, data)
}
```

**验收标准**:
- [ ] 心跳每 30 秒发送一次
- [ ] 服务启动时立即发送首次心跳
- [ ] 单元测试验证心跳消息格式

---

#### 1.5 Service.Start() 增强
**文件**: `sdk/go/service/service.go`

```go
// Start 启动服务
func (s *Service) Start() error {
    // 启动心跳
    if err := s.startHeartbeat(); err != nil {
        return fmt.Errorf("start heartbeat: %w", err)
    }

    // 原有的启动逻辑
    // ...

    return nil
}
```

**验收标准**:
- [ ] 调用 `Start()` 后自动发送心跳
- [ ] 不影响现有功能

---

#### 1.6 单元测试
**文件**: `sdk/go/service/metadata_test.go`

```go
package service

import (
    "testing"
    "github.com/LiteHomeLab/light_link/sdk/go/types"
    "github.com/nats-io/nats.go"
)

func TestRegisterMetadata(t *testing.T) {
    // 测试元数据注册
}

func TestHeartbeat(t *testing.T) {
    // 测试心跳发送
}

func TestMethodWithMetadata(t *testing.T) {
    // 测试带元数据的方法注册
}
```

**验收标准**:
- [ ] 测试覆盖率 > 80%
- [ ] 所有测试通过
- [ ] `go test ./sdk/go/service/...` 成功

---

### Phase 1 验收标准
- [ ] 元数据类型定义完整
- [ ] 服务可以发送注册消息到 NATS
- [ ] 服务可以发送心跳消息到 NATS
- [ ] 单元测试全部通过
- [ ] 不影响现有功能

---

## Phase 2: Service Manager 和 SQLite 存储

### 目标
实现 Web 控制台后端的服务管理器，监听 NATS 消息并存储到 SQLite。

### 任务清单

#### 2.1 SQLite 数据库初始化
**文件**: `console/server/storage/db.go`

```go
package storage

import (
    "database/sql"
    "fmt"

    _ "github.com/mattn/go-sqlite3"
)

type Database struct {
    db *sql.DB
}

func NewDatabase(path string) (*Database, error) {
    db, err := sql.Open("sqlite3", path)
    if err != nil {
        return nil, fmt.Errorf("open database: %w", err)
    }

    // 启用外键约束
    if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
        return nil, fmt.Errorf("enable foreign keys: %w", err)
    }

    d := &Database{db: db}
    if err := d.init(); err != nil {
        return nil, err
    }

    return d, nil
}

func (d *Database) init() error {
    schema := `
    CREATE TABLE IF NOT EXISTS services (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL UNIQUE,
        version TEXT,
        description TEXT,
        author TEXT,
        tags TEXT,
        registered_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS methods (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        service_id INTEGER NOT NULL,
        name TEXT NOT NULL,
        description TEXT,
        params TEXT,
        returns TEXT,
        example TEXT,
        tags TEXT,
        deprecated BOOLEAN DEFAULT 0,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE,
        UNIQUE(service_id, name)
    );

    CREATE TABLE IF NOT EXISTS service_status (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        service_id INTEGER NOT NULL,
        online BOOLEAN NOT NULL DEFAULT 0,
        last_seen DATETIME,
        version TEXT,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS service_status_history (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        service_id INTEGER NOT NULL,
        online BOOLEAN NOT NULL,
        timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS events (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        type TEXT NOT NULL,
        service TEXT,
        method TEXT,
        data TEXT,
        timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS call_history (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        service_id INTEGER NOT NULL,
        method_id INTEGER NOT NULL,
        service_name TEXT NOT NULL,
        method_name TEXT NOT NULL,
        input TEXT,
        output TEXT,
        success BOOLEAN NOT NULL,
        error TEXT,
        duration_ms INTEGER,
        called_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE,
        FOREIGN KEY (method_id) REFERENCES methods(id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL UNIQUE,
        password_hash TEXT NOT NULL,
        role TEXT NOT NULL DEFAULT 'viewer',
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );

    -- 索引
    CREATE INDEX IF NOT EXISTS idx_service_status_service_id ON service_status(service_id);
    CREATE INDEX IF NOT EXISTS idx_service_status_history_service_id ON service_status_history(service_id);
    CREATE INDEX IF NOT EXISTS idx_service_status_history_timestamp ON service_status_history(timestamp);
    CREATE INDEX IF NOT EXISTS idx_events_type ON events(type);
    CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp);
    CREATE INDEX IF NOT EXISTS idx_call_history_service_id ON call_history(service_id);
    CREATE INDEX IF NOT EXISTS idx_call_history_called_at ON call_history(called_at);
    `

    _, err := d.db.Exec(schema)
    return err
}

func (d *Database) Close() error {
    return d.db.Close()
}
```

**验收标准**:
- [ ] 数据库初始化成功
- [ ] 所有表创建正确
- [ ] 外键约束生效

---

#### 2.2 服务存储操作
**文件**: `console/server/storage/service.go`

```go
package storage

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "time"
)

type ServiceMetadata struct {
    ID          int64     `db:"id"`
    Name        string    `db:"name"`
    Version     string    `db:"version"`
    Description string    `db:"description"`
    Author      string    `db:"author"`
    Tags        []string  `db:"tags"`
    RegisteredAt time.Time `db:"registered_at"`
    UpdatedAt   time.Time `db:"updated_at"`
}

// SaveService 保存或更新服务
func (d *Database) SaveService(meta *ServiceMetadata) error {
    tagsJSON, _ := json.Marshal(meta.Tags)

    query := `
    INSERT INTO services (name, version, description, author, tags, registered_at, updated_at)
    VALUES (?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT(name) DO UPDATE SET
        version = excluded.version,
        description = excluded.description,
        author = excluded.author,
        tags = excluded.tags,
        updated_at = excluded.updated_at
    `

    now := time.Now()
    meta.RegisteredAt = now
    meta.UpdatedAt = now

    _, err := d.db.Exec(query, meta.Name, meta.Version, meta.Description,
        meta.Author, string(tagsJSON), now, now)
    return err
}

// GetService 获取服务
func (d *Database) GetService(name string) (*ServiceMetadata, error) {
    query := `SELECT id, name, version, description, author, tags, registered_at, updated_at
              FROM services WHERE name = ?`

    row := d.db.QueryRow(query, name)

    var s ServiceMetadata
    var tagsJSON string
    err := row.Scan(&s.ID, &s.Name, &s.Version, &s.Description,
        &s.Author, &tagsJSON, &s.RegisteredAt, &s.UpdatedAt)
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("service not found")
    }
    if err != nil {
        return nil, err
    }

    json.Unmarshal([]byte(tagsJSON), &s.Tags)
    return &s, nil
}

// ListServices 列出所有服务
func (d *Database) ListServices() ([]*ServiceMetadata, error) {
    query := `SELECT id, name, version, description, author, tags, registered_at, updated_at
              FROM services ORDER BY registered_at DESC`

    rows, err := d.db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var services []*ServiceMetadata
    for rows.Next() {
        var s ServiceMetadata
        var tagsJSON string
        if err := rows.Scan(&s.ID, &s.Name, &s.Version, &s.Description,
            &s.Author, &tagsJSON, &s.RegisteredAt, &s.UpdatedAt); err != nil {
            return nil, err
        }
        json.Unmarshal([]byte(tagsJSON), &s.Tags)
        services = append(services, &s)
    }

    return services, rows.Err()
}

// DeleteService 删除服务
func (d *Database) DeleteService(name string) error {
    _, err := d.db.Exec("DELETE FROM services WHERE name = ?", name)
    return err
}

// GetServiceID 获取服务ID
func (d *Database) GetServiceID(name string) (int64, error) {
    var id int64
    err := d.db.QueryRow("SELECT id FROM services WHERE name = ?", name).Scan(&id)
    return id, err
}
```

**验收标准**:
- [ ] 服务可以保存和更新
- [ ] 服务可以查询
- [ ] 服务列表正确返回
- [ ] 删除服务级联删除相关数据

---

#### 2.3 方法存储操作
**文件**: `console/server/storage/method.go`

```go
package storage

import (
    "database/sql"
    "encoding/json"
)

type MethodMetadata struct {
    ID          int64     `db:"id"`
    ServiceID   int64     `db:"service_id"`
    Name        string    `db:"name"`
    Description string    `db:"description"`
    Params      []ParameterMetadata `db:"params"`
    Returns     []ReturnMetadata   `db:"returns"`
    Example     *ExampleMetadata   `db:"example"`
    Tags        []string  `db:"tags"`
    Deprecated  bool      `db:"deprecated"`
    CreatedAt   time.Time `db:"created_at"`
}

// SaveMethod 保存或更新方法
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

// GetMethods 获取服务的所有方法
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

    var methods []*MethodMetadata
    for rows.Next() {
        var m MethodMetadata
        var paramsJSON, returnsJSON, exampleJSON, tagsJSON string
        if err := rows.Scan(&m.ID, &m.ServiceID, &m.Name, &m.Description,
            &paramsJSON, &returnsJSON, &exampleJSON, &tagsJSON,
            &m.Deprecated, &m.CreatedAt); err != nil {
            return nil, err
        }
        json.Unmarshal([]byte(paramsJSON), &m.Params)
        json.Unmarshal([]byte(returnsJSON), &m.Returns)
        json.Unmarshal([]byte(exampleJSON), &m.Example)
        json.Unmarshal([]byte(tagsJSON), &m.Tags)
        methods = append(methods, &m)
    }

    return methods, rows.Err()
}

// GetMethod 获取指定方法
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
        return nil, fmt.Errorf("method not found")
    }
    if err != nil {
        return nil, err
    }

    json.Unmarshal([]byte(paramsJSON), &m.Params)
    json.Unmarshal([]byte(returnsJSON), &m.Returns)
    json.Unmarshal([]byte(exampleJSON), &m.Example)
    json.Unmarshal([]byte(tagsJSON), &m.Tags)
    return &m, nil
}

// GetMethodID 获取方法ID
func (d *Database) GetMethodID(serviceID int64, methodName string) (int64, error) {
    var id int64
    err := d.db.QueryRow("SELECT id FROM methods WHERE service_id = ? AND name = ?",
        serviceID, methodName).Scan(&id)
    return id, err
}
```

**验收标准**:
- [ ] 方法可以保存和更新
- [ ] 方法可以按服务查询
- [ ] 单个方法可以查询

---

#### 2.4 状态存储操作
**文件**: `console/server/storage/status.go`

```go
package storage

import (
    "database/sql"
    "time"
)

type ServiceStatus struct {
    ID        int64     `db:"id"`
    ServiceID int64     `db:"service_id"`
    ServiceName string  `db:"service_name"`
    Online    bool      `db:"online"`
    LastSeen  time.Time `db:"last_seen"`
    Version   string    `db:"version"`
    UpdatedAt time.Time `db:"updated_at"`
}

// UpdateServiceStatus 更新服务状态
func (d *Database) UpdateServiceStatus(serviceName string, online bool, version string) error {
    // 获取服务ID
    serviceID, err := d.GetServiceID(serviceName)
    if err != nil {
        return err
    }

    now := time.Now()

    // 更新状态
    query := `
    INSERT INTO service_status (service_id, online, last_seen, version, updated_at)
    VALUES (?, ?, ?, ?, ?)
    ON CONFLICT(service_id) DO UPDATE SET
        online = excluded.online,
        last_seen = excluded.last_seen,
        version = excluded.version,
        updated_at = excluded.updated_at
    `
    _, err = d.db.Exec(query, serviceID, online, now, version, now)
    if err != nil {
        return err
    }

    // 记录历史
    _, err = d.db.Exec("INSERT INTO service_status_history (service_id, online, timestamp) VALUES (?, ?, ?)",
        serviceID, online, now)
    return err
}

// GetServiceStatus 获取服务状态
func (d *Database) GetServiceStatus(serviceName string) (*ServiceStatus, error) {
    query := `
    SELECT ss.id, ss.service_id, s.name as service_name, ss.online, ss.last_seen, ss.version, ss.updated_at
    FROM service_status ss
    INNER JOIN services s ON s.id = ss.service_id
    WHERE s.name = ?
    `

    row := d.db.QueryRow(query, serviceName)

    var s ServiceStatus
    err := row.Scan(&s.ID, &s.ServiceID, &s.ServiceName, &s.Online,
        &s.LastSeen, &s.Version, &s.UpdatedAt)
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("service status not found")
    }
    return &s, err
}

// ListServiceStatus 列出所有服务状态
func (d *Database) ListServiceStatus() ([]*ServiceStatus, error) {
    query := `
    SELECT ss.id, ss.service_id, s.name as service_name, ss.online, ss.last_seen, ss.version, ss.updated_at
    FROM service_status ss
    INNER JOIN services s ON s.id = ss.service_id
    ORDER BY s.name
    `

    rows, err := d.db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var statuses []*ServiceStatus
    for rows.Next() {
        var s ServiceStatus
        if err := rows.Scan(&s.ID, &s.ServiceID, &s.ServiceName, &s.Online,
            &s.LastSeen, &s.Version, &s.UpdatedAt); err != nil {
            return nil, err
        }
        statuses = append(statuses, &s)
    }

    return statuses, rows.Err()
}
```

**验收标准**:
- [ ] 状态可以更新
- [ ] 状态可以查询
- [ ] 历史记录正确保存

---

#### 2.5 事件存储操作
**文件**: `console/server/storage/event.go`

```go
package storage

type ServiceEvent struct {
    ID        int64     `db:"id"`
    Type      string    `db:"type"`
    Service   string    `db:"service"`
    Method    string    `db:"method"`
    Data      string    `db:"data"`
    Timestamp time.Time `db:"timestamp"`
}

// SaveEvent 保存事件
func (d *Database) SaveEvent(event *ServiceEvent) error {
    query := `
    INSERT INTO events (type, service, method, data, timestamp)
    VALUES (?, ?, ?, ?, ?)
    `
    _, err := d.db.Exec(query, event.Type, event.Service, event.Method,
        event.Data, event.Timestamp)
    return err
}

// ListEvents 列出事件
func (d *Database) ListEvents(limit, offset int) ([]*ServiceEvent, error) {
    query := `
    SELECT id, type, service, method, data, timestamp
    FROM events
    ORDER BY timestamp DESC
    LIMIT ? OFFSET ?
    `

    rows, err := d.db.Query(query, limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var events []*ServiceEvent
    for rows.Next() {
        var e ServiceEvent
        if err := rows.Scan(&e.ID, &e.Type, &e.Service, &e.Method,
            &e.Data, &e.Timestamp); err != nil {
            return nil, err
        }
        events = append(events, &e)
    }

    return events, rows.Err()
}
```

**验收标准**:
- [ ] 事件可以保存
- [ ] 事件可以分页查询

---

#### 2.6 用户存储操作
**文件**: `console/server/storage/user.go`

```go
package storage

import (
    "database/sql"
    "golang.org/x/crypto/bcrypt"
)

type User struct {
    ID           int64  `db:"id"`
    Username     string `db:"username"`
    PasswordHash string `db:"password_hash"`
    Role         string `db:"role"`
    CreatedAt    time.Time `db:"created_at"`
}

// CreateUser 创建用户
func (d *Database) CreateUser(username, password, role string) error {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }

    query := `INSERT INTO users (username, password_hash, role) VALUES (?, ?, ?)`
    _, err = d.db.Exec(query, username, string(hash), role)
    return err
}

// GetUser 获取用户
func (d *Database) GetUser(username string) (*User, error) {
    query := `SELECT id, username, password_hash, role, created_at FROM users WHERE username = ?`
    row := d.db.QueryRow(query, username)

    var u User
    err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt)
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("user not found")
    }
    return &u, err
}

// ValidateUser 验证用户
func (d *Database) ValidateUser(username, password string) (*User, error) {
    u, err := d.GetUser(username)
    if err != nil {
        return nil, err
    }

    err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
    if err != nil {
        return nil, fmt.Errorf("invalid password")
    }

    return u, nil
}

// InitAdminUser 初始化管理员用户
func (d *Database) InitAdminUser(username, password string) error {
    _, err := d.GetUser(username)
    if err == nil {
        return nil // 用户已存在
    }

    return d.CreateUser(username, password, "admin")
}
```

**验收标准**:
- [ ] 用户可以创建
- [ ] 密码正确加密
- [ ] 验证功能正常

---

#### 2.7 Service Manager - 注册消息处理器
**文件**: `console/server/manager/registry.go`

```go
package manager

import (
    "encoding/json"
    "fmt"
    "github.com/LiteHomeLab/light_link/console/server/storage"
    "github.com/LiteHomeLab/light_link/sdk/go/types"
    "github.com/nats-io/nats.go"
)

type Registry struct {
    db       *storage.Database
    nc       *nats.Conn
    eventCh  chan *types.ServiceEvent
}

func NewRegistry(db *storage.Database, nc *nats.Conn) *Registry {
    return &Registry{
        db:      db,
        nc:      nc,
        eventCh: make(chan *types.ServiceEvent, 100),
    }
}

// Subscribe 订阅注册消息
func (r *Registry) Subscribe() error {
    _, err := r.nc.Subscribe("$LL.register.>", r.handleRegister)
    return err
}

// handleRegister 处理注册消息
func (r *Registry) handleRegister(msg *nats.Msg) {
    var register types.RegisterMessage
    if err := json.Unmarshal(msg.Data, &register); err != nil {
        return
    }

    // 保存服务元数据
    meta := &storage.ServiceMetadata{
        Name:        register.Metadata.Name,
        Version:     register.Metadata.Version,
        Description: register.Metadata.Description,
        Author:      register.Metadata.Author,
        Tags:        register.Metadata.Tags,
    }

    if err := r.db.SaveService(meta); err != nil {
        return
    }

    // 获取服务ID
    serviceID, _ := r.db.GetServiceID(register.Metadata.Name)

    // 保存方法
    for _, m := range register.Metadata.Methods {
        methodMeta := &storage.MethodMetadata{
            ServiceID:   serviceID,
            Name:        m.Name,
            Description: m.Description,
            Params:      convertParams(m.Params),
            Returns:     convertReturns(m.Returns),
            Example:     convertExample(m.Example),
            Tags:        m.Tags,
            Deprecated:  m.Deprecated,
        }
        r.db.SaveMethod(serviceID, methodMeta)
    }

    // 更新状态为在线
    r.db.UpdateServiceStatus(register.Metadata.Name, true, register.Version)

    // 发送事件
    r.eventCh <- &types.ServiceEvent{
        Type:      "registered",
        Service:   register.Metadata.Name,
        Timestamp: register.Timestamp,
    }
}

// Events 返回事件通道
func (r *Registry) Events() <-chan *types.ServiceEvent {
    return r.eventCh
}
```

**验收标准**:
- [ ] 正确订阅 `$LL.register.>`
- [ ] 注册消息正确解析
- [ ] 服务和元数据正确保存
- [ ] 事件正确发送

---

#### 2.8 Service Manager - 心跳处理器
**文件**: `console/server/manager/heartbeat.go`

```go
package manager

import (
    "encoding/json"
    "github.com/LiteHomeLab/light_link/console/server/storage"
    "github.com/LiteHomeLab/light_link/sdk/go/types"
    "github.com/nats-io/nats.go"
    "time"
)

type HeartbeatMonitor struct {
    db          *storage.Database
    nc          *nats.Conn
    eventCh     chan *types.ServiceEvent
    timeout     time.Duration
    lastSeen    map[string]time.Time
}

func NewHeartbeatMonitor(db *storage.Database, nc *nats.Conn, timeout time.Duration) *HeartbeatMonitor {
    return &HeartbeatMonitor{
        db:       db,
        nc:       nc,
        eventCh:  make(chan *types.ServiceEvent, 100),
        timeout:  timeout,
        lastSeen: make(map[string]time.Time),
    }
}

// Subscribe 订阅心跳消息
func (h *HeartbeatMonitor) Subscribe() error {
    _, err := h.nc.Subscribe("$LL.heartbeat.>", h.handleHeartbeat)
    return err
}

// handleHeartbeat 处理心跳消息
func (h *HeartbeatMonitor) handleHeartbeat(msg *nats.Msg) {
    var heartbeat types.HeartbeatMessage
    if err := json.Unmarshal(msg.Data, &heartbeat); err != nil {
        return
    }

    // 更新最后活跃时间
    h.lastSeen[heartbeat.Service] = heartbeat.Timestamp

    // 更新数据库状态
    wasOnline := false
    if status, err := h.db.GetServiceStatus(heartbeat.Service); err == nil {
        wasOnline = status.Online
    }

    h.db.UpdateServiceStatus(heartbeat.Service, true, heartbeat.Version)

    // 如果之前离线，现在上线
    if !wasOnline {
        h.eventCh <- &types.ServiceEvent{
            Type:      "online",
            Service:   heartbeat.Service,
            Timestamp: heartbeat.Timestamp,
        }
    }
}

// StartChecker 启动超时检查器
func (h *HeartbeatMonitor) StartChecker() {
    ticker := time.NewTicker(10 * time.Second)
    go func() {
        for range ticker.C {
            h.checkTimeouts()
        }
    }()
}

// checkTimeouts 检查超时服务
func (h *HeartbeatMonitor) checkTimeouts() {
    now := time.Now()

    for service, lastSeen := range h.lastSeen {
        if now.Sub(lastSeen) > h.timeout {
            // 服务超时，标记离线
            h.db.UpdateServiceStatus(service, false, "")
            h.eventCh <- &types.ServiceEvent{
                Type:      "offline",
                Service:   service,
                Timestamp: now,
            }
            delete(h.lastSeen, service)
        }
    }
}

// Events 返回事件通道
func (h *HeartbeatMonitor) Events() <-chan *types.ServiceEvent {
    return h.eventCh
}
```

**验收标准**:
- [ ] 正确订阅心跳消息
- [ ] 心跳超时检测正常
- [ ] 状态变更事件正确发送

---

#### 2.9 Service Manager 主入口
**文件**: `console/server/manager/manager.go`

```go
package manager

import (
    "github.com/LiteHomeLab/light_link/console/server/storage"
    "github.com/LiteHomeLab/light_link/sdk/go/types"
    "github.com/nats-io/nats.go"
    "time"
)

type Manager struct {
    db       *storage.Database
    nc       *nats.Conn
    registry *Registry
    monitor  *HeartbeatMonitor
    eventCh  chan *types.ServiceEvent
}

func NewManager(db *storage.Database, nc *nats.Conn, heartbeatTimeout time.Duration) *Manager {
    m := &Manager{
        db:      db,
        nc:      nc,
        eventCh: make(chan *types.ServiceEvent, 100),
    }

    m.registry = NewRegistry(db, nc)
    m.monitor = NewHeartbeatMonitor(db, nc, heartbeatTimeout)

    return m
}

// Start 启动管理器
func (m *Manager) Start() error {
    // 订阅注册消息
    if err := m.registry.Subscribe(); err != nil {
        return err
    }

    // 订阅心跳消息
    if err := m.monitor.Subscribe(); err != nil {
        return err
    }

    // 启动超时检查
    m.monitor.StartChecker()

    // 启动事件转发
    go m.forwardEvents()

    return nil
}

// forwardEvents 转发事件
func (m *Manager) forwardEvents() {
    for {
        select {
        case e := <-m.registry.Events():
            m.db.SaveEvent(&storage.ServiceEvent{
                Type:      e.Type,
                Service:   e.Service,
                Timestamp: e.Timestamp,
            })
            m.eventCh <- e
        case e := <-m.monitor.Events():
            m.db.SaveEvent(&storage.ServiceEvent{
                Type:      e.Type,
                Service:   e.Service,
                Timestamp: e.Timestamp,
            })
            m.eventCh <- e
        }
    }
}

// Events 返回事件通道
func (m *Manager) Events() <-chan *types.ServiceEvent {
    return m.eventCh
}
```

**验收标准**:
- [ ] 管理器可以启动
- [ ] 所有订阅正常工作
- [ ] 事件正确转发

---

### Phase 2 验收标准
- [ ] SQLite 数据库正确初始化
- [ ] 所有 CRUD 操作正常
- [ ] 注册消息正确处理
- [ ] 心跳消息正确处理
- [ ] 超时检测正常工作
- [ ] 事件正确记录和发送

---

## Phase 3: Web 后端 REST API 和 WebSocket

### 目标
实现 Web 控制台的后端 API，包括 REST API、WebSocket、JWT 认证和 RPC 调用代理。

### 任务清单

#### 3.1 配置文件加载
**文件**: `console/server/config/config.go`

```go
package config

import (
    "fmt"
    "os"
    "time"

    "gopkg.in/yaml.v3"
)

type Config struct {
    Server    ServerConfig    `yaml:"server"`
    NATS      NATSConfig      `yaml:"nats"`
    Database  DatabaseConfig  `yaml:"database"`
    JWT       JWTConfig       `yaml:"jwt"`
    Heartbeat HeartbeatConfig `yaml:"heartbeat"`
    Admin     AdminConfig     `yaml:"admin"`
}

type ServerConfig struct {
    Host string `yaml:"host"`
    Port int    `yaml:"port"`
}

type NATSConfig struct {
    URL string      `yaml:"url"`
    TLS TLSConfig   `yaml:"tls"`
}

type TLSConfig struct {
    Enabled bool   `yaml:"enabled"`
    Cert    string `yaml:"cert"`
    Key     string `yaml:"key"`
    CA      string `yaml:"ca"`
}

type DatabaseConfig struct {
    Path string `yaml:"path"`
}

type JWTConfig struct {
    Secret string        `yaml:"secret"`
    Expiry time.Duration `yaml:"expiry"`
}

type HeartbeatConfig struct {
    Interval time.Duration `yaml:"interval"`
    Timeout  time.Duration `yaml:"timeout"`
}

type AdminConfig struct {
    Username string `yaml:"username"`
    Password string `yaml:"password"`
}

func Load(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}

func (c *Config) ServerAddr() string {
    return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
```

**验收标准**:
- [ ] YAML 配置正确加载
- [ ] 默认值正确设置

---

#### 3.2 JWT 认证中间件
**文件**: `console/server/auth/jwt.go`

```go
package auth

import (
    "fmt"
    "net/http"
    "strings"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}

type AuthMiddleware struct {
    secret  string
    expiry  time.Duration
}

func NewAuthMiddleware(secret string, expiry time.Duration) *AuthMiddleware {
    return &AuthMiddleware{
        secret: secret,
        expiry: expiry,
    }
}

// GenerateToken 生成 JWT Token
func (a *AuthMiddleware) GenerateToken(username, role string) (string, error) {
    claims := Claims{
        Username: username,
        Role:     role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.expiry)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(a.secret))
}

// ValidateToken 验证 JWT Token
func (a *AuthMiddleware) ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(a.secret), nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }

    return nil, fmt.Errorf("invalid token")
}

// Middleware 中间件
func (a *AuthMiddleware) Middleware() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 跳过登录接口
            if r.URL.Path == "/api/auth/login" {
                next.ServeHTTP(w, r)
                return
            }

            // 获取 token
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            tokenString := strings.TrimPrefix(authHeader, "Bearer ")
            claims, err := a.ValidateToken(tokenString)
            if err != nil {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            // 将用户信息存入 context
            ctx := context.WithValue(r.Context(), "username", claims.Username)
            ctx = context.WithValue(ctx, "role", claims.Role)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// RequireAdmin 要求管理员权限
func RequireAdmin(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        role := r.Context().Value("role")
        if role != "admin" {
            http.Error(w, "Forbidden", http.StatusForbidden)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

**验收标准**:
- [ ] Token 可以生成
- [ ] Token 可以验证
- [ ] 中间件正确拦截未授权请求

---

#### 3.3 REST API 路由和处理器
**文件**: `console/server/api/routes.go`

```go
package api

import (
    "github.com/LiteHomeLab/light_link/console/server/auth"
    "github.com/LiteHomeLab/light_link/console/server/manager"
    "github.com/LiteHomeLab/light_link/console/server/storage"
    "net/http"
)

type Handler struct {
    db      *storage.Database
    manager *manager.Manager
    auth    *auth.AuthMiddleware
}

func NewHandler(db *storage.Database, mgr *manager.Manager, auth *auth.AuthMiddleware) *Handler {
    return &Handler{
        db:      db,
        manager: mgr,
        auth:    auth,
    }
}

// Routes 注册路由
func (h *Handler) Routes() http.Handler {
    mux := http.NewServeMux()

    // 认证
    mux.HandleFunc("/api/auth/login", h.handleLogin)

    // 服务
    mux.HandleFunc("/api/services", h.withAuth(h.handleServices))
    mux.HandleFunc("/api/services/", h.withAuth(h.handleServiceDetail))

    // 状态
    mux.HandleFunc("/api/status", h.withAuth(h.handleStatus))
    mux.HandleFunc("/api/status/", h.withAuth(h.handleServiceStatus))

    // 方法
    mux.HandleFunc("/api/services/", h.withAuth(h.handleMethods))

    // 事件
    mux.HandleFunc("/api/events", h.withAuth(h.handleEvents))

    // 调用
    mux.HandleFunc("/api/call", h.withAuth(h.handleCall))

    // WebSocket
    mux.HandleFunc("/api/ws", h.withAuth(h.handleWebSocket))

    // 静态文件
    mux.Handle("/", http.FileServer(http.Dir("web/dist")))

    return h.auth.Middleware()(mux)
}

func (h *Handler) withAuth(fn http.HandlerFunc) http.HandlerFunc {
    return fn
}
```

**验收标准**:
- [ ] 所有路由正确注册
- [ ] 认证中间件正确应用

---

#### 3.4 服务 API
**文件**: `console/server/api/service.go`

```go
package api

import (
    "encoding/json"
    "net/http"
)

// handleServices 获取服务列表
func (h *Handler) handleServices(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    services, err := h.db.ListServices()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(services)
}

// handleServiceDetail 获取服务详情
func (h *Handler) handleServiceDetail(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // 从 URL 获取服务名
    // /api/services/demo-service -> demo-service
    serviceName := strings.TrimPrefix(r.URL.Path, "/api/services/")

    service, err := h.db.GetService(serviceName)
    if err != nil {
        http.Error(w, "Service not found", http.StatusNotFound)
        return
    }

    json.NewEncoder(w).Encode(service)
}
```

**验收标准**:
- [ ] 服务列表正确返回
- [ ] 服务详情正确返回

---

#### 3.5 WebSocket Hub
**文件**: `console/server/ws/hub.go`

```go
package ws

import (
    "encoding/json"
    "log"
    "sync"

    "github.com/gorilla/websocket"
    "github.com/LiteHomeLab/light_link/sdk/go/types"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // 生产环境需要验证
    },
}

type Client struct {
    conn     *websocket.Conn
    send     chan []byte
    channels map[string]bool
}

type Hub struct {
    clients    map[*Client]bool
    register   chan *Client
    unregister chan *Client
    broadcast  chan *Message
    eventCh    chan *types.ServiceEvent
    mu         sync.RWMutex
}

type Message struct {
    Channel string      `json:"channel"`
    Event   interface{} `json:"event"`
}

func NewHub() *Hub {
    return &Hub{
        clients:    make(map[*Client]bool),
        register:   make(chan *Client),
        unregister: make(chan *Client),
        broadcast:  make(chan *Message, 256),
        eventCh:    make(chan *types.ServiceEvent, 256),
    }
}

// Run 运行 Hub
func (h *Hub) Run() {
    go h.processEvents()
    go h.broadcastLoop()

    for {
        select {
        case client := <-h.register:
            h.clients[client] = true

        case client := <-h.unregister:
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                close(client.send)
            }

        case message := <-h.broadcast:
            h.broadcastMessage(message)
        }
    }
}

// processEvents 处理事件并广播
func (h *Hub) processEvents() {
    for event := range h.eventCh {
        msg := &Message{
            Channel: "events",
            Event:   event,
        }
        h.broadcast <- msg
    }
}

// broadcastLoop 广播消息
func (h *Hub) broadcastLoop() {
    for {
        select {
        case message := <-h.broadcast:
            h.broadcastMessage(message)
        }
    }
}

// broadcastMessage 广播消息到所有订阅的客户端
func (h *Hub) broadcastMessage(message *Message) {
    h.mu.RLock()
    defer h.mu.RUnlock()

    data, _ := json.Marshal(message)

    for client := range h.clients {
        if client.channels[message.Channel] {
            select {
            case client.send <- data:
            default:
                delete(h.clients, client)
                close(client.send)
            }
        }
    }
}

// Events 返回事件通道
func (h *Hub) Events() chan<- *types.ServiceEvent {
    return h.eventCh
}

// HandleWebSocket 处理 WebSocket 连接
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("WebSocket upgrade error:", err)
        return
    }

    client := &Client{
        conn:     conn,
        send:     make(chan []byte, 256),
        channels: make(map[string]bool),
    }

    h.register <- client

    go client.writePump()
    go client.readPump(h)
}

// readPump 读取客户端消息
func (c *Client) readPump(hub *Hub) {
    defer func() {
        hub.unregister <- c
        c.conn.Close()
    }()

    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            break
        }

        // 解析订阅消息
        var msg struct {
            Action   string   `json:"action"`
            Channels []string `json:"channels"`
        }

        if err := json.Unmarshal(message, &msg); err == nil {
            if msg.Action == "subscribe" {
                for _, ch := range msg.Channels {
                    c.channels[ch] = true
                }
            }
        }
    }
}

// writePump 写入消息到客户端
func (c *Client) writePump() {
    defer c.conn.Close()

    for {
        select {
        case message, ok := <-c.send:
            if !ok {
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            c.conn.WriteMessage(websocket.TextMessage, message)
        }
    }
}
```

**验收标准**:
- [ ] WebSocket 连接正常
- [ ] 订阅机制正常工作
- [ ] 事件正确广播

---

#### 3.6 RPC 调用代理
**文件**: `console/server/proxy/caller.go`

```go
package proxy

import (
    "encoding/json"
    "fmt"
    "time"

    "github.com/LiteHomeLab/light_link/sdk/go/client"
)

type Caller struct {
    clients map[string]*client.Client
}

func NewCaller() *Caller {
    return &Caller{
        clients: make(map[string]*client.Client),
    }
}

// Call 调用 RPC 方法
func (c *Caller) Call(service, method string, params map[string]interface{}) (interface{}, error) {
    cli, ok := c.clients[service]
    if !ok {
        return nil, fmt.Errorf("service client not found: %s", service)
    }

    return cli.Call(method, params)
}

// CallWithTimeout 带超时的调用
func (c *Caller) CallWithTimeout(service, method string, params map[string]interface{}, timeout time.Duration) (interface{}, error) {
    cli, ok := c.clients[service]
    if !ok {
        return nil, fmt.Errorf("service client not found: %s", service)
    }

    return cli.CallWithTimeout(method, params, timeout)
}

// RegisterClient 注册客户端
func (c *Caller) RegisterClient(service string, client *client.Client) {
    c.clients[service] = client
}
```

**验收标准**:
- [ ] RPC 调用正常工作
- [ ] 超时控制正常

---

#### 3.7 主服务器入口
**文件**: `console/server/main.go`

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"

    "github.com/LiteHomeLab/light_link/console/server/api"
    "github.com/LiteHomeLab/light_link/console/server/auth"
    "github.com/LiteHomeLab/light_link/console/server/config"
    "github.com/LiteHomeLab/light_link/console/server/manager"
    "github.com/LiteHomeLab/light_link/console/server/storage"
    "github.com/LiteHomeLab/light_link/console/server/ws"
    "github.com/nats-io/nats.go"
)

func main() {
    // 加载配置
    cfg, err := config.Load("console.yaml")
    if err != nil {
        log.Fatal("Load config:", err)
    }

    // 连接 NATS
    nc, err := connectNATS(cfg)
    if err != nil {
        log.Fatal("Connect NATS:", err)
    }
    defer nc.Close()

    // 初始化数据库
    db, err := storage.NewDatabase(cfg.Database.Path)
    if err != nil {
        log.Fatal("Init database:", err)
    }
    defer db.Close()

    // 初始化管理员
    if err := db.InitAdminUser(cfg.Admin.Username, cfg.Admin.Password); err != nil {
        log.Fatal("Init admin:", err)
    }

    // 启动服务管理器
    mgr := manager.NewManager(db, nc, cfg.Heartbeat.Timeout)
    if err := mgr.Start(); err != nil {
        log.Fatal("Start manager:", err)
    }

    // JWT 认证
    auth := auth.NewAuthMiddleware(cfg.JWT.Secret, cfg.JWT.Expiry)

    // WebSocket Hub
    hub := ws.NewHub()
    go hub.Run()

    // 连接管理器事件到 Hub
    go func() {
        for event := range mgr.Events() {
            hub.Events() <- event
        }
    }()

    // API Handler
    handler := api.NewHandler(db, mgr, auth)

    // 启动服务器
    server := &http.Server{
        Addr:    cfg.ServerAddr(),
        Handler: handler.Routes(),
    }

    go func() {
        log.Printf("Server started on %s", cfg.ServerAddr())
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal("Server error:", err)
        }
    }()

    // 等待退出信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    // 关闭服务器
    ctx := context.Background()
    server.Shutdown(ctx)
    log.Println("Server stopped")
}

func connectNATS(cfg *config.Config) (*nats.Conn, error) {
    opts := []nats.Option{}

    if cfg.NATS.TLS.Enabled {
        opts = append(opts, nats.Secure())
    }

    return nats.Connect(cfg.NATS.URL, opts...)
}
```

**验收标准**:
- [ ] 服务器正常启动
- [ ] 所有组件正确初始化
- [ ] 优雅关闭正常工作

---

### Phase 3 验收标准
- [ ] 所有 REST API 正常工作
- [ ] WebSocket 连接和消息推送正常
- [ ] JWT 认证正常
- [ ] RPC 调用代理正常
- [ ] 服务器可以正常启动和关闭

---

## Phase 4: Web 前端 (Vue 3 + Element Plus)

### 目标
实现 Web 控制台的前端界面。

### 任务清单

#### 4.1 项目初始化

```bash
cd console/web
npm create vite@latest . -- --template vue-ts
npm install
npm install element-plus @element-plus/icons-vue
npm install axios pinia vue-router
```

**验收标准**:
- [ ] 项目可以正常启动
- [ ] Element Plus 可以正常使用

---

#### 4.2 项目结构

```
console/web/src/
├── api/
│   ├── index.ts        # API 客户端
│   └── types.ts        # 类型定义
├── components/
│   ├── ServiceCard.vue
│   ├── MethodList.vue
│   └── ...
├── views/
│   ├── LoginView.vue
│   ├── ServicesView.vue
│   ├── ServiceDetailView.vue
│   ├── DebugView.vue
│   ├── DashboardView.vue
│   └── EventsView.vue
├── router/
│   └── index.ts
├── stores/
│   ├── user.ts
│   └── services.ts
├── utils/
│   └── websocket.ts
├── App.vue
└── main.ts
```

---

#### 4.3 API 客户端
**文件**: `console/web/src/api/index.ts`

```typescript
import axios from 'axios'
import type { AxiosInstance } from 'axios'

const api: AxiosInstance = axios.create({
  baseURL: '/api',
  timeout: 10000
})

// 请求拦截器 - 添加 token
api.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截器
api.interceptors.response.use(
  response => response.data,
  error => {
    if (error.response?.status === 401) {
      // 跳转登录
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export interface ServiceMetadata {
  name: string
  version: string
  description: string
  author: string
  tags: string[]
  registered_at: string
  updated_at: string
}

export interface MethodMetadata {
  name: string
  description: string
  params: ParameterMetadata[]
  returns: ReturnMetadata[]
  example?: ExampleMetadata
  tags: string[]
  deprecated: boolean
}

export interface ServiceStatus {
  service_name: string
  online: boolean
  last_seen: string
  version: string
}

export interface ServiceEvent {
  type: string
  service: string
  timestamp: string
}

export const servicesApi = {
  list: () => api.get<ServiceMetadata[]>('/services'),
  get: (name: string) => api.get<ServiceMetadata>(`/services/${name}`),
  getMethods: (name: string) => api.get<MethodMetadata[]>(`/services/${name}/methods`),
  getMethod: (service: string, method: string) =>
    api.get<MethodMetadata>(`/services/${service}/methods/${method}`)
}

export const statusApi = {
  list: () => api.get<ServiceStatus[]>('/status'),
  get: (name: string) => api.get<ServiceStatus>(`/status/${name}`)
}

export const eventsApi = {
  list: (limit = 100, offset = 0) => api.get<ServiceEvent[]>(`/events?limit=${limit}&offset=${offset}`)
}

export const callApi = {
  call: (service: string, method: string, params: Record<string, any>) =>
    api.post('/call', { service, method, params })
}

export const authApi = {
  login: (username: string, password: string) =>
    api.post<{ token: string }>('/auth/login', { username, password })
}
```

**验收标准**:
- [ ] API 客户端正常工作
- [ ] 拦截器正确处理 token

---

#### 4.4 WebSocket 工具
**文件**: `console/web/src/utils/websocket.ts`

```typescript
import { Event } from './types'

export class WebSocketClient {
  private ws: WebSocket | null = null
  private url: string
  private channels: Set<string> = new Set()
  private handlers: Map<string, ((event: any) => void)[]> = new Map()
  private reconnectTimer: number | null = null

  constructor(url: string) {
    this.url = url
  }

  connect() {
    const token = localStorage.getItem('token')
    const wsUrl = `${this.url}?token=${token}`

    this.ws = new WebSocket(wsUrl)

    this.ws.onopen = () => {
      console.log('WebSocket connected')
      // 订阅默认频道
      this.subscribe(['services', 'status', 'events'])
    }

    this.ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data)
        const handlers = this.handlers.get(message.channel) || []
        handlers.forEach(handler => handler(message.event))
      } catch (e) {
        console.error('Parse message error:', e)
      }
    }

    this.ws.onclose = () => {
      console.log('WebSocket disconnected')
      // 重连
      this.reconnect()
    }

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error)
    }
  }

  disconnect() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
    }
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }

  private reconnect() {
    this.reconnectTimer = window.setTimeout(() => {
      console.log('Reconnecting...')
      this.connect()
    }, 3000)
  }

  subscribe(channels: string[]) {
    channels.forEach(ch => this.channels.add(ch))
    this.sendSubscription()
  }

  unsubscribe(channels: string[]) {
    channels.forEach(ch => this.channels.delete(ch))
    this.sendSubscription()
  }

  private sendSubscription() {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({
        action: 'subscribe',
        channels: Array.from(this.channels)
      }))
    }
  }

  on(channel: string, handler: (event: any) => void) {
    if (!this.handlers.has(channel)) {
      this.handlers.set(channel, [])
    }
    this.handlers.get(channel)!.push(handler)
  }

  off(channel: string, handler: (event: any) => void) {
    const handlers = this.handlers.get(channel)
    if (handlers) {
      const index = handlers.indexOf(handler)
      if (index > -1) {
        handlers.splice(index, 1)
      }
    }
  }
}

export const ws = new WebSocketClient(`ws://${location.host}/api/ws`)
```

**验收标准**:
- [ ] WebSocket 连接正常
- [ ] 重连机制正常
- [ ] 订阅机制正常

---

#### 4.5 Pinia Store
**文件**: `console/web/src/stores/user.ts`

```typescript
import { defineStore } from 'pinia'
import { authApi } from '@/api'

export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    username: localStorage.getItem('username') || '',
    role: localStorage.getItem('role') || ''
  }),

  getters: {
    isLoggedIn: (state) => !!state.token,
    isAdmin: (state) => state.role === 'admin'
  },

  actions: {
    async login(username: string, password: string) {
      const data = await authApi.login(username, password)
      this.token = data.token
      this.username = username
      // 解析 token 获取 role (简化)
      this.role = 'admin'

      localStorage.setItem('token', data.token)
      localStorage.setItem('username', username)
      localStorage.setItem('role', this.role)
    },

    logout() {
      this.token = ''
      this.username = ''
      this.role = ''

      localStorage.removeItem('token')
      localStorage.removeItem('username')
      localStorage.removeItem('role')
    }
  }
})
```

**文件**: `console/web/src/stores/services.ts`

```typescript
import { defineStore } from 'pinia'
import { servicesApi, statusApi, type ServiceMetadata, type ServiceStatus } from '@/api'
import { ws } from '@/utils/websocket'

export const useServicesStore = defineStore('services', {
  state: () => ({
    services: [] as ServiceMetadata[],
    servicesStatus: new Map<string, ServiceStatus>(),
    loading: false
  }),

  actions: {
    async loadServices() {
      this.loading = true
      try {
        this.services = await servicesApi.list()
      } finally {
        this.loading = false
      }
    },

    async loadStatus() {
      const statuses = await statusApi.list()
      this.servicesStatus.clear()
      statuses.forEach(s => {
        this.servicesStatus.set(s.service_name, s)
      })
    },

    setupWebSocket() {
      ws.on('status', (event: any) => {
        if (event.service) {
          this.servicesStatus.set(event.service, event)
        }
      })
    }
  }
})
```

**验收标准**:
- [ ] 用户状态管理正常
- [ ] 服务状态管理正常
- [ ] WebSocket 集成正常

---

#### 4.6 登录页面
**文件**: `console/web/src/views/LoginView.vue`

```vue
<template>
  <div class="login-container">
    <el-card class="login-card">
      <h2>LightLink 服务管理平台</h2>
      <el-form :model="form" :rules="rules" ref="formRef">
        <el-form-item prop="username">
          <el-input v-model="form.username" placeholder="用户名" prefix-icon="User" />
        </el-form-item>
        <el-form-item prop="password">
          <el-input v-model="form.password" type="password" placeholder="密码" prefix-icon="Lock" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleLogin" :loading="loading" style="width: 100%">
            登录
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'

const router = useRouter()
const userStore = useUserStore()

const formRef = ref()
const loading = ref(false)

const form = reactive({
  username: '',
  password: ''
})

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

const handleLogin = async () => {
  await formRef.value.validate()
  loading.value = true
  try {
    await userStore.login(form.username, form.password)
    ElMessage.success('登录成功')
    router.push('/')
  } catch (e: any) {
    ElMessage.error(e.message || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-card {
  width: 400px;
  padding: 20px;
}

h2 {
  text-align: center;
  margin-bottom: 20px;
}
</style>
```

**验收标准**:
- [ ] 登录功能正常
- [ ] 表单验证正常

---

#### 4.7 服务列表页面
**文件**: `console/web/src/views/ServicesView.vue`

```vue
<template>
  <div class="services-view">
    <el-page-header title="首页" @back="$router.push('/')">
      <template #content>服务列表</template>
    </el-page-header>

    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="8">
        <el-input v-model="search" placeholder="搜索服务..." prefix-icon="Search" />
      </el-col>
      <el-col :span="8">
        <el-radio-group v-model="filter">
          <el-radio-button label="全部" />
          <el-radio-button label="在线" />
          <el-radio-button label="离线" />
        </el-radio-group>
      </el-col>
      <el-col :span="8" style="text-align: right">
        <el-button @click="loadData" :loading="loading">
          <el-icon><Refresh /></el-icon> 刷新
        </el-button>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="8" v-for="service in filteredServices" :key="service.name">
        <el-card class="service-card" @click="viewService(service.name)">
          <div class="service-header">
            <el-badge :value="status(service.name)?.online ? '在线' : '离线'"
                      :type="status(service.name)?.online ? 'success' : 'danger'">
              <h3>{{ service.name }}</h3>
            </el-badge>
          </div>
          <div class="service-info">
            <p>版本: {{ service.version }}</p>
            <p>描述: {{ service.description || '无' }}</p>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useServicesStore } from '@/stores'
import { storeToRefs } from 'pinia'

const router = useRouter()
const servicesStore = useServicesStore()
const { services, servicesStatus, loading } = storeToRefs(servicesStore)

const search = ref('')
const filter = ref('全部')

const filteredServices = computed(() => {
  let result = services.value

  if (search.value) {
    result = result.filter(s => s.name.includes(search.value))
  }

  if (filter.value === '在线') {
    result = result.filter(s => servicesStatus.value.get(s.name)?.online)
  } else if (filter.value === '离线') {
    result = result.filter(s => !servicesStatus.value.get(s.name)?.online)
  }

  return result
})

const status = (name: string) => servicesStatus.value.get(name)

const loadData = () => {
  servicesStore.loadServices()
  servicesStore.loadStatus()
}

const viewService = (name: string) => {
  router.push(`/services/${name}`)
}

onMounted(() => {
  loadData()
  servicesStore.setupWebSocket()
})
</script>

<style scoped>
.service-card {
  cursor: pointer;
  transition: all 0.3s;
}

.service-card:hover {
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  transform: translateY(-2px);
}

.service-header h3 {
  margin: 0;
}

.service-info p {
  margin: 5px 0;
  color: #666;
}
</style>
```

**验收标准**:
- [ ] 服务列表正确显示
- [ ] 搜索和过滤正常
- [ ] 实时状态更新

---

#### 4.8 路由配置
**文件**: `console/web/src/router/index.ts`

```typescript
import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores'

const routes = [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/LoginView.vue')
  },
  {
    path: '/',
    name: 'home',
    component: () => import('@/views/ServicesView.vue),
    meta: { requiresAuth: true }
  },
  {
    path: '/services/:name',
    name: 'service-detail',
    component: () => import('@/views/ServiceDetailView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/debug',
    name: 'debug',
    component: () => import('@/views/DebugView.vue'),
    meta: { requiresAuth: true, requiresAdmin: true }
  },
  {
    path: '/dashboard',
    name: 'dashboard',
    component: () => import('@/views/DashboardView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/events',
    name: 'events',
    component: () => import('@/views/EventsView.vue'),
    meta: { requiresAuth: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const userStore = useUserStore()

  if (to.meta.requiresAuth && !userStore.isLoggedIn) {
    next('/login')
  } else if (to.meta.requiresAdmin && !userStore.isAdmin) {
    next('/')
  } else {
    next()
  }
})

export default router
```

**验收标准**:
- [ ] 路由正常工作
- [ ] 认证守卫正常

---

### Phase 4 验收标准
- [ ] 所有页面可以正常访问
- [ ] 登录功能正常
- [ ] 服务列表可以正常查看
- [ ] 实时状态更新正常

---

## Phase 5: SDK 集成和示例

### 目标
更新示例代码展示元数据注册功能。

### 任务清单

#### 5.1 创建带元数据的示例
**文件**: `examples/with-metadata/main.go`

```go
package main

import (
    "fmt"
    "log"

    "github.com/LiteHomeLab/light_link/sdk/go/client"
    "github.com/LiteHomeLab/light_link/sdk/go/service"
    "github.com/LiteHomeLab/light_link/sdk/go/types"
    "github.com/nats-io/nats.go"
)

func main() {
    // 连接 NATS
    nc, err := nats.Connect("nats://localhost:4222")
    if err != nil {
        log.Fatal("Connect NATS:", err)
    }
    defer nc.Close()

    // 创建服务
    svc := service.New("math-service", nc)

    // 定义方法元数据
    addMeta := &types.MethodMetadata{
        Name:        "add",
        Description: "两数相加",
        Params: []types.ParameterMetadata{
            {Name: "a", Type: "number", Required: true, Description: "第一个数"},
            {Name: "b", Type: "number", Required: true, Description: "第二个数"},
        },
        Returns: []types.ReturnMetadata{
            {Name: "result", Type: "number", Description: "计算结果"},
        },
        Example: &types.ExampleMetadata{
            Input:       map[string]any{"a": 10, "b": 20},
            Output:      map[string]any{"result": 30},
            Description: "10 + 20 = 30",
        },
        Tags: []string{"math", "basic"},
    }

    // 注册方法
    if err := svc.RegisterMethod("add", add, addMeta); err != nil {
        log.Fatal("Register add:", err)
    }

    // 注册服务元数据
    if err := svc.RegisterMetadata(&types.ServiceMetadata{
        Name:        "math-service",
        Version:     "v1.0.0",
        Description: "数学运算服务",
        Author:      "LiteHomeLab",
        Tags:        []string{"demo", "math"},
    }); err != nil {
        log.Fatal("Register metadata:", err)
    }

    // 启动服务
    fmt.Println("Math service starting...")
    svc.Start()
    select {}
}

func add(params map[string]interface{}) (interface{}, error) {
    a := params["a"].(float64)
    b := params["b"].(float64)
    return map[string]interface{}{"result": a + b}, nil
}
```

**验收标准**:
- [ ] 示例可以正常运行
- [ ] 元数据正确发送到 NATS
- [ ] 控制台可以看到服务

---

### Phase 5 验收标准
- [ ] 示例代码正常工作
- [ ] SDK API 友好易用
- [ ] 文档完整

---

## Phase 6: 测试与优化

### 目标
完成端到端测试、性能优化和文档完善。

### 任务清单

#### 6.1 端到端测试
- 服务注册流程
- 心跳和状态检测
- RPC 调用
- Web 界面操作

#### 6.2 性能优化
- 数据库查询优化
- WebSocket 消息批量处理
- 前端渲染优化

#### 6.3 文档完善
- API 文档
- 部署文档
- 使用指南

---

## 总结

本实施计划分为 6 个阶段，每个阶段都有明确的任务列表和验收标准。建议按照顺序逐个完成，每个阶段完成后进行测试和代码提交。

| 阶段 | 主要内容 | 预计工作量 |
|------|----------|-----------|
| Phase 1 | SDK 元数据系统和 NATS 消息 | 3-4 天 |
| Phase 2 | Service Manager 和 SQLite 存储 | 3-4 天 |
| Phase 3 | Web 后端 API 和 WebSocket | 3-4 天 |
| Phase 4 | Web 前端界面 | 5-7 天 |
| Phase 5 | SDK 集成和示例 | 1-2 天 |
| Phase 6 | 测试与优化 | 2-3 天 |

**总计**: 约 17-24 天
