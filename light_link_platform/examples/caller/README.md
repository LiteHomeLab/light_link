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

### call-math-service-go

调用 math-service-go 提供的数学计算服务的示例程序。

**功能特性：**
- 依赖检查：启动前检查所需的服务和方法是否已注册
- 自动等待：如果依赖未满足，自动等待服务注册（无限等待）
- 实时状态：显示当前依赖满足情况（✓ 已满足 / ✗ 未满足）
- 自动调用：依赖满足后自动执行 5 次计算
  - add(10, 20) = 30
  - multiply(5, 6) = 30
  - power(2, 10) = 1024
  - divide(100, 4) = 25
  - 复杂计算：power(add(10, 5), 2) = 225

**运行方式：**

```bash
# 1. 确保已启动管理平台
cd light_link_platform/manager_base/server
go run main.go

# 2. 确保已启动 math-service-go（新终端窗口）
cd light_link_platform/examples/provider/go/math-service
go run main.go

# 3. 运行 caller 示例（新终端窗口）
cd light_link_platform/examples/caller/go/call-math-service-go
run.bat
```

**测试场景：**

1. **正常启动**：先启动 provider，再启动 caller → 立即开始调用
2. **依赖等待**：先启动 caller，再启动 provider → caller 等待直到 provider 注册
3. **方法缺失**：provider 只注册部分方法 → caller 等待直到所有方法可用

**输出示例：**

```
[call-math-service-go] 正在连接到 NATS: nats://172.18.200.47:4222...
[call-math-service-go] 使用 TLS 证书连接
[call-math-service-go] 连接成功
[call-math-service-go] 正在检查依赖服务...

依赖检查状态:
math-service-go:
  - add [✓]
  - multiply [✓]
  - power [✓]
  - divide [✓]

[call-math-service-go] 所有依赖已满足！

[call-math-service-go] 调用 add(10, 20) = 30
[call-math-service-go] 调用 multiply(5, 6) = 30
[call-math-service-go] 调用 power(2, 10) = 1024
[call-math-service-go] 调用 divide(100, 4) = 25
[call-math-service-go] 复杂计算: add(10, 5) = 15, power(15, 2) = 225
```

**代码结构：**

- `main.go` - 主程序入口
  - 配置依赖检查（需要哪些服务和方法）
  - 等待依赖满足
  - 执行 RPC 调用
- `run.bat` - Windows 启动脚本
