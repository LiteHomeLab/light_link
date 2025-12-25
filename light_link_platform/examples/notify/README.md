# Notify - 消息通知与状态管理

本目录包含使用消息发布订阅和状态管理的示例程序。

## 适合放在这里的项目类型

### 1. 发布/订阅 (Pub/Sub)
- **发布者**: 发布消息到 NATS 主题
- **订阅者**: 订阅 NATS 主题并接收消息
- **适用场景**: 实时通知、事件广播、日志收集

### 2. 状态管理 (KV)
- **发布状态**: 使用 NATS KV 存储最新状态
- **获取状态**: 从 NATS KV 读取最新状态
- **监听状态变化**: 监听 KV 键的变化
- **适用场景**: 配置中心、状态同步、缓存更新

### 3. 消息队列 (JetStream)
- **生产者**: 发送消息到 JetStream 流
- **消费者**: 从 JetStream 流消费消息
- **适用场景**: 可靠消息传递、事件溯源、任务队列

## 示例项目

| 语言 | 项目 | 类型 | 说明 |
|------|------|------|------|
| C# | PubSubDemo | 发布订阅 | 发布和接收消息演示 |

## 功能对比

| 功能 | Pub/Sub | KV (状态) | JetStream |
|------|---------|-----------|-----------|
| 实时性 | 实时 | 按需 | 实时/持久 |
| 持久化 | 否 | 是 | 是 |
| 消息历史 | 否 | 最后值 | 完整历史 |
| 离线接收 | 错过 | 可获取 | 可重放 |
| 典型场景 | 通知 | 状态同步 | 可靠消息 |

## 创建新的 Notify 项目

### 发布消息
```go
// Go
nc.Publish("subject", data)

// C#
conn.Publish("subject", encodedData);

// Python
nc.publish("subject", data)
```

### 订阅消息
```go
// Go
nc.Subscribe("subject", handler)

// C#
var sub = conn.SubscribeSync("subject");

// Python
nc.subscribe("subject", cb)
```

### 状态管理 (KV)
```go
// Go
kv, _ := js.CreateKeyValue(&nats.KeyValueConfig{Bucket: "config"})
kv.Put("key", value)
entry, _ := kv.Get("key")

// Python
kv = await js.create_key_value(bucket="config")
await kv.put("key", value)
entry = await kv.get("key")
```

## 与 Provider/Caller 的区别

| 目录 | 角色 | 主要操作 |
|------|------|----------|
| **provider** | 服务提供者 | 注册 RPC 方法，响应调用 |
| **caller** | 服务调用者 | 调用 RPC 方法 |
| **notify** | 消息处理 | 发布/订阅消息，状态管理 |
