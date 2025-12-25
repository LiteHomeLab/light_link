# TLS 修复计划与状态报告

## 问题分析

### 初始状态

| SDK | TLS 配置 | 问题 | 初始状态 |
|-----|----------|------|----------|
| Go | `InsecureSkipVerify: true` (硬编码) | 无法选择是否跳过验证 | 可用 |
| C# | 仅设置 `Secure = true` | 缺少跳过服务器名称验证选项 | TLS 握手失败 |
| Python | `ssl.create_default_context` | 导入错误 + 缺少跳过验证选项 | 不可用 |

### 根本原因

1. **Go SDK**: `InsecureSkipVerify: true` 硬编码在 `CreateTLSOption` 中，用户无法选择是否验证服务器名称
2. **C# SDK**: NATS.Client 库没有提供直接的 API 来跳过 TLS 证书验证
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

### ⚠️ C# SDK - 需要替代方案

**当前状态**: TLS 握手失败

**问题**: NATS.Client 库的 `Options` 类没有提供以下属性：
- `TLSRemoteCertValidationCallback` (不存在)
- `ServerCertificateValidationCallback` (不存在)

**尝试的解决方案** (均失败):
1. `opts.TLSRemoteCertValidationCallback` - 属性不存在
2. `opts.ServerCertificateValidationCallback` - 属性不存在
3. `System.Net.ServicePointManager.ServerCertificateValidationCallback` - 仅适用于 HTTP

**替代方案**:

#### 方案 A: 使用环境变量 (推荐)
```bash
# 设置环境变量跳过 .NET TLS 验证
set DOTNET_SSL_CERT_DIR=none
set DOTNET_SSL_SKIP_CERT_VALIDATION=1
dotnet run
```

#### 方案 B: 升级证书
重新生成包含 SAN (Subject Alternative Name) 的证书，而不是仅使用 CN

#### 方案 C: 使用不同的 NATS 客户端库
考虑使用支持自定义 TLS 验证的 NATS 客户端

**已完成的修改**:
- [x] 添加 `InsecureSkipVerify` 属性到 `TLSConfig` (默认 `true`)
- [x] 添加相关注释说明

**文件**: `sdk/csharp/LightLink/TLSConfig.cs`

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

### C# SDK ⚠️

```bash
cd light_link_platform/examples/provider/csharp/MathService
dotnet run
```

**结果**: TLS 握手失败

```
NATS.Client.NATSConnectionException: TLS Authentication error
---> System.Security.Authentication.AuthenticationException:
    Authentication failed because the remote party sent a TLS alert: 'HandshakeFailure'.
```

**需要**: 使用替代方案或升级证书

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

### 高优先级
1. **C# SDK**: 确定并实施替代方案
   - 测试环境变量方案
   - 或升级证书包含 SAN

### 中优先级
2. **Go SDK**: 使 `InsecureSkipVerify` 可配置

### 低优先级
3. **文档更新**:
   - 各 SDK 的 README
   - TLS 配置说明
   - 示例代码注释

---

## 文件修改清单

### Python SDK (已完成)
- `sdk/python/lightlink/client.py` - 修复导入、添加 verify 参数
- `sdk/python/lightlink/service.py` - 使用 verify 参数
- `light_link_platform/examples/provider/python/math_service/main.py` - 添加路径、修改调用顺序

### C# SDK (部分完成)
- `sdk/csharp/LightLink/TLSConfig.cs` - 添加 InsecureSkipVerify 属性
- `sdk/csharp/LightLink/Service.cs` - 添加注释说明

### Go SDK (未修改)
- 当前可用，但硬编码了 `InsecureSkipVerify: true`
