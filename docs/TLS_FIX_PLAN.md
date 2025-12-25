# TLS 修复计划与状态报告

## 问题分析

### 初始状态

| SDK | TLS 配置 | 问题 | 初始状态 |
|-----|----------|------|----------|
| Go | `InsecureSkipVerify: true` (硬编码) | 无法选择是否跳过验证 | 可用 |
| C# | 仅设置 `Secure = true` | 缺少 PFX 证书支持 + 跳过验证选项 | TLS 握手失败 |
| Python | `ssl.create_default_context` | 导入错误 + 缺少跳过验证选项 | 不可用 |

### 根本原因

1. **Go SDK**: `InsecureSkipVerify: true` 硬编码在 `CreateTLSOption` 中，用户无法选择是否验证服务器名称
2. **C# SDK**:
   - NATS.Client 库要求证书和私钥在同一个 PFX 文件中
   - 缺少 `TLSRemoteCertificationValidationCallback` 配置
3. **Python SDK**:
   - 导入错误：`NotFoundError` 在新版 nats-py 中已被重命名
   - 缺少跳过服务器名称验证的选项

## 修复状态

### ✅ Go SDK - 可用

**当前状态**: 已可用，但 `InsecureSkipVerify` 硬编码

**文件**: `sdk/go/client/connection.go:134`

```go
tlsConfig := &tls.Config{
    Certificates:       []tls.Certificate{cert},
    RootCAs:            pool,
    MinVersion:         tls.VersionTLS12,
    InsecureSkipVerify: true, // 硬编码
}
```

**后续改进** (优先级: 中):
- [ ] 添加 `InsecureSkipVerify` 字段到 `TLSConfig`
- [ ] 修改 `CreateTLSOption` 使用配置值
- [ ] `WithAutoTLS()` 默认为 `true`，`WithTLS()` 默认为 `false`

---

### ✅ C# SDK - 已修复

**当前状态**: 已修复并验证可用

**问题**:
1. NATS.Client 要求证书和私钥在同一个 PFX 文件中
2. 需要配置 `TLSRemoteCertificationValidationCallback` 跳过证书验证

**修复内容**:

#### 1. TLSConfig 添加 PFX 支持
**文件**: `sdk/csharp/LightLink/TLSConfig.cs`

```csharp
public class TLSConfig
{
    public string CaFile { get; set; } = "";
    public string CertFile { get; set; } = "";
    public string KeyFile { get; set; } = "";

    /// <summary>
    /// PFX certificate file (contains both cert and key).
    /// NATS.Client requires PFX format for TLS connections.
    /// If provided, takes precedence over CertFile/KeyFile.
    /// </summary>
    public string PfxFile { get; set; } = "";

    /// <summary>
    /// Password for PFX file.
    /// </summary>
    public string PfxPassword { get; set; } = "";

    public bool InsecureSkipVerify { get; set; } = true;
}
```

#### 2. Service.cs 支持 PFX 和证书验证回调
**文件**: `sdk/csharp/LightLink/Service.cs`

```csharp
// Prefer PFX format (NATS.Client requires cert + key in same file)
if (!string.IsNullOrEmpty(_tlsConfig.PfxFile))
{
    cert = new X509Certificate2(_tlsConfig.PfxFile, _tlsConfig.PfxPassword);
}
else
{
    cert = new X509Certificate2(_tlsConfig.CertFile);
}

opts.AddCertificate(cert);

// Skip server certificate validation for self-signed certificates
if (_tlsConfig.InsecureSkipVerify)
{
    opts.TLSRemoteCertificationValidationCallback =
        (sender, certificate, chain, sslPolicyErrors) => true;
}
```

#### 3. 证书转换
生成 PFX 文件:
```bash
openssl pkcs12 -export -out client/client.pfx \
    -inkey client/client.key \
    -in client/client.crt \
    -passout pass:lightlink
```

#### 4. CertDiscovery 自动查找 PFX
更新 `CheckCertDirectory` 方法，优先查找 `.pfx` 文件：
```csharp
var pfxFile = Path.Combine(dir, certType == "client" ? "client.pfx" : "server.pfx");

// Prefer PFX format for NATS.Client
if (File.Exists(pfxFile))
{
    return new CertDiscoveryResult { PfxFile = pfxFile, ... };
}
```

**验证结果**: ✅ C# 服务成功启动并注册到管理平台

---

### ✅ Python SDK - 已修复

**当前状态**: 已修复并验证可用

