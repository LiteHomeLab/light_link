# LightLink 框架实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**目标:** 基于 NATS 的多语言后端服务通信框架，支持 RPC 调用、发布订阅、状态保留、大文件传输和 TLS 证书认证。

**架构:** Go SDK 作为参考实现，提供客户端和服务端抽象。底层使用 NATS Core + JetStream (KV + Object Store)。TLS 双向认证保障安全。

**技术栈:**
- NATS Server 2.10+
- Go 1.21+
- github.com/nats-io/nats.go
- github.com/nats-io/nats.go/jetstream

---

## Phase 1: 项目初始化和基础结构

### Task 1.1: 创建 Go Module 和目录结构

**文件:**
- Create: `go.mod`
- Create: `sdk/go/client/client.go`
- Create: `sdk/go/service/service.go`
- Create: `sdk/go/types/types.go`
- Create: `deploy/nats/nats-server.conf`

**Step 1: 创建 go.mod**

```bash
cd C:\WorkSpace\Go2Hell\src\github.com\LiteHomeLab\light_link
go mod init github.com/LiteHomeLab/light_link
```

**Step 2: 添加 NATS 依赖**

```bash
go get github.com/nats-io/nats.go@latest
```

**Step 3: 创建 types.go - 公共类型定义**

文件: `sdk/go/types/types.go`

```go
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
    NATSURL     string   `json:"nats_url"`
    ServiceName string   `json:"service_name"`
    TLS         *TLSConfig `json:"tls,omitempty"`
}

type TLSConfig struct {
    CaFile      string `json:"ca_file"`
    CertFile    string `json:"cert_file"`
    KeyFile     string `json:"key_file"`
}
```

**Step 4: 运行编译检查**

```bash
go build ./sdk/go/types
```

**Step 5: 提交**

```bash
git add go.mod sdk/go/types/types.go
git commit -m "feat: initialize Go module and define common types"
```

---

### Task 1.2: 创建 NATS 服务器配置文件

**文件:**
- Create: `deploy/nats/nats-server.conf`

**Step 1: 写入 NATS 配置**

文件: `deploy/nats/nats-server.conf`

```nginx
# NATS Server Configuration
# 监听地址
host: "0.0.0.0"
port: 4222

# JetStream 配置
jetstream {
    # 存储目录
    store_dir: "./data"

    # 内存存储
    max_memory: 1GB

    # 文件存储
    max_file: 10GB
}

# TLS 配置
tls {
    # CA 证书
    ca_file: "./deploy/nats/tls/ca.crt"

    # 服务器证书
    cert_file: "./deploy/nats/tls/server.crt"

    # 服务器私钥
    key_file: "./deploy/nats/tls/server.key"

    # 验证客户端证书
    verify: true

    # 从证书映射用户
    verify_and_map: true

    # 最小 TLS 版本
    min_version: 1.2
}

# 日志配置
log_file: "./logs/nats-server.log"
logtime: true
debug: false
trace: false

# 最大连接数
max_connections: 1000

# 最大有效订阅数
max_subs_per_client: 1000
```

**Step 2: 创建 TLS 证书占位目录**

```bash
mkdir -p deploy/nats/tls
mkdir -p deploy/nats/users
```

**Step 3: 创建证书生成说明**

文件: `deploy/nats/tls/README.md`

```markdown
# TLS 证书生成

## 生成 CA 证书
```bash
openssl genrsa -out ca.key 2048
openssl req -new -x509 -days 365 -key ca.key -out ca.crt \
    -subj "/CN=LightLink CA"
```

## 生成服务器证书
```bash
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr \
    -subj "/CN=nats-server"
openssl x509 -req -days 365 -in server.csr -CA ca.crt -CAkey ca.key \
    -CAcreateserial -out server.crt
```

## 生成客户端证书 (示例: user-service)
```bash
openssl genrsa -out user-service.key 2048
openssl req -new -key user-service.key -out user-service.csr \
    -subj "/CN=user-service"
openssl x509 -req -days 365 -in user-service.csr -CA ca.crt -CAkey ca.key \
    -CAcreateserial -out user-service.crt
```

注意: 客户端证书的 CN (Common Name) 将作为 NATS 用户名
```

**Step 4: 提交**

```bash
git add deploy/nats/
git commit -m "feat: add NATS server configuration and TLS certificate guide"
```

---

## Phase 2: 客户端实现

### Task 2.1: 实现客户端连接管理

**文件:**
- Create: `sdk/go/client/connection.go`
- Create: `sdk/go/client/connection_test.go`

**Step 1: 写连接管理的测试**

文件: `sdk/go/client/connection_test.go`

