# LightLink TLS 证书管理

## 概述

此目录包含 LightLink NATS 通信的 TLS 证书模板和参考文档。

证书生成脚本已移至独立仓库：**[create_tls](https://github.com/LiteHomeLab/create_tls)**

## 快速开始

### 生成新证书

证书生成现在由 `create_tls` 子模块管理：

```batch
cd deploy/nats/create_tls
setup-certs.bat
```

这将生成一个带时间戳的文件夹 `certs-YYYY-MM-DD_HH-MM/`，包含：

```
certs-YYYY-MM-DD_HH-MM/
├── nats-server/     # 部署到 NATS 服务器
│   ├── ca.crt
│   ├── server.crt
│   ├── server.key
│   └── README.txt
└── client/          # 部署到所有客户端服务
    ├── ca.crt
    ├── client.crt
    ├── client.key
    └── README.txt
```

### 部署证书

1. **NATS 服务器**：将 `nats-server/` 文件夹复制到你的 NATS 服务器
2. **客户端服务**：将 `client/` 文件夹复制到你的服务目录

## 外部仓库

- **仓库地址**：[git@github.com:LiteHomeLab/create_tls.git](https://github.com/LiteHomeLab/create_tls)
- **位置**：`deploy/nats/create_tls/`（Git 子模块）
- **用途**：证书生成脚本和文档

## 证书架构

| 证书 | CN（通用名称） | 使用者 |
|-------------|------------------|---------|
| CA 根证书 | `LightLink CA` | 签署所有证书 |
| NATS 服务器证书 | `nats-server` | 仅 NATS 服务器 |
| 客户端证书 | `lightlink-client` | 所有客户端服务（共享） |

## 客户端连接配置

所有 SDK 现在支持**自动证书发现**！只需将 `client/` 文件夹复制到项目目录即可使用。

### 使用自动发现（推荐）

**Go SDK:**
```go
// 方式1: 使用 WithAutoTLS() 自动发现证书
client, err := client.NewClient("tls://172.18.200.47:4222", client.WithAutoTLS())

// 方式2: 使用 WithTLS() 手动指定
tlsConfig := &client.TLSConfig{
    CaFile:     "client/ca.crt",
    CertFile:   "client/client.crt",
    KeyFile:    "client/client.key",
    ServerName: "nats-server",
}
client, err := client.NewClient("tls://172.18.200.47:4222", client.WithTLS(tlsConfig))
```

**Python SDK:**
```python
# 方式1: 使用 auto_tls=True 自动发现证书
client = Client(auto_tls=True)
await client.connect()

# 方式2: 使用 TLSConfig 手动指定
tls_config = TLSConfig(
    ca_file="client/ca.crt",
    cert_file="client/client.crt",
    key_file="client/client.key",
    server_name="nats-server"
)
client = Client(tls_config=tls_config)
```

**C# SDK:**
```csharp
// 方式1: 使用 CertDiscovery.GetAutoTLSConfig() 自动发现
var tlsConfig = CertDiscovery.GetAutoTLSConfig();
var opts = ConnectionFactory.GetDefaultOptions();
opts.Url = "tls://172.18.200.47:4222";
opts.SetCertificate(tlsConfig.CaFile, tlsConfig.CertFile, tlsConfig.KeyFile);

// 方式2: 手动创建 TLSConfig
var tlsConfig = new TLSConfig {
    CaFile = "client/ca.crt",
    CertFile = "client/client.crt",
    KeyFile = "client/client.key",
    ServerName = "nats-server"
};
```

### 服务端自动发现

**Go SDK:**
```go
// 使用 WithServiceAutoTLS() 自动发现服务器证书
service, err := service.NewService("my-service", "tls://172.18.200.47:4222", service.WithServiceAutoTLS())
```

**Python SDK:**
```python
# 使用 auto_tls=True 自动发现服务器证书
service = Service(name="my-service", nats_url="tls://172.18.200.47:4222", auto_tls=True)
await service.start()
```

### 证书搜索路径

SDK 会在以下位置搜索证书：
- `./client` 或 `./nats-server` (当前目录)
- `../client` 或 `../nats-server` (上级目录)
- `../../client` 或 `../../nats-server` (上上级目录)

### 完整部署流程

1. **生成证书:**
   ```batch
   cd deploy/nats/create_tls
   setup-certs.bat
   ```

2. **部署证书:**
   - 将 `certs-TIMESTAMP/nats-server/` 复制到 `deploy/nats/` 目录
   - 将 `certs-TIMESTAMP/client/` 复制到你的服务项目目录

3. **启动 NATS 服务器:**
   ```batch
   nats-server -c deploy/nats/nats-server.conf
   ```

4. **启动服务** (自动发现证书):
   ```bash
   # Go 服务
   go run main.go

   # Python 服务
   python main.py

   # C# 服务
   dotnet run
   ```

## 环境变量（可选）

```bash
set NATS_URL=tls://172.18.200.47:4222
set TLS_CA=client/ca.crt
set TLS_CERT=client/client.crt
set TLS_KEY=client/client.key
set TLS_SERVER_NAME=nats-server
```

## 安全注意事项

- **私钥文件（.key 文件）必须妥善保管！**
- 切勿将 .key 文件提交到版本控制系统
- 使用安全渠道分发证书包
- CA 证书（ca.crt）是公开的，可以自由分发

## 更新子模块

将 `create_tls` 子模块更新到最新版本：

```batch
cd deploy/nats/create_tls
git pull origin main
cd ../..
git add deploy/nats/create_tls
git commit -m "chore: update create_tls submodule"
```

## 目录结构

```
deploy/nats/
├── tls/              # 本目录（模板和文档）
│   └── README.md
└── create_tls/       # Git 子模块 -> git@github.com:LiteHomeLab/create_tls.git
    ├── setup-certs.bat
    ├── find-openssl.bat
    ├── setup-certs-dependency.txt
    └── README.md
```

## 故障排查

如果 `setup-certs.bat` 无法找到 OpenSSL，请运行：

```batch
cd deploy/nats/create_tls
find-openssl.bat
```

此诊断工具将搜索系统中的 OpenSSL。

## 另请参阅

- [create_tls 仓库](https://github.com/LiteHomeLab/create_tls)
- [项目文档](../../../docs/)
