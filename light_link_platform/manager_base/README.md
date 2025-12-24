# LightLink 服务管理平台部署指南

## 概述

LightLink 服务管理平台是一个完整的 Web 控制台，提供服务注册、发现、监控和 RPC 调试功能。

## 架构

```
+----------------+     +----------------+     +----------------+
|                |     |                |     |                |
|   Vue 3 前端   |<--->|   Go 后端      |<--->|    NATS       |
|                |     |                |     |                |
+----------------+     +----------------+     +----------------+
                              |
                              v
                       +----------------+
                       |                |
                       |  SQLite 数据库 |
                       |                |
                       +----------------+
```

## 快速启动

### 1. 前置要求

- Go 1.21+
- Node.js 18+
- NATS Server 2.10+ (支持 JetStream)

### 2. 启动 NATS 服务器

```bash
nats-server -config deploy/nats/nats-server.conf
```

### 3. 启动后端服务

```bash
cd console/server
go run main.go
```

后端服务将在 `http://localhost:8080` 启动。

### 4. 启动前端（开发模式）

```bash
cd console/web
npm install
npm run dev
```

前端开发服务器将在 `http://localhost:5173` 启动。

### 5. 访问控制台

浏览器打开 `http://localhost:5173`，使用默认账号登录：
- 用户名: `admin`
- 密码: `admin123`

## 生产部署

### 构建前端

```bash
cd console/web
npm run build
```

构建产物将输出到 `console/web/dist/`。

### 配置后端

编辑 `console/server/console.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8080

nats:
  url: "nats://localhost:4222"
  tls:
    enabled: false
    cert: ""
    key: ""
    ca: ""

database:
  path: "./data/console.db"

jwt:
  secret: "your-secret-key-change-in-production"
  expiry: 24h

heartbeat:
  interval: 30s
  timeout: 90s

admin:
  username: "admin"
  password: "admin123"
```

### 启动后端服务（生产模式）

```bash
cd console/server
go build -o lightlink-console.exe
./lightlink-console.exe
```

### 使用反向代理（可选）

使用 Nginx 作为反向代理：

```nginx
server {
    listen 80;
    server_name console.lightlink.local;

    # 前端静态文件
    location / {
        root /path/to/light_link/console/web/dist;
        try_files $uri $uri/ /index.html;
    }

    # API 代理
    location /api/ {
        proxy_pass http://localhost:8080/api/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

    # WebSocket 代理
    location /api/ws {
        proxy_pass http://localhost:8080/api/ws;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

## REST API 文档

### 认证

所有 API 请求需要在 Header 中携带 JWT Token：

```
Authorization: Bearer <token>
```

### API 端点

#### 1. 登录

```
POST /api/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}
```

响应：
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "role": "admin"
}
```

#### 2. 获取服务列表

```
GET /api/services
```

响应：
```json
[
  {
    "name": "math-service",
    "version": "v1.0.0",
    "description": "Math operations service",
    "author": "LiteHomeLab",
    "tags": ["demo", "math"],
    "registered_at": "2024-12-24T10:00:00Z",
    "updated_at": "2024-12-24T10:00:00Z"
  }
]
```

#### 3. 获取服务详情

```
GET /api/services/{service_name}
```

#### 4. 获取服务方法列表

```
GET /api/services/{service_name}/methods
```

响应：
```json
[
  {
    "name": "add",
    "description": "Add two numbers",
    "parameters": [
      {"name": "a", "type": "number", "required": true, "description": "First number"},
      {"name": "b", "type": "number", "required": true, "description": "Second number"}
    ],
    "return_info": {
      "type": "object",
      "description": "Result object"
    },
    "examples": [
      {
        "name": "Basic addition",
        "input": {"a": 10, "b": 20},
        "output": {"sum": 30},
        "description": "10 + 20 = 30"
      }
    ]
  }
]
```

#### 5. 获取服务状态列表

```
GET /api/status
```

响应：
```json
[
  {
    "service_name": "math-service",
    "online": true,
    "last_seen": "2024-12-24T10:30:00Z",
    "version": "v1.0.0"
  }
]
```

#### 6. 获取事件列表

```
GET /api/events?limit=100&offset=0
```

响应：
```json
[
  {
    "id": 1,
    "type": "registered",
    "service_name": "math-service",
    "message": "Service registered",
    "created_at": "2024-12-24T10:00:00Z"
  }
]
```

#### 7. RPC 调用

```
POST /api/call
Content-Type: application/json

{
  "service": "math-service",
  "method": "add",
  "params": {
    "a": 10,
    "b": 20
  }
}
```

响应：
```json
{
  "success": true,
  "data": {"sum": 30},
  "duration": 15
}
```

## WebSocket API

### 连接

```
ws://localhost:8080/api/ws
```

连接时需要在 URL 中携带 token：
```
ws://localhost:8080/api/ws?token=<jwt_token>
```

### 订阅频道

发送订阅消息：
```json
{
  "action": "subscribe",
  "channels": ["events", "status"]
}
```

### 消息格式

```json
{
  "channel": "events",
  "event": {
    "type": "online",
    "service_name": "math-service",
    "message": "Service is online",
    "created_at": "2024-12-24T10:00:00Z"
  }
}
```

## 服务元数据注册

服务启动时自动向 NATS 发送注册消息：

### 注册消息格式

主题: `$LL.register.{service_name}`

```json
{
  "service": "math-service",
  "version": "v1.0.0",
  "metadata": {
    "name": "math-service",
    "version": "v1.0.0",
    "description": "Math operations service",
    "author": "LiteHomeLab",
    "tags": ["demo", "math"],
    "methods": [...]
  },
  "timestamp": "2024-12-24T10:00:00Z"
}
```

### 心跳消息格式

主题: `$LL.heartbeat.{service_name}`

```json
{
  "service": "math-service",
  "version": "v1.0.0",
  "timestamp": "2024-12-24T10:00:00Z"
}
```

心跳间隔: 30 秒
超时判定: 90 秒

## 故障排查

### 1. 服务无法注册

- 检查 NATS 服务器是否运行
- 检查服务是否正确发送注册消息
- 使用 `nats-sub "$LL.register.>`` 监听注册消息

### 2. 前端无法连接后端

- 检查后端是否运行在 8080 端口
- 检查 CORS 配置
- 查看浏览器控制台错误信息

### 3. WebSocket 连接失败

- 检查 JWT token 是否有效
- 检查 WebSocket 代理配置
- 使用 WebSocket 测试工具验证连接

### 4. 服务显示离线

- 检查心跳消息是否正常发送
- 检查心跳超时配置
- 使用 `nats-sub "$LL.heartbeat.>`` 监听心跳消息