```go
package client

import (
    "testing"
    "time"
)

func TestNewClient(t *testing.T) {
    // 测试创建客户端（不需要实际连接 NATS）
    client, err := NewClient("nats://localhost:4222", nil)
    if err != nil {
        t.Fatalf("NewClient failed: %v", err)
    }
    if client == nil {
        t.Fatal("client is nil")
    }
    client.Close()
}

func TestNewClientWithTLS(t *testing.T) {
    config := &TLSConfig{
        CaFile:   "../../deploy/nats/tls/ca.crt",
        CertFile: "../../deploy/nats/tls/user-service.crt",
        KeyFile:  "../../deploy/nats/tls/user-service.key",
    }

    // 注意: 这个测试需要证书存在，实际连接时才会验证
    client, err := NewClient("tls://localhost:4222", config)
    if err != nil {
        // 证书不存在时预期失败，这是 OK 的
        t.Logf("Expected failure without certs: %v", err)
        return
    }
    defer client.Close()
    t.Log("Client created with TLS config")
}

func TestClientClose(t *testing.T) {
    client, _ := NewClient("nats://localhost:4222", nil)
    err := client.Close()
    if err != nil {
        t.Fatalf("Close failed: %v", err)
    }
}

func TestClientGetNATSConn(t *testing.T) {
    client, _ := NewClient("nats://localhost:4222", nil)
    defer client.Close()

    conn := client.GetNATSConn()
    if conn == nil {
        t.Fatal("NATS connection is nil")
    }
}
```

**Step 2: 运行测试确认失败**

```bash
cd sdk/go/client
go test -v -run TestNewClient
```

预期: `undefined: NewClient`

**Step 3: 实现连接管理**

文件: `sdk/go/client/connection.go`

```go
package client

import (
    "crypto/tls"
    "crypto/x509"
    "io/ioutil"
    "github.com/nats-io/nats.go"
)

// TLS 配置
type TLSConfig struct {
    CaFile   string
    CertFile string
    KeyFile  string
}

// Client 客户端
type Client struct {
    nc *nats.Conn
}

// NewClient 创建新客户端
func NewClient(url string, tlsConfig *TLSConfig) (*Client, error) {
    opts := []nats.Option{
        nats.Name("LightLink Client"),
        nats.ReconnectWait(2 * time.Second),
        nats.MaxReconnects(10),
        nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
            if err != nil {
                println("Disconnected:", err.Error())
            }
        }),
        nats.ReconnectHandler(func(nc *nats.Conn) {
            println("Reconnected to", nc.ConnectedUrl())
        }),
    }

    // 配置 TLS
    if tlsConfig != nil {
        tlsOpt, err := createTLSOption(tlsConfig)
        if err != nil {
            return nil, err
        }
        opts = append(opts, tlsOpt)
    }

    nc, err := nats.Connect(url, opts...)
    if err != nil {
        return nil, err
    }

    return &Client{nc: nc}, nil
}

// createTLSOption 创建 TLS 选项
func createTLSOption(config *TLSConfig) (nats.Option, error) {
    // 加载客户端证书
    cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
    if err != nil {
        return nil, err
    }

    // 创建 CA 池
    pool := x509.NewCertPool()
    caCert, err := ioutil.ReadFile(config.CaFile)
    if err != nil {
        return nil, err
    }
    pool.AppendCertsFromPEM(caCert)

    // 创建 TLS 配置
    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
        RootCAs:      pool,
        MinVersion:   tls.VersionTLS12,
    }

    return nats.Secure(tlsConfig), nil
}

// GetNATSConn 获取 NATS 连接
func (c *Client) GetNATSConn() *nats.Conn {
    return c.nc
}

// Close 关闭客户端
func (c *Client) Close() error {
    if c.nc != nil {
        c.nc.Close()
    }
    return nil
}
```

**Step 4: 运行测试确认通过**

```bash
cd sdk/go/client
go test -v
```

**Step 5: 提交**

```bash
git add sdk/go/client/
git commit -m "feat: implement client connection management with TLS support"
```

---

### Task 2.2: 实现 RPC 调用

**文件:**
- Create: `sdk/go/client/rpc.go`
- Create: `sdk/go/client/rpc_test.go`
- Create: `sdk/go/service/service.go` (用于测试的服务端)

**Step 1: 写 RPC 测试**

文件: `sdk/go/client/rpc_test.go`

