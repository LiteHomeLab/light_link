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

### Go SDK

```go
tlsConfig := &client.TLSConfig{
    CaFile:     "client/ca.crt",
    CertFile:   "client/client.crt",
    KeyFile:    "client/client.key",
    ServerName: "nats-server",  // 必须与服务器证书 CN 匹配
}
client, err := client.NewClient("tls://172.18.200.47:4222", tlsConfig)
```

### Python SDK

```python
tls_config = TLSConfig(
    ca_file="client/ca.crt",
    cert_file="client/client.crt",
    key_file="client/client.key",
    server_name="nats-server"  # 必须与服务器证书 CN 匹配
)
client = Client(tls_config=tls_config)
```

### C# SDK

```csharp
Options opts = ConnectionFactory.GetDefaultOptions();
opts.Url = "tls://172.18.200.47:4222";
opts.SSL = true;
opts.SetCertificate("client/ca.crt", "client/client.crt", "client/client.key");
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
