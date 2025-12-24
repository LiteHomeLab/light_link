# LightLink 管理平台启动指南

本文档说明如何启动 LightLink 管理平台及其依赖的多语言示例服务。

---

## 一、NATS 服务器启动

### 1.1 TLS 证书要求

NATS 服务器需要以下 TLS 证书文件，位于 `deploy/nats/tls/` 目录：

| 证书文件 | 说明 |
|---------|------|
| `ca.crt` | CA 根证书 |
| `server.crt` | 服务器证书 |
| `server.key` | 服务器私钥 |

### 1.2 NATS 配置文件

配置文件路径：`deploy/nats/nats-server.conf`

关键配置：
- 监听端口：`4222`
- JetStream 存储：`./data`
- 日志文件：`./logs/nats-server.log`
- TLS 最低版本：`1.2`

### 1.3 启动命令

```bash
nats-server -config deploy/nats/nats-server.conf
```

---

## 二、管理平台启动 (manager_base)

### 2.1 配置文件

配置文件路径：`light_link_platform/manager_base/server/console.yaml`

```yaml
server:
  host: "0.0.0.0"
  port: 8080

nats:
  url: "nats://localhost:4222"
  tls:
    enabled: false

database:
  path: "data/light_link.db"

jwt:
  secret: "light-link-secret-key-change-in-production"
  expiry: 24h

heartbeat:
  timeout: 90s

admin:
  username: "admin"
  password: "admin123"
```

### 2.2 启动命令

```bash
cd light_link_platform/manager_base/server
go run main.go
```

### 2.3 访问管理平台

- URL: `http://localhost:8080`
- 管理员账号：
  - 用户名: `admin`
  - 密码: `admin123`

---

## 三、多语言示例服务启动

### 3.1 Go 示例服务

| 服务目录 | 服务名 | 说明 | 启动命令 |
|---------|-------|------|---------|
| `examples/go/metadata-demo` | `math-service` | 数学运算服务（加减乘除、幂运算） | `go run main.go` |
| `examples/go/metadata-client` | - | 数学服务客户端（调用测试） | `go run main.go` |

**需要启动的服务：** `metadata-demo`

```bash
cd light_link_platform/examples/go/metadata-demo
go run main.go
```

### 3.2 C# 示例服务

| 服务目录 | 服务名 | 说明 | 启动命令 |
|---------|-------|------|---------|
| `examples/csharp/TextServiceDemo` | `csharp-text-service` | 文本处理服务（反转、大写、词统计） | `dotnet run` |
| `examples/csharp/MetadataDemo` | - | 元数据示例 | `dotnet run` |
| `examples/csharp/RpcDemo` | - | RPC 调用示例 | `dotnet run` |

**需要启动的服务：** `TextServiceDemo`

```bash
cd light_link_platform/examples/csharp/TextServiceDemo
dotnet run
```

### 3.3 Python 示例服务

| 服务文件 | 服务名 | 说明 | 启动命令 |
|---------|-------|------|---------|
| `examples/python/data_service.py` | `python-data-service` | 数据处理服务（过滤、转换、聚合） | `python data_service.py` |
| `examples/python/metadata_demo.py` | `math-service` | 元数据注册示例 | `python metadata_demo.py` |
| `examples/python/rpc_service.py` | - | RPC 服务示例 | `python rpc_service.py` |
| `examples/python/rpc_service_tls.py` | - | TLS RPC 服务示例 | `python rpc_service_tls.py` |

**需要启动的服务：** `data_service.py`

```bash
cd light_link_platform/examples/python
python data_service.py
```

---

## 四、完整启动顺序

### 方式一：手动启动

按以下顺序依次启动：

1. **启动 NATS 服务器**
   ```bash
   nats-server -config deploy/nats/nats-server.conf
   ```

2. **启动管理平台**
   ```bash
   cd light_link_platform/manager_base/server
   go run main.go
   ```

3. **启动 Go 示例服务**
   ```bash
   cd light_link_platform/examples/go/metadata-demo
   go run main.go
   ```

4. **启动 C# 示例服务**
   ```bash
   cd light_link_platform/examples/csharp/TextServiceDemo
   dotnet run
   ```

5. **启动 Python 示例服务**
   ```bash
   cd light_link_platform/examples/python
   python data_service.py
   ```

### 方式二：多窗口启动

在不同终端窗口中并行启动各服务（需先启动 NATS）。

---

## 五、服务注册验证

启动所有服务后，访问管理平台 `http://localhost:8080`，在服务列表页面应能看到以下已注册的服务：

| 服务名 | 语言 | 说明 |
|-------|------|------|
| `math-service` | Go | 数学运算服务 |
| `csharp-text-service` | C# | 文本处理服务 |
| `python-data-service` | Python | 数据处理服务 |

---

## 六、目录结构参考

```
light_link/
├── deploy/nats/              # NATS 配置和证书
│   ├── nats-server.conf      # NATS 配置文件
│   └── tls/                  # TLS 证书目录
│       ├── ca.crt
│       ├── server.crt
│       └── server.key
├── light_link_platform/
│   ├── manager_base/         # 管理平台
│   │   └── server/
│   │       ├── main.go
│   │       └── console.yaml  # 管理平台配置
│   └── examples/             # 多语言示例服务
│       ├── go/
│       │   └── metadata-demo/
│       ├── csharp/
│       │   └── TextServiceDemo/
│       └── python/
│           └── data_service.py
```