```go
package client

import (
    "testing"
    "time"
)

func TestCall(t *testing.T) {
    // 这个测试需要运行中的 NATS 服务器和服务
    client, err := NewClient("nats://localhost:4222", nil)
    if err != nil {
        t.Skip("Need running NATS server")
    }
    defer client.Close()

    // 测试调用不存在的服务
    result, err := client.Call("test-service", "testMethod", map[string]interface{}{"key": "value"})
    if err == nil {
        t.Error("Expected error for non-existent service")
    }
    t.Logf("Expected error: %v", err)
    t.Logf("Result: %v", result)
}

func TestCallWithTimeout(t *testing.T) {
    client, err := NewClient("nats://localhost:4222", nil)
    if err != nil {
        t.Skip("Need running NATS server")
    }
    defer client.Close()

    // 测试超时
    done := make(chan bool)
    go func() {
        _, _ = client.Call("timeout-service", "slowMethod", nil)
        done <- true
    }()

    select {
    case <-done:
        t.Log("Call completed")
    case <-time.After(2 * time.Second):
        t.Error("Call should timeout or fail")
    }
}
```

**Step 2: 运行测试确认失败**

```bash
cd sdk/go/client
go test -v -run TestCall
```

**Step 3: 实现 RPC 客户端**

文件: `sdk/go/client/rpc.go`

```go
package client

import (
    "encoding/json"
    "fmt"
    "time"
    "github.com/google/uuid"
    "github.com/LiteHomeLab/light_link/sdk/go/types"
)

// Call 同步 RPC 调用
func (c *Client) Call(service, method string, args map[string]interface{}) (map[string]interface{}, error) {
    return c.CallWithTimeout(service, method, args, 5*time.Second)
}

// CallWithTimeout 带超时的 RPC 调用
func (c *Client) CallWithTimeout(service, method string, args map[string]interface{}, timeout time.Duration) (map[string]interface{}, error) {
    // 生成请求 ID
    requestID := uuid.New().String()

    // 构建请求
    request := types.RPCRequest{
        ID:     requestID,
        Method: method,
        Args:   args,
    }

    reqData, err := json.Marshal(request)
    if err != nil {
        return nil, fmt.Errorf("marshal request: %w", err)
    }

    // 构建主题
    subject := fmt.Sprintf("$SRV.%s.%s", service, method)

    // 发送请求并等待响应
    msg, err := c.nc.Request(subject, reqData, timeout)
    if err != nil {
        return nil, fmt.Errorf("RPC request failed: %w", err)
    }

    // 解析响应
    var response types.RPCResponse
    if err := json.Unmarshal(msg.Data, &response); err != nil {
        return nil, fmt.Errorf("unmarshal response: %w", err)
    }

    if !response.Success {
        return nil, fmt.Errorf("RPC error: %s", response.Error)
    }

    return response.Result, nil
}
```

**Step 4: 添加依赖**

```bash
go get github.com/google/uuid
```

**Step 5: 运行测试**

```bash
cd sdk/go/client
go test -v -run TestCall
```

**Step 6: 提交**

```bash
git add sdk/go/client/rpc.go sdk/go/client/rpc_test.go go.mod go.sum
git commit -m "feat: implement RPC client call method"
```

---

### Task 2.3: 实现服务端

**文件:**
- Create: `sdk/go/service/service.go`
- Create: `sdk/go/service/service_test.go`

**Step 1: 写服务端测试**

文件: `sdk/go/service/service_test.go`

```go
package service

import (
    "testing"
    "time"
)

func TestNewService(t *testing.T) {
    svc, err := NewService("test-service", "nats://localhost:4222", nil)
    if err != nil {
        t.Skip("Need running NATS server")
    }
    defer svc.Stop()

    if svc.Name() != "test-service" {
        t.Errorf("Expected name 'test-service', got '%s'", svc.Name())
    }
}

func TestRegisterRPC(t *testing.T) {
    svc, err := NewService("test-service", "nats://localhost:4222", nil)
    if err != nil {
        t.Skip("Need running NATS server")
    }
    defer svc.Stop()

    handler := func(args map[string]interface{}) (map[string]interface{}, error) {
        return map[string]interface{}{"result": "ok"}, nil
    }

    err = svc.RegisterRPC("testMethod", handler)
    if err != nil {
        t.Fatalf("RegisterRPC failed: %v", err)
    }

    // 验证处理器已注册
    if !svc.HasRPC("testMethod") {
        t.Error("RPC method not registered")
    }
}

func TestStartStop(t *testing.T) {
    svc, err := NewService("test-service", "nats://localhost:4222", nil)
    if err != nil {
        t.Skip("Need running NATS server")
    }

    err = svc.Start()
    if err != nil {
        t.Fatalf("Start failed: %v", err)
    }

    // 给服务一点时间启动
    time.Sleep(100 * time.Millisecond)

    err = svc.Stop()
    if err != nil {
        t.Fatalf("Stop failed: %v", err)
    }
}
```

**Step 2: 运行测试确认失败**

