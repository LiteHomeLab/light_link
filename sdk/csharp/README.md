# LightLink C# SDK

C# 客户端和服务端 SDK 用于 LightLink 多语言 RPC 框架。

## 安装

```bash
dotnet add package LightLink
```

## 快速开始

### 服务端

```csharp
using LightLink;

var svc = new Service("my-service", "nats://localhost:4222");

svc.RegisterRPC("echo", async (args) => {
    return args;
});

svc.Start();
```

### 元数据注册

```csharp
var meta = new MethodMetadata {
    Name = "add",
    Params = new List<ParameterMetadata> {
        new() { Name = "a", Type = "number", Required = true }
    }
};

svc.RegisterMethodWithMetadata("add", handler, meta);
```

## 协议兼容性

- JSON 格式与 Go SDK 完全一致
- 支持 NATS JetStream
- 支持参数验证

## 示例

- `Examples/RpcDemo` - 基本 RPC 服务示例
- `Examples/MetadataDemo` - 元数据注册示例
