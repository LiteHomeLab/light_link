# Provider - 服务提供者

本目录包含可以被 `manager_base` 管理平台调用的服务。

## 适合放在这里的项目类型

### RPC 服务提供者
- 注册 RPC 方法供其他服务调用
- 使用 `service.NewService()` 创建服务
- 使用 `RegisterMethod()` 或 `RegisterMethodWithMetadata()` 注册方法
- 发送心跳维持服务状态

### 示例项目

| 语言 | 项目 | 服务名 | 方法 |
|------|------|--------|------|
| Go | math-service | math-service | add, multiply, power, divide |
| C# | MathService | math-service | add |
| C# | TextService | csharp-text-service | reverse, uppercase, wordcount |
| Python | math_service | math-service | add, multiply, power, divide |
| Python | data_service | python-data-service | filter, transform, aggregate |

## 创建新的 Provider 服务

### Go 示例
```go
svc, err := service.NewService("service-name", natsURL, nil)
svc.RegisterMethodWithMetadata("method", handler, metadata)
svc.Start()
```

### C# 示例
```csharp
var svc = new Service("service-name", "nats://localhost:4222");
svc.RegisterMethodWithMetadata("method", Handler, metadata);
svc.Start();
```

### Python 示例
```python
svc = Service("service-name", "nats://localhost:4222")
await svc.register_method_with_metadata("method", handler, metadata)
await svc.start()
```

## 服务特性

- **自动注册**: 服务启动后自动向管理平台注册元数据
- **心跳维持**: 每 30 秒发送一次心跳
- **元数据支持**: 提供完整的方法描述、参数说明、使用示例
- **TLS 支持**: 可选的 TLS 加密连接
