# 自动证书发现功能 - 实现计划

## 概述

本计划为 Go、C# 和 Python SDK 添加自动证书发现功能，允许用户将生成的 `client/` 或 `nats-server/` 证书文件夹复制到项目目录即可使用，无需手动配置证书路径。

## 证书结构回顾

```
client/              # 客户端服务使用
├── ca.crt          # CA 根证书
├── client.crt      # 客户端证书
└── client.key      # 客户端私钥

nats-server/        # NATS 服务器使用
├── ca.crt          # CA 根证书
├── server.crt      # 服务器证书
└── server.key      # 服务器私钥
```

---

## 第一部分: Go SDK 实现

### 任务 1.1: 在 `sdk/go/types/types.go` 添加证书发现函数

**文件**: `sdk/go/types/types.go`

**修改内容**:

在文件末尾添加证书发现相关的常量和函数：

```go
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
```

---

### 任务 1.2: 在 `sdk/go/client/connection.go` 添加 WithAutoTLS 选项

**文件**: `sdk/go/client/connection.go`

**修改内容**:

添加 `WithAutoTLS()` 选项函数：

```go
// WithAutoTLS 自动发现并使用 TLS 证书
// 在 ./client 目录查找证书
func WithAutoTLS() Option {
    return func(c *Client) error {
        result, err := types.DiscoverClientCerts()
        if err != nil {
            return fmt.Errorf("auto-discover TLS failed: %w", err)
        }
        c.tlsConfig = &TLSConfig{
            CaFile:     result.CaFile,
            CertFile:   result.CertFile,
            KeyFile:    result.KeyFile,
            ServerName: result.ServerName,
        }
        return nil
    }
}
```

---

### 任务 1.3: 在 `sdk/go/service/service.go` 添加服务器证书自动发现

**文件**: `sdk/go/service/service.go`

**修改内容**:

```go
// WithServiceAutoTLS 自动发现并使用服务器 TLS 证书
// 在 ./nats-server 目录查找证书
func WithServiceAutoTLS() ServiceOption {
    return func(s *Service) error {
        result, err := types.DiscoverServerCerts()
        if err != nil {
            return fmt.Errorf("auto-discover server TLS failed: %w", err)
        }
        s.tlsConfig = &client.TLSConfig{
            CaFile:     result.CaFile,
            CertFile:   result.CertFile,
            KeyFile:    result.KeyFile,
            ServerName: result.ServerName,
        }
        return nil
    }
}
```

---

### 任务 1.4: 为 Go SDK 添加测试

**文件**: `sdk/go/types/types_test.go` (新建或追加)

---

## 第二部分: C# SDK 实现

### 任务 2.1: 创建 TLS 配置类和证书发现器

**文件**: `sdk/csharp/LightLink/TLSConfig.cs` (新建)

**内容**:

```csharp
using System;
using System.IO;

namespace LightLink
{
    public class TLSConfig
    {
        public string CaFile { get; set; } = "";
        public string CertFile { get; set; } = "";
        public string KeyFile { get; set; } = "";
        public string ServerName { get; set; } = "nats-server";
    }

    public class CertDiscoveryResult
    {
        public string CaFile { get; set; } = "";
        public string CertFile { get; set; } = "";
        public string KeyFile { get; set; } = "";
        public string ServerName { get; set; } = "nats-server";
        public bool Found { get; set; }
    }

    public static class CertDiscovery
    {
        private const string DefaultClientCertDir = "./client";
        private const string DefaultServerCertDir = "./nats-server";
        private const string DefaultServerName = "nats-server";

        public static CertDiscoveryResult DiscoverClientCerts()
        {
            var searchPaths = new[] { DefaultClientCertDir, "../client", "../../client" };
            foreach (var path in searchPaths)
            {
                var result = CheckCertDirectory(path, "client");
                if (result.Found) return result;
            }
            return new CertDiscoveryResult { Found = false };
        }

        public static CertDiscoveryResult DiscoverServerCerts()
        {
            var searchPaths = new[] { DefaultServerCertDir, "../nats-server", "../../nats-server" };
            foreach (var path in searchPaths)
            {
                var result = CheckCertDirectory(path, "server");
                if (result.Found) return result;
            }
            return new CertDiscoveryResult { Found = false };
        }

        private static CertDiscoveryResult CheckCertDirectory(string dir, string certType)
        {
            var certFile = Path.Combine(dir, certType == "client" ? "client.crt" : "server.crt");
            var keyFile = Path.Combine(dir, certType == "client" ? "client.key" : "server.key");
            var caFile = Path.Combine(dir, "ca.crt");

            if (File.Exists(caFile) && File.Exists(certFile) && File.Exists(keyFile))
            {
                return new CertDiscoveryResult
                {
                    CaFile = caFile,
                    CertFile = certFile,
                    KeyFile = keyFile,
                    ServerName = DefaultServerName,
                    Found = true
                };
            }
            return new CertDiscoveryResult { Found = false };
        }

        public static TLSConfig ToTLSConfig(CertDiscoveryResult result)
        {
            return new TLSConfig
            {
                CaFile = result.CaFile,
                CertFile = result.CertFile,
                KeyFile = result.KeyFile,
                ServerName = result.ServerName
            };
        }
    }
}
```