```bash
cd sdk/go/service
go test -v -run TestNewService
```

**Step 3: 实现服务端**

文件: `sdk/go/service/service.go`

```go
package service

import (
    "encoding/json"
    "fmt"
    "sync"
    "time"
    "github.com/nats-io/nats.go"
    "github.com/LiteHomeLab/light_link/sdk/go/types"
)

// RPCHandler RPC 处理器函数类型
type RPCHandler func(args map[string]interface{}) (map[string]interface{}, error)

// Service 服务端
type Service struct {
    name     string
    nc       *nats.Conn
    rpcMap   map[string]RPCHandler
    rpcMutex sync.RWMutex
    running  bool
}

// NewService 创建新服务
func NewService(name, natsURL string, tlsConfig interface{}) (*Service, error) {
    nc, err := nats.Connect(natsURL,
        nats.Name("LightLink Service: "+name),
        nats.ReconnectWait(2*time.Second),
        nats.MaxReconnects(10),
    )
    if err != nil {
        return nil, err
    }

    return &Service{
        name:   name,
        nc:     nc,
        rpcMap: make(map[string]RPCHandler),
    }, nil
}

// Name 返回服务名
func (s *Service) Name() string {
    return s.name
}

// RegisterRPC 注册 RPC 方法
func (s *Service) RegisterRPC(method string, handler RPCHandler) error {
    s.rpcMutex.Lock()
    defer s.rpcMutex.Unlock()

    s.rpcMap[method] = handler
    return nil
}

// HasRPC 检查 RPC 方法是否已注册
func (s *Service) HasRPC(method string) bool {
    s.rpcMutex.RLock()
    defer s.rpcMutex.RUnlock()

    _, exists := s.rpcMap[method]
    return exists
}

// Start 启动服务
func (s *Service) Start() error {
    if s.running {
        return fmt.Errorf("service already running")
    }

    // 订阅所有 RPC 方法
    subject := fmt.Sprintf("$SRV.%s.>", s.name)
    _, err := s.nc.Subscribe(subject, s.handleRPC)
    if err != nil {
        return fmt.Errorf("subscribe failed: %w", err)
    }

    s.running = true
    return nil
}

// handleRPC 处理 RPC 请求
func (s *Service) handleRPC(msg *nats.Msg) {
    // 解析请求
    var request types.RPCRequest
    if err := json.Unmarshal(msg.Data, &request); err != nil {
        s.sendError(msg, "invalid request: "+err.Error())
        return
    }

    // 查找处理器
    s.rpcMutex.RLock()
    handler, exists := s.rpcMap[request.Method]
    s.rpcMutex.RUnlock()

    if !exists {
        s.sendError(msg, "method not found: "+request.Method)
        return
    }

    // 调用处理器
    result, err := handler(request.Args)
    if err != nil {
        s.sendError(msg, err.Error())
        return
    }

    // 发送响应
    response := types.RPCResponse{
        ID:      request.ID,
        Success: true,
        Result:  result,
    }

    respData, _ := json.Marshal(response)
    msg.Respond(respData)
}

// sendError 发送错误响应
func (s *Service) sendError(msg *nats.Msg, errMsg string) {
    response := types.RPCResponse{
        ID:      "",
        Success: false,
        Error:   errMsg,
    }
    respData, _ := json.Marshal(response)
    msg.Respond(respData)
}

// Stop 停止服务
func (s *Service) Stop() error {
    if !s.running {
        return nil
    }

    s.nc.Close()
    s.running = false
    return nil
}
```

**Step 4: 运行测试**

```bash
cd sdk/go/service
go test -v
```

**Step 5: 创建 RPC 集成测试**

文件: `examples/rpc-demo/main.go`

```go
package main

import (
    "fmt"
    "time"
    "github.com/LiteHomeLab/light_link/sdk/go/client"
    "github.com/LiteHomeLab/light_link/sdk/go/service"
)

func main() {
    // 启动服务
    svc, _ := service.NewService("demo-service", "nats://localhost:4222", nil)

    svc.RegisterRPC("add", func(args map[string]interface{}) (map[string]interface{}, error) {
        a := int(args["a"].(float64))
        b := int(args["b"].(float64))
        return map[string]interface{}{"sum": a + b}, nil
    })

    svc.Start()
    defer svc.Stop()

    // 等待服务就绪
    time.Sleep(100 * time.Millisecond)

    // 客户端调用
    cli, _ := client.NewClient("nats://localhost:4222", nil)
    defer cli.Close()

    result, err := cli.Call("demo-service", "add", map[string]interface{}{
        "a": 10,
        "b": 20,
    })

    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    fmt.Println("Result:", result)
}
```

**Step 6: 提交**

