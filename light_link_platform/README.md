# LightLink 平台示例

本目录包含 LightLink 框架的管理平台和多语言示例服务。

## 目录结构

```
light_link_platform/
├── manager_base/       # 管理平台（后端 + 前端）
│   ├── server/        # Go 后端服务器
│   ├── web/           # Vue 3 前端
│   ├── data/          # 数据库和存储
│   └── README.md      # 平台文档
│
└── examples/          # 多语言示例服务
    ├── provider/      # 服务提供者（被 manager_base 调用）
    │   ├── go/
    │   │   └── math-service/     # Go 数学服务
    │   ├── csharp/
    │   │   ├── MathService/      # C# 数学服务
    │   │   └── TextService/      # C# 文本服务
    │   └── python/
    │       ├── math_service/     # Python 数学服务
    │       └── data_service/     # Python 数据服务
    │
    └── caller/        # 服务使用者
        └── csharp/
            └── PubSubDemo/       # C# 发布订阅示例
```

## 快速开始

### 1. 启动 NATS 服务器

```bash
nats-server -config ../deploy/nats/nats-server.conf
```

### 2. 启动管理平台

```bash
cd manager_base/server
go run main.go
```

然后在浏览器中打开 `http://localhost:8080`

### 3. 运行示例服务

**服务提供者 (Provider):**

```bash
# Go 数学服务
cd examples/provider/go/math-service
go run main.go

# C# 数学服务
cd examples/provider/csharp/MathService
dotnet run

# C# 文本服务
cd examples/provider/csharp/TextService
dotnet run

# Python 数学服务
cd examples/provider/python/math_service
python main.py

# Python 数据服务
cd examples/provider/python/data_service
python main.py
```

**服务使用者 (Caller):**

```bash
# C# 发布订阅示例
cd examples/caller/csharp/PubSubDemo
dotnet run
```

## 服务概览

### Provider 服务（被管理平台调用）

| 服务 | 语言 | 方法 |
|------|------|------|
| math-service | Go | add, multiply, power, divide |
| math-service | C# | add |
| csharp-text-service | C# | reverse, uppercase, wordcount |
| math-service | Python | add, multiply, power, divide |
| python-data-service | Python | filter, transform, aggregate |

### Caller 示例

| 项目 | 语言 | 说明 |
|------|------|------|
| PubSubDemo | C# | 消息发布订阅演示 |

## 开发说明

- 所有服务会自动向管理平台注册元数据
- 服务每 30 秒发送一次心跳
- 平台显示服务状态、方法列表，并支持 RPC 调用
- 支持 TLS 认证（参考 `rpc_service_tls.py`）