---

### 任务 2.2: 更新 Service 类支持自动 TLS

**文件**: `sdk/csharp/LightLink/Service.cs`

---

### 任务 2.3: 为 C# SDK 添加测试

**文件**: `sdk/csharp/LightLink.Tests/CertDiscoveryTests.cs` (新建)

---

## 第三部分: Python SDK 实现

### 任务 3.1: 在 client.py 添加证书自动发现

**文件**: `sdk/python/lightlink/client.py`

**修改内容**:

```python
import os
import ssl
from pathlib import Path
from typing import Optional

# 证书发现相关常量
DEFAULT_CLIENT_CERT_DIR = "./client"
DEFAULT_SERVER_CERT_DIR = "./nats-server"
DEFAULT_SERVER_NAME = "nats-server"


class CertDiscoveryResult:
    """证书发现结果"""
    def __init__(self, ca_file: str, cert_file: str, key_file: str, server_name: str, found: bool):
        self.ca_file = ca_file
        self.cert_file = cert_file
        self.key_file = key_file
        self.server_name = server_name
        self.found = found


def discover_client_certs() -> CertDiscoveryResult:
    """自动发现客户端证书"""
    search_paths = [DEFAULT_CLIENT_CERT_DIR, "../client", "../../client"]

    for base_path in search_paths:
        result = _check_cert_directory(base_path, "client")
        if result.found:
            return result

    raise FileNotFoundError(
        f"Client certificates not found in search paths: {search_paths}. "
        f"Please copy the 'client/' folder from generated certificates to your project."
    )


def discover_server_certs() -> CertDiscoveryResult:
    """自动发现服务器证书"""
    search_paths = [DEFAULT_SERVER_CERT_DIR, "../nats-server", "../../nats-server"]

    for base_path in search_paths:
        result = _check_cert_directory(base_path, "server")
        if result.found:
            return result

    raise FileNotFoundError(
        f"Server certificates not found in search paths: {search_paths}. "
        f"Please copy the 'nats-server/' folder from generated certificates to your project."
    )


def _check_cert_directory(dir_path: str, cert_type: str) -> CertDiscoveryResult:
    """检查目录中的证书文件是否存在"""
    cert_file = os.path.join(dir_path, f"{cert_type}.crt")
    key_file = os.path.join(dir_path, f"{cert_type}.key")
    ca_file = os.path.join(dir_path, "ca.crt")

    if os.path.isfile(ca_file) and os.path.isfile(cert_file) and os.path.isfile(key_file):
        return CertDiscoveryResult(
            ca_file=ca_file,
            cert_file=cert_file,
            key_file=key_file,
            server_name=DEFAULT_SERVER_NAME,
            found=True
        )
    return CertDiscoveryResult("", "", "", "", False)


class TLSConfig:
    """TLS configuration"""
    def __init__(self, ca_file, cert_file, key_file, server_name=None):
        self.ca_file = ca_file
        self.cert_file = cert_file
        self.key_file = key_file
        self.server_name = server_name or DEFAULT_SERVER_NAME

    @classmethod
    def from_auto_discovery(cls) -> 'TLSConfig':
        """从自动发现创建 TLS 配置"""
        result = discover_client_certs()
        return cls(
            ca_file=result.ca_file,
            cert_file=result.cert_file,
            key_file=result.key_file,
            server_name=result.server_name
        )
```

---

### 任务 3.2: 在 service.py 添加服务器证书自动发现

**文件**: `sdk/python/lightlink/service.py`

---

### 任务 3.3: 为 Python SDK 添加测试

**文件**: `sdk/python/tests/test_tls_discovery.py` (新建)

---

## 第四部分: 文档更新

### 任务 4.1: 更新主 README 文档

### 任务 4.2: 为每个 SDK 添加使用示例

---

## 实施顺序

```
第一阶段: Go SDK
├── 任务 1.1: types.go 证书发现函数
├── 任务 1.4: types_test.go 测试
├── 任务 1.2: client.go WithAutoTLS 选项
└── 任务 1.3: service.go WithServiceAutoTLS 选项

第二阶段: C# SDK
├── 任务 2.1: TLSConfig.cs 新建
├── 任务 2.3: CertDiscoveryTests.cs 测试
└── 任务 2.2: Service.cs 更新

第三阶段: Python SDK
├── 任务 3.1: client.py 更新
├── 任务 3.3: tls_discovery 测试
└── 任务 3.2: service.py 更新

第四阶段: 文档和示例
├── 任务 4.1: README 更新
└── 任务 4.2: 示例代码
```

---

## 向后兼容性

所有实现保持完全向后兼容：

- Go SDK: 保留原有 API，添加新的选项函数
- C# SDK: 添加新的静态工厂方法，保留原有构造函数
- Python SDK: 通过新的 `auto_tls` 参数启用，默认为 false

---

## 关键文件

- `sdk/go/types/types.go` - 添加证书发现核心函数
- `sdk/go/client/connection.go` - 添加 WithAutoTLS() 选项
- `sdk/csharp/LightLink/TLSConfig.cs` - 新建 C# 证书发现类
- `sdk/python/lightlink/client.py` - 添加 auto_tls 参数
- `sdk/python/lightlink/service.py` - 添加服务器证书自动发现