```bash
git add sdk/go/service/ examples/rpc-demo/
git commit -m "feat: implement service with RPC registration"
```

---

## Phase 3: 发布订阅

### Task 3.1: 实现发布订阅

**文件:**
- Create: `sdk/go/client/pubsub.go`
- Create: `sdk/go/client/pubsub_test.go`

**Step 1: 写测试**

文件: `sdk/go/client/pubsub_test.go`

```go
package client

import (
    "testing"
    "time"
)

func TestPublishSubscribe(t *testing.T) {
    subClient, err := NewClient("nats://localhost:4222", nil)
    if err != nil {
        t.Skip("Need running NATS server")
    }
    defer subClient.Close()

    pubClient, _ := NewClient("nats://localhost:4222", nil)
    defer pubClient.Close()

    received := make(chan map[string]interface{}, 1)

    // 订阅
    sub, err := subClient.Subscribe("test.subject", func(data map[string]interface{}) {
        received <- data
    })
    if err != nil {
        t.Fatalf("Subscribe failed: %v", err)
    }
    defer sub.Unsubscribe()

    // 发布
    err = pubClient.Publish("test.subject", map[string]interface{}{"msg": "hello"})
    if err != nil {
        t.Fatalf("Publish failed: %v", err)
    }

    // 等待接收
    select {
    case data := <-received:
        if data["msg"] != "hello" {
            t.Errorf("Expected 'hello', got '%v'", data["msg"])
        }
    case <-time.After(2 * time.Second):
        t.Error("Timeout waiting for message")
    }
}
```

**Step 2: 实现发布订阅**

文件: `sdk/go/client/pubsub.go`

```go
package client

import (
    "encoding/json"
    "sync"
    "github.com/nats-io/nats.go"
)

// MessageHandler 消息处理器
type MessageHandler func(data map[string]interface{})

// Subscription 订阅
type Subscription struct {
    sub *nats.Subscription
}

// Unsubscribe 取消订阅
func (s *Subscription) Unsubscribe() error {
    if s.sub != nil {
        return s.sub.Unsubscribe()
    }
    return nil
}

// Publish 发布消息
func (c *Client) Publish(subject string, data map[string]interface{}) error {
    msgData, err := json.Marshal(data)
    if err != nil {
        return err
    }

    return c.nc.Publish(subject, msgData)
}

// Subscribe 订阅消息
func (c *Client) Subscribe(subject string, handler MessageHandler) (*Subscription, error) {
    sub, err := c.nc.Subscribe(subject, func(msg *nats.Msg) {
        var data map[string]interface{}
        if err := json.Unmarshal(msg.Data, &data); err != nil {
            return
        }
        handler(data)
    })

    return &Subscription{sub: sub}, err
}
```

**Step 3: 运行测试**

```bash
cd sdk/go/client
go test -v -run TestPublishSubscribe
```

**Step 4: 提交**

```bash
git add sdk/go/client/pubsub.go sdk/go/client/pubsub_test.go
git commit -m "feat: implement publish/subscribe"
```

---

## Phase 4: 状态管理 (NATS KV)

### Task 4.1: 实现 KV 状态管理

**文件:**
- Create: `sdk/go/client/state.go`
- Create: `sdk/go/client/state_test.go`

**Step 1: 写测试**

文件: `sdk/go/client/state_test.go`

```go
package client

import (
    "testing"
)

func TestSetGetState(t *testing.T) {
    client, err := NewClient("nats://localhost:4222", nil)
    if err != nil {
        t.Skip("Need running NATS server with JetStream")
    }
    defer client.Close()

    // 设置状态
    err = client.SetState("test.key", map[string]interface{}{"value": 123})
    if err != nil {
        t.Fatalf("SetState failed: %v", err)
    }

    // 获取状态
    state, err := client.GetState("test.key")
    if err != nil {
        t.Fatalf("GetState failed: %v", err)
    }

    if state["value"].(float64) != 123 {
        t.Errorf("Expected 123, got %v", state["value"])
    }
}

func TestWatchState(t *testing.T) {
    client, err := NewClient("nats://localhost:4222", nil)
    if err != nil {
        t.Skip("Need running NATS server with JetStream")
    }
    defer client.Close()

    changes := make(chan map[string]interface{}, 1)

    // 监听状态变化
    stop, err := client.WatchState("test.watch", func(state map[string]interface{}) {
        changes <- state
    })
    if err != nil {
        t.Fatalf("WatchState failed: %v", err)
    }
    defer stop()

    // 修改状态
    client.SetState("test.watch", map[string]interface{}{"status": "updated"})

    // 等待通知
    select {
    case state := <-changes:
        if state["status"] != "updated" {
            t.Errorf("Expected 'updated', got '%v'", state["status"])
        }
    case <-time.After(2 * time.Second):
        t.Error("Timeout waiting for state change")
    }
}
```

