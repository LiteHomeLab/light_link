# LightLink

基于 NATS 的多语言后端服务通信框架，支持 C++、Python、C#、Go、JS 等语言编写的服务在内网多台服务器间通信。

## 功能特性

| 功能 | 说明 |
|------|------|
| **RPC 远程调用** | 服务间函数调用，支持请求/响应模式 |
| **消息发布/订阅** | 实时消息通知和广播，支持多订阅者 |
| **状态保留** | 类似 MQTT retain 的最新状态功能（基于 NATS KV） |
| **大文件传输** | 最大 1GB 文件传输（基于 NATS Object Store） |
| **TLS 证书认证** | 双向 TLS 认证 + 用户权限配置 |
| **服务管理平台** | Web 控制台，支持服务注册、发现、调试和监控 |

## 服务管理平台

LightLink 服务管理平台是一个完整的 Web 控制台，提供：

- **服务注册与发现** - 自动服务注册，元数据管理
- **服务监控** - 实时在线/离线状态，心跳检测
- **RPC 调试** - Web 界面直接调用服务方法
- **事件追踪** - 服务生命周期事件记录

### 启动服务管理平台

```bash
# 1. 启动 NATS 服务器
nats-server -config deploy/nats/nats-server.conf

# 2. 启动 Web 控制台后端
cd console/server
go run main.go

# 3. 启动 Web 前端（开发模式）
cd console/web
npm install
npm run dev

# 4. 访问控制台
# 浏览器打开: http://localhost:5173
# 默认账号: admin / admin123
```

### 注册带元数据的服务

```go
package main

import (
    "github.com/LiteHomeLab/light_link/sdk/go/service"
    "github.com/LiteHomeLab/light_link/sdk/go/types"
)

func main() {
    // 创建服务
    svc, _ := service.NewService("my-service", "nats://localhost:4222", nil)

    // 定义方法元数据
    methodMeta := &types.MethodMetadata{
        Name:        "add",
        Description: "Add two numbers",
        Params: []types.ParameterMetadata{
            {Name: "a", Type: "number", Required: true, Description: "First number"},
            {Name: "b", Type: "number", Required: true, Description: "Second number"},
        },
        Returns: []types.ReturnMetadata{
            {Name: "sum", Type: "number", Description: "The sum"},
        },
        Example: &types.ExampleMetadata{
            Input:  map[string]any{"a": 10, "b": 20},
            Output: map[string]any{"sum": 30},
        },
    }

    // 注册方法带元数据
    svc.RegisterMethodWithMetadata("add", addHandler, methodMeta)

    // 注册服务元数据
    svc.RegisterMetadata(&types.ServiceMetadata{
        Name:        "my-service",
        Version:     "v1.0.0",
        Description: "My awesome service",
        Author:      "Your Name",
        Tags:        []string{"demo", "example"},
    })

    // 启动服务
    svc.Start()
    select {}
}

func addHandler(args map[string]interface{}) (map[string]interface{}, error) {
    a := args["a"].(float64)
    b := args["b"].(float64)
    return map[string]interface{}{"sum": a + b}, nil
}
```

更多示例请查看：
- [metadata-demo](examples/metadata-demo/main.go) - 带元数据的服务示例
- [metadata-client](examples/metadata-client/main.go) - 客户端调用示例

## 支持的语言

| 语言 | 状态 | SDK 路径 |
|------|------|----------|
| **Go** | ✅ 完成 | `sdk/go/` |
| **Python** | ✅ 完成 | `sdk/python/` |
| **C#** | ✅ 完成 | `sdk/csharp/` |
| **C++** | 🚧 基础实现 | `sdk/cpp/` |
| **JavaScript** | 📋 计划中 | `sdk/js/` |

## 快速开始

### 前置要求

- **NATS Server** 2.10+ （支持 JetStream）
- **Go** 1.21+ （开发 Go SDK）
- **Python** 3.8+ （开发 Python SDK）
- **.NET** 6.0+ （开发 C# SDK）
- **CMake** 3.15+ （开发 C++ SDK）

### 1. 启动 NATS 服务器

```bash
nats-server -config deploy/nats/nats-server.conf
```

### 2. 生成 TLS 证书（可选）

```bash
cd deploy/nats/tls
generate-certs.bat
```

### 3. 运行 Go 示例

```bash
# RPC 演示
go run examples/rpc-demo/main.go

# 发布订阅演示
go run examples/pubsub-demo/main.go

# 状态管理演示
go run examples/state-demo/main.go

# 文件传输演示
go run examples/file-transfer-demo/main.go
```

### 4. 运行测试

```bash
# Go SDK 测试
go test ./sdk/go/...

# Python SDK 测试
cd sdk/python
pip install -r requirements.txt
pytest

# C# SDK 测试
cd sdk/csharp/LightLink.Tests
dotnet test
```

## 目录结构

