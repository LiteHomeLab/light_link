# LightLink 项目规则

## 项目概述

LightLink 是一个基于 NATS 的多语言后端服务通信框架，支持 C++、Python、C#、Go、JS 等语言编写的服务在内网多台服务器间通信。

## 核心功能

1. **RPC 远程调用** - 服务间函数调用
2. **消息发布/订阅** - 实时消息通知和广播
3. **状态保留** - 类似 MQTT retain 的最新状态功能（NATS KV）
4. **大文件传输** - 最大 1GB 文件传输（NATS Object Store）
5. **TLS 证书认证** - 双向 TLS 认证 + 用户权限配置

## 目录结构

```
light_link/
├── sdk/go/              # Go SDK (参考实现)
│   ├── client/          # 客户端 (RPC, 发布订阅, 状态管理, 文件传输)
│   ├── service/         # 服务端 (RPC 注册和处理)
│   └── types/           # 公共类型定义
├── sdk/python/          # Python SDK
├── deploy/nats/         # NATS 服务器配置和 TLS 证书
├── examples/            # 示例项目
└── docs/                # 文档
```

## 开发规则

1. 使用中文回答问题
2. 当前开发系统是 Windows
3. 开发、编辑 BAT 脚本里面不要有中文字符
4. 修复脚本时在原有脚本上修复，非必需不要新建脚本
5. 遵循 TDD 开发模式：先写测试，再实现功能
6. 每个功能完成后提交一次代码
7. 所有测试必须通过后才能提交

## NATS 服务配置

- 配置文件：`deploy/nats/nats-server.conf`
- TLS 证书目录：`deploy/nats/tls/`
- 默认端口：4222
- 需要 JetStream 支持（KV 和 Object Store）

## 快速开始

### 启动 NATS 服务器

```bash
nats-server -config deploy/nats/nats-server.conf
```

### 运行测试

```bash
go test ./...
```

### 运行示例

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