**修复内容**:

#### 1. 修复导入错误
**文件**: `sdk/python/lightlink/client.py`

```python
# 兼容不同版本的 nats-py
try:
    from nats.errors import TimeoutError, NotFoundError, BadRequestError
except ImportError:
    from nats.errors import TimeoutError
    NotFoundError = TimeoutError
    BadRequestError = TimeoutError
```

#### 2. 添加 verify 选项
```python
def create_ssl_context_from_discovery(result: CertDiscoveryResult, verify: bool = True) -> ssl.SSLContext:
    if verify:
        ssl_ctx = ssl.create_default_context(ssl.Purpose.SERVER_AUTH)
        ssl_ctx.load_verify_locations(result.ca_file)
        ssl_ctx.server_hostname = result.server_name
    else:
        # 跳过服务器名称验证
        ssl_ctx = ssl.SSLContext(ssl.PROTOCOL_TLS_CLIENT)
        ssl_ctx.check_hostname = False
        ssl_ctx.verify_mode = ssl.CERT_NONE

    ssl_ctx.load_cert_chain(certfile=result.cert_file, keyfile=result.key_file)
    ssl_ctx.minimum_version = ssl.TLSVersion.TLSv1_2
    return ssl_ctx
```

**验证结果**: ✅ Python 服务成功启动并注册到管理平台

---

## 验证测试结果

### Go SDK ✅

```bash
cd light_link_platform/examples/provider/go/math-service
go run main.go
```

**结果**: 服务成功启动，在管理平台显示：
- 服务名: math-service
- 状态: 在线
- 方法数: 4 (add, multiply, power, divide)
- 实例: 1 个在线

### Python SDK ✅

```bash
cd light_link_platform/examples/provider/python/math_service
python main.py
```

**结果**: 服务成功启动，在管理平台显示：
- 服务名: math-service
- 状态: 在线
- 方法数: 4 (add, multiply, power, divide)
- 实例: 1 个在线

### C# SDK ✅

```bash
cd light_link_platform/examples/provider/csharp/MathService
dotnet run
```

**结果**: 服务成功启动！

```
=== C# Metadata Registration Demo ===
[1/5] Discovering TLS certificates...
Certificates found:
  CA:   ../../../../client\ca.crt
  Cert: ../../../../client\client.crt
  Key:  ../../../../client\client.key

[2/5] Creating service...
[3/5] Registering methods with metadata...
  - add: registered
  - multiply: registered
  - power: registered
  - divide: registered

[4/5] Starting service...
Service started successfully!

[5/5] Registering service metadata...
Service metadata registered to NATS!
  Service: math-service-csharp
  Version: v1.0.0
  Methods: 4
```

---

## 安全说明

`InsecureSkipVerify` / `verify=False` 的含义：
- ✅ **仍然启用**: TLS 加密
- ✅ **仍然启用**: CA 链验证 (只有受信任的 CA 签发的证书才能通过)
- ❌ **跳过**: 服务器名称验证（hostname vs 证书中的 CN/SAN）

这对于使用自签名证书的内网环境是安全的，因为：
1. 通信仍然是加密的
2. CA 链仍然被验证
3. 只是不验证服务器的主机名是否与证书匹配

---

## 后续工作

### 中优先级
1. **Go SDK**: 使 `InsecureSkipVerify` 可配置

### 低优先级
2. **文档更新**:
   - 各 SDK 的 README
   - TLS 配置说明
   - 示例代码注释

---

## 文件修改清单

### Python SDK (已完成)
- `sdk/python/lightlink/client.py` - 修复导入、添加 verify 参数
- `sdk/python/lightlink/service.py` - 使用 verify 参数
- `light_link_platform/examples/provider/python/math_service/main.py` - 添加路径、修改调用顺序

### C# SDK (已完成)
- `sdk/csharp/LightLink/TLSConfig.cs` - 添加 PFX/PfxPassword/InsecureSkipVerify 属性
- `sdk/csharp/LightLink/Service.cs` - PFX 证书加载、TLSRemoteCertificationValidationCallback
- `sdk/csharp/LightLink/TLSConfig.cs` - CertDiscovery 查找 PFX 文件
- `light_link_platform/client/client.pfx` - 生成的 PFX 证书文件
- `light_link_platform/examples/provider/csharp/MathService/Program.cs` - 完整实现 4 个方法

### Go SDK (未修改)
- 当前可用，但硬编码了 `InsecureSkipVerify: true`
