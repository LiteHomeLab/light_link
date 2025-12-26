# Caller 依赖检查功能文档

## 概述

依赖检查功能允许客户端程序在执行 RPC 调用之前，先验证所需的服务和方法是否已注册到系统中。这避免了在服务不可用时进行无效调用，提高了系统的健壮性。

## 核心组件

### DependencyChecker

依赖检查器位于 `sdk/go/client/dependency.go`，提供以下功能：

1. **服务发现**：从 NATS KV 存储中查询已注册的服务
2. **实时监听**：订阅 `$LL.register.>` 主题，接收新服务注册通知
3. **依赖验证**：检查所需的服务和方法是否全部可用
4. **状态显示**：实时显示依赖满足进度

## API 使用

### 创建依赖检查器

```go
import "github.com/LiteHomeLab/light_link/sdk/go/client"

deps := []client.Dependency{
    {
        ServiceName: "math-service-go",
        Methods:     []string{"add", "multiply", "power", "divide"},
    },
}

checker := client.NewDependencyChecker(cli.GetNATSConn(), deps)
defer checker.Close()
```

### 等待依赖满足

```go
import "context"

// 无限等待，直到所有依赖满足
err := checker.WaitForDependencies(context.Background())
if err != nil {
    log.Fatal(err)
}

fmt.Println("所有依赖已满足，可以开始调用服务")
```

### 依赖结构

```go
type Dependency struct {
    ServiceName string   // 服务名称
    Methods     []string // 需要的方法列表
}
```

## 工作流程

```
┌─────────────┐
│   启动程序   │
└──────┬──────┘
       │
       ▼
┌─────────────────────┐
│ 连接到 NATS         │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│ 查询 KV 存储中的    │
│ 已注册服务          │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│ 订阅 $LL.register.> │
│ 接收新服务注册      │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│ 检查依赖是否满足    │
└──────┬──────────────┘
       │
   ┌───┴───┐
   │       │
   ▼       ▼
 满足    未满足
   │       │
   │       ▼
   │   ┌──────────┐
   │   │显示进度  │
   │   │等待1秒   │
   │   └────┬─────┘
   │        │
   │        └────►┐
   │              │
   ◄──────────────┘
   │
   ▼
┌─────────────────────┐
│ 开始调用服务        │
└─────────────────────┘
```

## 实现细节

### KV 查询

依赖检查器启动时会从 NATS KV 存储中查询已注册的服务：

```go
func (dc *DependencyChecker) queryExistingServices() error {
    js, err := jetstream.New(dc.nc)
    kv, err := js.KeyValue(ctx, "light_link_state")

    keys, err := kv.Keys(ctx)
    for _, key := range keys {
        if strings.HasPrefix(key, "service.") {
            entry, err := kv.Get(ctx, key)
            // 解析服务元数据...
        }
    }
}
```

### 注册消息监听

同时订阅服务注册消息，接收实时更新：

```go
sub, err := dc.nc.Subscribe("$LL.register.>", func(msg *nats.Msg) {
    dc.handleRegisterMessage(msg)
})
```

### 线程安全

依赖检查器使用 `sync.RWMutex` 保护并发访问：

- `registered` map 的读写操作
- `printAllSatisfied()` 输出
- `Close()` 方法

## 示例程序

### call-math-service-go

完整示例位于 `light_link_platform/examples/caller/go/call-math-service-go/`

**功能：**
1. 检查 `math-service-go` 的 4 个方法是否可用
2. 等待所有依赖满足
3. 执行 5 次计算调用

**运行方式：**

```bash
cd light_link_platform/examples/caller/go/call-math-service-go
run.bat
```

## 测试场景

### 场景 1：正常启动

1. 启动 math-service-go
2. 启动 call-math-service-go
3. **结果**：立即检测到依赖满足，开始调用

### 场景 2：依赖等待

1. 启动 call-math-service-go
2. 等待依赖检查输出
3. 启动 math-service-go
4. **结果**：caller 检测到服务注册，开始调用

### 场景 3：部分方法缺失

1. 修改 math-service-go，只注册部分方法
2. 启动 call-math-service-go
3. **结果**：等待所有方法注册完成

## 最佳实践

1. **设置合理的超时**：虽然支持无限等待，但建议设置超时避免永久阻塞
2. **清晰的依赖声明**：只声明真正需要的方法，减少等待时间
3. **错误处理**：检查返回错误，提供友好的错误信息
4. **资源清理**：使用 defer 确保检查器正确关闭

## 与 OpenAPI 文档的集成

依赖检查功能可以与 OpenAPI 文档配合使用：

1. 服务提供者发布 OpenAPI 文档描述其 API
2. 调用者读取 OpenAPI 文档了解可用方法
3. 使用依赖检查确保所需方法可用
4. 执行 RPC 调用

## 相关文件

- `sdk/go/client/dependency.go` - 依赖检查器实现
- `sdk/go/client/dependency_test.go` - 单元测试
- `light_link_platform/examples/caller/go/call-math-service-go/` - 示例程序
- `light_link_platform/manager_base/server/manager/registry.go` - 注册管理（KV 存储）