```
light_link/
├── sdk/                    # 多语言 SDK 实现
│   ├── go/                 # Go SDK（参考实现）
│   ├── python/             # Python SDK
│   ├── csharp/             # C# SDK
│   ├── cpp/                # C++ SDK
│   └── js/                 # JavaScript SDK（待实现）
├── examples/               # 示例代码
│   ├── rpc-demo/           # RPC 调用示例
│   ├── pubsub-demo/        # 发布订阅示例
│   ├── state-demo/         # 状态管理示例
│   ├── file-transfer-demo/ # 文件传输示例
│   ├── metadata-demo/      # 带元数据的服务示例
│   ├── metadata-client/    # 元数据服务客户端示例
│   └── python-demo/        # Python 示例
├── console/                # Web 服务管理平台
│   ├── server/             # Go 后端服务
│   │   ├── api/            # REST API 处理器
│   │   ├── auth/           # JWT 认证
│   │   ├── manager/        # 服务管理器
│   │   ├── storage/        # SQLite 数据库
│   │   ├── ws/             # WebSocket Hub
│   │   └── proxy/          # RPC 调用代理
│   └── web/                # Vue 3 前端
│       ├── src/
│       │   ├── api/        # API 客户端
│       │   ├── views/      # 页面组件
│       │   ├── stores/     # Pinia 状态管理
│       │   └── router/     # Vue Router
│       └── package.json
├── deploy/                 # 部署配置
│   └── nats/               # NATS 服务器配置和 TLS 证书
├── docs/                   # 项目文档
│   ├── getting-started.md  # 快速开始指南
│   ├── service-management-platform.md # 服务管理平台设计
│   └── sdk-api-comparison.md # SDK API 对比
├── CLAUDE.md               # 项目开发规则
└── README.md               # 本文档
```

## 使用指南

### Go SDK

```go
package main

import "github.com/LiteHomeLab/light_link/sdk/go/client"

func main() {
    // 连接到 NATS 服务器
    cli, _ := client.Connect(client.Config{
        URLs: []string{"nats://localhost:4222"},
    })

    // RPC 调用
    var result string
    cli.Call("service.method", "request", &result)

    // 发布消息
    cli.Publish("topic", "message")

    // 订阅消息
    cli.Subscribe("topic", func(msg []byte) {
        println(string(msg))
    })

    // 设置状态
    cli.SetState("key", "value")

    // 获取状态
    var value string
    cli.GetState("key", &value)

    // 上传文件
    cli.UploadFile("bucket", "remote.txt", "local.txt")

    // 下载文件
    cli.DownloadFile("bucket", "remote.txt", "local.txt")
}
```

### Python SDK

```python
from lightlink import Client

# 连接到 NATS 服务器
client = Client(urls=["nats://localhost:4222"])

# RPC 调用
result = client.call("service.method", request="data")

# 发布消息
client.publish("topic", "message")

# 订阅消息
def callback(msg):
    print(msg)

client.subscribe("topic", callback)

# 设置状态
client.set_state("key", "value")

# 获取状态
value = client.get_state("key")

# 上传文件
client.upload_file("bucket", "remote.txt", "local.txt")

# 下载文件
client.download_file("bucket", "remote.txt", "local.txt")
```

### C# SDK

```csharp
using LightLink;

// 连接到 NATS 服务器
var client = new Client(new ClientConfig {
    Urls = new[] { "nats://localhost:4222" }
});

// RPC 调用
var result = await client.CallAsync<string>("service.method", "request");

// 发布消息
await client.PublishAsync("topic", "message");

// 订阅消息
await client.SubscribeAsync("topic", (msg) => {
    Console.WriteLine(msg);
});

// 设置状态
await client.SetStateAsync("key", "value");

// 获取状态
var value = await client.GetStateAsync<string>("key");

// 上传文件
await client.UploadFileAsync("bucket", "remote.txt", "local.txt");

// 下载文件
await client.DownloadFileAsync("bucket", "remote.txt", "local.txt");
```

## 配置

### NATS 服务器配置

配置文件位于 `deploy/nats/nats-server.conf`：

```conf
# 监听端口
port: 4222

# JetStream 支持
jetstream: {
    store_dir: "./data"
}

# TLS 配置
tls: {
    cert_file: "./tls/server/server-cert.pem"
    key_file: "./tls/server/server-key.pem"
    ca_file: "./tls/ca-cert.pem"
    verify: true
}

# 连接限制
max_connections: 1000
max_subs: 1000
```

### 客户端配置

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `urls` | []string | ["nats://localhost:4222"] | NATS 服务器地址 |
| `username` | string | "" | 用户名（TLS 认证） |
| `password` | string | "" | 密码（TLS 认证） |
| `tls_cert` | string | "" | 客户端证书路径 |
| `tls_key` | string | "" | 客户端私钥路径 |
| `tls_ca` | string | "" | CA 证书路径 |

## 开发

### TDD 开发模式

项目遵循测试驱动开发（TDD）原则：

1. 先编写测试用例
2. 实现功能代码
3. 确保所有测试通过
4. 提交代码

### 添加新语言 SDK

1. 在 `sdk/` 下创建语言目录
2. 参考 Go SDK 实现以下功能：
   - 连接管理 (`connection`)
   - RPC 调用 (`rpc`)
   - 发布订阅 (`pubsub`)
   - 状态管理 (`state`)
   - 文件传输 (`file`)
3. 编写单元测试
4. 添加示例代码
5. 更新文档

### 提交规范

每个功能完成后提交一次代码：

```bash
git add .
git commit -m "feat: add JavaScript SDK basic implementation"
```

## 文档

- [快速开始指南](docs/getting-started.md)
- [SDK API 对比](docs/sdk-api-comparison.md)
- [TLS 证书生成](deploy/nats/tls/README.md)
- [开发规则](CLAUDE.md)

## 许可证

[MIT License](LICENSE)

## 贡献

欢迎提交 Issue 和 Pull Request！

## 联系方式

- 项目地址: https://github.com/LiteHomeLab/light_link