**Step 2: 实现状态管理**

文件: `sdk/go/client/state.go`

```go
package client

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/nats-io/nats.go/jetstream"
)

// SetState 设置状态
func (c *Client) SetState(key string, value map[string]interface{}) error {
    js, err := jetstream.New(c.nc)
    if err != nil {
        return err
    }

    // 获取或创建 KV bucket
    kv, err := js.KeyValue(context.Background(), "light_link_state")
    if err != nil {
        // 创建 bucket
        kv, err = js.CreateKeyValue(context.Background(), jetstream.KeyValueConfig{
            Bucket: "light_link_state",
        })
        if err != nil {
            return err
        }
    }

    // 序列化值
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }

    // 设置 KV
    _, err = kv.Put(context.Background(), key, data)
    return err
}

// GetState 获取状态
func (c *Client) GetState(key string) (map[string]interface{}, error) {
    js, err := jetstream.New(c.nc)
    if err != nil {
        return nil, err
    }

    kv, err := js.KeyValue(context.Background(), "light_link_state")
    if err != nil {
        return nil, err
    }

    entry, err := kv.Get(context.Background(), key)
    if err != nil {
        return nil, err
    }

    var value map[string]interface{}
    if err := json.Unmarshal(entry.Value(), &value); err != nil {
        return nil, err
    }

    return value, nil
}

// WatchState 监听状态变化
func (c *Client) WatchState(key string, handler func(map[string]interface{})) (func(), error) {
    js, err := jetstream.New(c.nc)
    if err != nil {
        return nil, err
    }

    kv, err := js.KeyValue(context.Background(), "light_link_state")
    if err != nil {
        return nil, err
    }

    watcher, err := kv.Watch(context.Background(), key, jetstream.IgnoreDeletes())
    if err != nil {
        return nil, err
    }

    stop := make(chan struct{})

    go func() {
        for {
            select {
            case <-stop:
                watcher.Stop()
                return
            default:
                select {
                case entry := <-watcher.Updates():
                    if entry != nil {
                        var value map[string]interface{}
                        json.Unmarshal(entry.Value(), &value)
                        handler(value)
                    }
                case <-stop:
                    watcher.Stop()
                    return
                }
            }
        }
   }()

    return func() { close(stop) }, nil
}
```

**Step 3: 运行测试**

```bash
cd sdk/go/client
go test -v -run TestSetGetState
```

**Step 4: 提交**

```bash
git add sdk/go/client/state.go sdk/go/client/state_test.go
git commit -m "feat: implement state management with NATS KV"
```

---

## Phase 5: 文件传输

### Task 5.1: 实现文件传输

**文件:**
- Create: `sdk/go/client/file.go`
- Create: `sdk/go/client/file_test.go`

**Step 1: 写测试**

文件: `sdk/go/client/file_test.go`

```go
package client

import (
    "os"
    "testing"
    "io/ioutil"
)

func TestFileTransfer(t *testing.T) {
    client, err := NewClient("nats://localhost:4222", nil)
    if err != nil {
        t.Skip("Need running NATS server with JetStream")
    }
    defer client.Close()

    // 创建测试文件
    tmpFile, err := ioutil.TempFile("", "test-*.txt")
    if err != nil {
        t.Fatal(err)
    }
    defer os.Remove(tmpFile.Name())

    testData := []byte("Hello, LightLink File Transfer!")
    tmpFile.Write(testData)
    tmpFile.Close()

    // 上传文件
    fileID, err := client.UploadFile(tmpFile.Name(), "test.txt")
    if err != nil {
        t.Fatalf("UploadFile failed: %v", err)
    }

    // 下载文件
    outFile, err := ioutil.TempFile("", "download-*.txt")
    if err != nil {
        t.Fatal(err)
    }
    defer os.Remove(outFile.Name())
    outFile.Close()

    err = client.DownloadFile(fileID, outFile.Name())
    if err != nil {
        t.Fatalf("DownloadFile failed: %v", err)
    }

    // 验证内容
    downloaded, _ := ioutil.ReadFile(outFile.Name())
    if string(downloaded) != string(testData) {
        t.Error("File content mismatch")
    }
}
```

**Step 2: 实现文件传输**

文件: `sdk/go/client/file.go`

