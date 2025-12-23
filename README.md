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
│   └── python-demo/        # Python 示例
├── deploy/                 # 部署配置
│   └── nats/               # NATS 服务器配置和 TLS 证书
├── docs/                   # 项目文档
│   ├── getting-started.md  # 快速开始指南
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
