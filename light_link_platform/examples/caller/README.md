# Caller - 服务调用者

本目录包含调用 Provider 服务的客户端示例程序。

## 适合放在这里的项目类型

### RPC 客户端
- 调用 Provider 服务的 RPC 方法
- 使用 `client.NewClient()` 创建客户端
- 使用 `Call()` 方法调用远程服务

### 示例场景

- 调用数学服务进行计算
- 调用文本服务处理字符串
- 调用数据服务进行分析
- 批量调用多个服务

## 创建新的 Caller 客户端

### Go 示例
```go
cli, err := client.NewClient(natsURL)
result, err := cli.Call("service-name", "method-name", params)
```

### C# 示例
```csharp
var client = new Client("nats://localhost:4222");
var result = await client.CallAsync("service-name", "method-name", parameters);
```

### Python 示例
```python
cli = Client("nats://localhost:4222")
result = await cli.call("service-name", "method-name", params)
```

## 注意事项

- Caller 不注册任何 RPC 方法
- Caller 不发送心跳
- Caller 可以调用多个 Provider 服务
- Caller 应该处理服务不可用的情况

## 当前状态

> 待添加示例项目...