```go
package client

import (
    "context"
    "fmt"
    "os"
    "github.com/nats-io/nats.go/jetstream"
    "github.com/google/uuid"
)

// UploadFile 上传文件到 Object Store
func (c *Client) UploadFile(filePath, fileName string) (string, error) {
    js, err := jetstream.New(c.nc)
    if err != nil {
        return "", err
    }

    // 获取或创建 Object Store
    store, err := js.ObjectStore(context.Background(), "light_link_files")
    if err != nil {
        store, err = js.CreateObjectStore(context.Background(), jetstream.ObjectStoreConfig{
            Bucket: "light_link_files",
        })
        if err != nil {
            return "", err
        }
    }

    // 读取文件
    data, err := os.ReadFile(filePath)
    if err != nil {
        return "", err
    }

    // 生成文件 ID
    fileID := uuid.New().String()

    // 上传到 Object Store
    _, err = store.Put(context.Background(), fileID, data)
    if err != nil {
        return "", err
    }

    return fileID, nil
}

// DownloadFile 从 Object Store 下载文件
func (c *Client) DownloadFile(fileID, destPath string) error {
    js, err := jetstream.New(c.nc)
    if err != nil {
        return err
    }

    store, err := js.ObjectStore(context.Background(), "light_link_files")
    if err != nil {
        return err
    }

    // 获取文件
    result, err := store.Get(fileID)
    if err != nil {
        return err
    }

    // 写入文件
    data, err := result.Bytes()
    if err != nil {
        return err
    }

    return os.WriteFile(destPath, data, 0644)
}

// SendFile 发送文件到服务（上传 + 通知）
func (c *Client) SendFile(filePath, fileName, targetService string) error {
    fileID, err := c.UploadFile(filePath, fileName)
    if err != nil {
        return err
    }

    // 发送文件元数据通知
    metadata := map[string]interface{}{
        "file_id":   fileID,
        "file_name": fileName,
        "to":        targetService,
    }

    return c.Publish("file.transfer", metadata)
}
```

**Step 3: 运行测试**

```bash
cd sdk/go/client
go test -v -run TestFileTransfer
```

**Step 4: 提交**

```bash
git add sdk/go/client/file.go sdk/go/client/file_test.go
git commit -m "feat: implement file transfer with NATS Object Store"
```

---

## Phase 6: 示例和文档

### Task 6.1: 创建完整示例

**文件:**
- Create: `examples/pubsub-demo/main.go`
- Create: `examples/state-demo/main.go`
- Create: `examples/file-transfer-demo/main.go`

**Step 1: 创建发布订阅示例**

文件: `examples/pubsub-demo/main.go`

```go
package main

import (
    "fmt"
    "time"
    "github.com/LiteHomeLab/light_link/sdk/go/client"
)

func main() {
    cli, _ := client.NewClient("nats://localhost:4222", nil)
    defer cli.Close()

    // 订阅
    cli.Subscribe("events.user", func(data map[string]interface{}) {
        fmt.Printf("Received: %v\n", data)
    })

    // 发布
    for i := 0; i < 5; i++ {
        cli.Publish("events.user", map[string]interface{}{
            "event":   "user_login",
            "user_id": fmt.Sprintf("U%03d", i),
        })
        time.Sleep(500 * time.Millisecond)
    }

    time.Sleep(1 * time.Second)
}
```

**Step 2: 创建状态管理示例**

文件: `examples/state-demo/main.go`

```go
package main

import (
    "fmt"
    "time"
    "github.com/LiteHomeLab/light_link/sdk/go/client"
)

func main() {
    cli, _ := client.NewClient("nats://localhost:4222", nil)
    defer cli.Close()

    // 监听状态变化
    stop, _ := cli.WatchState("device.sensor01", func(state map[string]interface{}) {
        fmt.Printf("State updated: %v\n", state)
    })
    defer stop()

    // 更新状态
    for i := 0; i < 3; i++ {
        cli.SetState("device.sensor01", map[string]interface{}{
            "temperature": 20.0 + float64(i),
            "humidity":    50 + i,
        })
        time.Sleep(500 * time.Millisecond)
    }

    // 获取最新状态
    state, _ := cli.GetState("device.sensor01")
    fmt.Printf("Latest state: %v\n", state)
}
```

**Step 3: 创建文件传输示例**

文件: `examples/file-transfer-demo/main.go`

```go
package main

import (
    "fmt"
    "github.com/LiteHomeLab/light_link/sdk/go/client"
)

func main() {
    cli, _ := client.NewClient("nats://localhost:4222", nil)
    defer cli.Close()

    // 上传文件
    fileID, err := cli.UploadFile("./test.txt", "test.txt")
    if err != nil {
        fmt.Println("Upload error:", err)
        return
    }
    fmt.Println("Uploaded, file ID:", fileID)

    // 下载文件
    err = cli.DownloadFile(fileID, "./downloaded.txt")
    if err != nil {
        fmt.Println("Download error:", err)
        return
    }
    fmt.Println("Downloaded successfully")
}
```

