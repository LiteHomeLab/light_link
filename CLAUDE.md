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
├── sdk/csharp/          # C# SDK
├── deploy/nats/         # NATS 服务器配置
│   ├── tls/             # TLS 证书模板和文档
│   └── create_tls/      # TLS 证书生成脚本 (Git submodule)
├── examples/            # SDK 基础功能示例
├── light_link_platform/ # 管理平台和多语言示例服务
│   ├── manager_base/    # 管理平台 (server + web)
│   └── examples/        # 多语言示例服务
│       ├── provider/    # 服务提供者（被 manager_base 调用）
│       │   ├── go/
│       │   ├── csharp/
│       │   └── python/
│       └── caller/      # 服务使用者
│           └── csharp/
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
8. **平台示例管理**: 所有管理平台和多语言示例服务统一放在 `light_link_platform/` 目录下
   - `manager_base/` - 管理平台（后端 + 前端在一个文件夹）
   - `examples/provider/` - 服务提供者（被 manager_base 调用的服务）
     - `go/` - Go 示例服务
     - `csharp/` - C# 示例服务
     - `python/` - Python 示例服务
   - `examples/caller/` - 服务使用者（调用其他服务或发布订阅消息）
     - `csharp/` - C# 调用者示例
   - 新增示例服务时，根据角色放入 provider 或 caller 目录

## NATS 服务配置

- 远程服务器地址：`172.18.200.47:4222` (已部署，无需本地启动)
- 配置文件：`deploy/nats/nats-server.conf`
- TLS 证书：通过 `deploy/nats/create_tls/` 子模块生成（Git submodule -> git@github.com:LiteHomeLab/create_tls.git）
- 默认端口：4222
- 需要 JetStream 支持（KV 和 Object Store）

### 生成 TLS 证书

证书生成脚本已移至独立仓库 `create_tls`，作为 Git 子模块引入：

```bash
# 生成证书
cd deploy/nats/create_tls
setup-certs.bat

# 更新子模块到最新版本
cd deploy/nats/create_tls
git pull origin main
```

## 快速开始

### 连接 NATS 服务器

NATS 服务器已部署在远程地址 `172.18.200.47:4222`，本地调试无需启动。
如需使用环境变量覆盖：`set NATS_URL=nats://custom-address:4222`

### 运行测试

```bash
go test ./...
```

### 运行示例

**SDK 基础功能示例:**
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

**管理平台和示例服务 (light_link_platform):**
```bash
# 启动管理平台
cd light_link_platform/manager_base/server
go run main.go
# 访问 http://localhost:8080

# 启动 Provider 服务（被管理平台调用）
cd light_link_platform/examples/provider/go/math-service
go run main.go

cd light_link_platform/examples/provider/csharp/MathService
dotnet run

cd light_link_platform/examples/provider/python/math_service
python main.py

# 启动 Caller 示例（调用其他服务）
cd light_link_platform/examples/caller/csharp/PubSubDemo
dotnet run
```
