# RPC Services 测试报告

## 测试结果摘要

| 服务实现 | 服务名称 | Subject | 测试结果 | 问题 |
|---------|---------|---------|---------|------|
| Go | `math-service` | `$SRV.math-service.add` | ✅ **通过** | 无 |
| C# | `math-service-csharp` | `$SRV.math-service-csharp.add` | ✅ **通过** | JSON 序列化问题已修复 |
| Python | `math-service` | `$SRV.math-service.add` | ✅ **通过** | async/await 问题已修复 |

## 发现的问题及修复

### 问题 1: C# SDK - JSON 序列化问题

**现象**: 调用 C# 服务时返回 `Method not found:` (方法名为空)

**原因**: `RPCRequest` 类型缺少 `JsonPropertyName` 属性，导致 camelCase JSON 无法正确反序列化到 PascalCase 属性

**修复**: 在 `sdk/csharp/LightLink/Types.cs` 中添加 JSON 属性映射：

```csharp
public class RPCRequest
{
    [JsonPropertyName("id")]
    public string Id { get; set; } = "";

    [JsonPropertyName("method")]
    public string Method { get; set; } = "";

    [JsonPropertyName("args")]
    public Dictionary<string, object> Args { get; set; } = new();
}
```

**文件**: `sdk/csharp/LightLink/Types.cs`

### 问题 2: C# 服务 - 类型转换问题

**现象**: `Unable to cast object of type 'System.Text.Json.JsonElement' to type 'System.IConvertible'`

**原因**: 反序列化后的参数值是 `JsonElement` 类型，不能直接使用 `Convert.ToDouble()`

**修复**: 在 `light_link_platform/examples/provider/csharp/MathService/Program.cs` 中添加辅助方法：

```csharp
static double GetDouble(Dictionary<string, object> args, string key)
{
    var value = args[key];
    if (value is System.Text.Json.JsonElement element)
    {
        if (element.ValueKind == System.Text.Json.JsonValueKind.Number)
            return element.GetDouble();
        // ...
    }
    return Convert.ToDouble(value);
}
```

**文件**: `light_link_platform/examples/provider/csharp/MathService/Program.cs`

### 问题 3: Python SDK - async/await 问题

**现象**: 警告 `coroutine 'Msg.respond' was never awaited`，服务无法响应请求

**原因**: `msg.respond()` 返回 coroutine 但没有 await

**修复**: 在 `sdk/python/lightlink/service.py` 中添加 await：

```python
async def _send_success(self, msg, request_id: str, result: Dict[str, Any]) -> None:
    response = RPCResponse(id=request_id, success=True, result=result)
    await msg.respond(json.dumps(response.__dict__).encode())
```

**文件**: `sdk/python/lightlink/service.py`

### 问题 4: 管理平台调用服务名称不匹配

**现象**: 管理平台调用 `math-service` 时，C# 服务响应 `no responders`

**原因**: C# 服务注册为 `math-service-csharp`，但管理平台调用 `math-service`

**状态**: 这是设计问题，不是 bug。需要在管理平台添加服务发现和路由功能

## 测试命令

### 测试 Go 服务
```bash
# 终端 1: 启动服务
cd light_link_platform/examples/provider/go/math-service
go run main.go

# 终端 2: 测试
cd light_link
go run test_rpc_client.go math-service add
```

### 测试 C# 服务
```bash
# 终端 1: 启动服务
cd light_link_platform/examples/provider/csharp/MathService
dotnet run

# 终端 2: 测试 (使用正确的服务名)
cd light_link
go run test_rpc_client.go math-service-csharp add
```

### 测试 Python 服务
```bash
# 终端 1: 启动服务
cd light_link_platform/examples/provider/python/math_service
python main.py

# 终端 2: 测试
cd light_link
go run test_rpc_client.go math-service add
```

## 管理平台集成问题

### 当前状态
- ✅ 服务可以正常启动和注册
- ✅ 心跳正常发送
- ✅ 服务元数据可以获取
- ❌ **RPC 调用失败** (因为服务名称不匹配)

### 解决方案

#### 选项 1: 统一服务名称 (推荐用于演示)
将所有服务的名称改为 `math-service`：

**C# 修改**:
```csharp
// light_link_platform/examples/provider/csharp/MathService/Program.cs:24
var svc = new Service("math-service", "nats://172.18.200.47:4222", tlsConfig);
```

**优点**: 简单直接，管理平台无需修改
**缺点**: 同一时间只能运行一个语言版本的服务

#### 选项 2: 管理平台添加服务发现
在管理平台中实现智能路由：

1. 监听所有服务的心跳 (`$LL.heartbeat.>`)
2. 维护服务列表，包括服务名称、语言、版本等信息
3. 当调用 `math-service` 时，如果有多个版本：
   - 提示用户选择
   - 或按优先级选择 (Go > C# > Python)
   - 或实现负载均衡

#### 选项 3: 使用命名空间区分
服务名称改为带语言前缀：
- `go.math-service`
- `csharp.math-service`
- `python.math-service`

管理平台添加前缀选择器。

## 已创建的文件

1. **test_rpc_client.go** - Go 测试客户端，可直接测试 RPC 服务
2. **test_rpc_services.bat** - 交互式测试脚本
3. **RPC_TEST_PLAN.md** - 本测试报告

## 下一步工作

1. **修复 C# 服务名称** - 改为 `math-service` 以匹配管理平台调用
2. **在管理平台添加服务发现功能** - 支持多语言版本同时运行
3. **添加集成测试** - 验证管理平台到服务的完整调用链路