**Step 4: 创建文档**

文件: `docs/getting-started.md`

```markdown
# LightLink 快速开始

## 安装 NATS Server

\`\`\`bash
# Windows
nats-server -config deploy/nats/nats-server.conf
\`\`\`

## 生成 TLS 证书

参考 `deploy/nats/tls/README.md`

## 客户端使用

\`\`\`go
import "github.com/LiteHomeLab/light_link/sdk/go/client"

cli, _ := client.NewClient("nats://localhost:4222", nil)
defer cli.Close()

// RPC 调用
result, _ := cli.Call("user-service", "getUser", map[string]interface{}{
    "user_id": "U001",
})

// 发布订阅
cli.Publish("events.user", map[string]interface{}{"msg": "hello"})
cli.Subscribe("events.user", func(data map[string]interface{}) {
    fmt.Println(data)
})

// 状态管理
cli.SetState("device.temp", map[string]interface{}{"value": 25.5})
state, _ := cli.GetState("device.temp")

// 文件传输
fileID, _ := cli.UploadFile("./data.csv", "data.csv")
cli.DownloadFile(fileID, "./downloaded.csv")
\`\`\`

## 服务端使用

\`\`\`go
import "github.com/LiteHomeLab/light_link/sdk/go/service"

svc, _ := service.NewService("my-service", "nats://localhost:4222", nil)

svc.RegisterRPC("add", func(args map[string]interface{}) (map[string]interface{}, error) {
    a := int(args["a"].(float64))
    b := int(args["b"].(float64))
    return map[string]interface{}{"sum": a + b}, nil
})

svc.Start()
\`\`\`
```

**Step 5: 提交**

```bash
git add examples/ docs/
git commit -m "docs: add examples and getting started guide"
```

---

## Phase 7: Python SDK (基础)

### Task 7.1: 创建 Python SDK 基础

**文件:**
- Create: `sdk/python/lightlink/__init__.py`
- Create: `sdk/python/lightlink/client.py`

**Step 1: 创建 Python 包结构**

```bash
mkdir -p sdk/python/lightlink
```

**Step 2: 创建客户端**

文件: `sdk/python/lightlink/client.py`

```python
import asyncio
import json
from nats.aio.client import Client as NATSClient
from nats.errors import TimeoutError

class Client:
    def __init__(self, url="nats://localhost:4222"):
        self.url = url
        self.nc = None

    async def connect(self):
        self.nc = NATSClient()
        await self.nc.connect(self.url)

    async def close(self):
        if self.nc:
            await self.nc.close()

    async def call(self, service, method, args, timeout=5.0):
        """RPC 调用"""
        subject = f"$SRV.{service}.{method}"
        request = {
            "id": str(uuid.uuid4()),
            "method": method,
            "args": args
        }

        try:
            msg = await self.nc.request(
                subject,
                json.dumps(request).encode(),
                timeout=timeout
            )
            response = json.loads(msg.data.decode())
            if not response.get("success"):
                raise Exception(response.get("error"))
            return response.get("result")
        except TimeoutError:
            raise Exception("RPC timeout")

    async def publish(self, subject, data):
        """发布消息"""
        await self.nc.publish(subject, json.dumps(data).encode())

    async def subscribe(self, subject, handler):
        """订阅消息"""
        async def cb(msg):
            data = json.loads(msg.data.decode())
            await handler(data)

        await self.nc.subscribe(subject, cb)
```

**Step 3: 创建 Python 示例**

文件: `examples/python-demo/main.py`

```python
import asyncio
from lightlink import Client

async def main():
    client = Client()
    await client.connect()

    # RPC 调用
    result = await client.call("user-service", "getUser", {"user_id": "U001"})
    print("Result:", result)

    # 发布
    await client.publish("events.test", {"msg": "hello"})

    # 订阅
    await client.subscribe("events.test", lambda data: print(data))

    await asyncio.sleep(1)
    await client.close()

asyncio.run(main())
```

**Step 4: 提交**

```bash
git add sdk/python/ examples/python-demo/
git commit -m "feat: add Python SDK foundation"
```

---

## 总结

完成以上步骤后，LightLink 框架将包含：

1. **Go SDK** - 完整功能的参考实现
2. **Python SDK** - 基础客户端实现
3. **RPC 调用** - 服务间远程调用
4. **发布订阅** - 消息通知和广播
5. **状态管理** - NATS KV 实现
6. **文件传输** - NATS Object Store 实现
7. **TLS 认证** - 双向证书认证
8. **示例项目** - 演示各功能用法
9. **文档** - 快速开始指南
