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
    ├── go/            # Go 示例服务
    │   ├── metadata-demo/      # 元数据注册演示
    │   └── metadata-client/    # 元数据查询客户端演示
    │
    ├── csharp/        # C# 示例服务
    │   ├── MetadataDemo/       # 元数据注册演示
    │   ├── TextServiceDemo/    # 文本处理服务
    │   ├── RpcDemo/            # RPC 演示
    │   └── PubSubDemo.cs       # 发布订阅演示
    │
    └── python/        # Python 示例服务
        ├── metadata_demo.py    # 元数据注册演示
        ├── data_service.py     # 数据处理服务
        ├── rpc_service.py      # RPC 服务演示
        └── rpc_service_tls.py  # 支持 TLS 的 RPC 服务
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

**Go 元数据服务：**
```bash
cd examples/go/metadata-demo
go run main.go
```

**C# 文本服务：**
```bash
cd examples/csharp/TextServiceDemo
dotnet run
```

**Python 数据服务：**
```bash
cd examples/python
python data_service.py
```

## 服务概览

| 服务 | 语言 | 方法 |
|------|------|------|
| math-service | Go | add, multiply, power, divide |
| csharp-text-service | C# | reverse, uppercase, wordcount |
| python-data-service | Python | filter, transform, aggregate |

## 开发说明

- 所有服务会自动向管理平台注册元数据
- 服务每 30 秒发送一次心跳
- 平台显示服务状态、方法列表，并支持 RPC 调用
- 支持 TLS 认证（参考 `rpc_service_tls.py`）
