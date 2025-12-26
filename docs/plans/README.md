# LightLink 功能开发实施计划索引

本目录包含 LightLink 框架功能开发的详细实施计划，按优先级组织。

## 计划概览

| 优先级 | 计划 | 描述 | 文件 |
|--------|------|------|------|
| **P0** | 恢复 C# Client.cs | 修复 C# SDK 缺失的客户端功能 | `2024-12-26-restore-csharp-client.md` |
| **P1** | Caller (Python) | Python Caller 示例 | `2024-12-26-p1-caller-python.md` |
| **P1** | Caller (C#) | C# Caller 示例 | `2024-12-26-p1-caller-csharp.md` |
| **P1** | 文件传输示例 | Go/C#/Python 文件传输示例 | `2024-12-26-p1-file-transfer.md` |
| **P1** | 状态管理 (KV) 示例 | Go/C#/Python KV 状态管理示例 | `2024-12-26-p1-state-kv.md` |
| **P2** | 备份功能示例 | Backup Agent 和 Client 示例 | `2024-12-26-p2-backup.md` |
| **P2** | Notify (Go/Python) | Go/Python 发布订阅示例 | `2024-12-26-p2-notify-pubsub.md` |
| **P2** | Python SDK 状态管理 | Python SDK KV 功能实现 | `2024-12-26-p2-python-sdk-state.md` |
| **P3** | C++ SDK 和示例 | C++ SDK 完善 + Provider/Caller/PubSub 示例 | `2024-12-26-p3-cpp.md` |

## 优先级说明

### P0 - 阻塞性问题（必须立即解决）
- 恢复 C# Client.cs：C# SDK 缺少整个客户端代码，阻塞所有 C# 客户端功能

### P1 - 高优先级（核心功能缺失）
- Caller 示例：展示多语言 RPC 调用能力
- 文件传输示例：框架 5 大核心功能之一
- 状态管理示例：NATS JetStream KV 核心功能

### P2 - 中优先级（增强功能）
- 备份功能示例：SDK 已实现，需集成示例
- Notify 示例：补充 Go/Python 的发布订阅示例
- Python SDK 状态管理：补充 Python SDK 的 KV 功能

### P3 - 低优先级（完善阶段）
- C++ SDK 完善：补全头文件、Service 端实现
- C++ 示例：Provider、Caller、PubSub 基础示例

## 执行顺序建议

```
第一阶段：修复阻塞问题
  └─ P0: 恢复 C# Client.cs

第二阶段：核心功能补齐
  ├─ P1: Caller (Python)
  ├─ P1: Caller (C#) - 依赖 P0
  ├─ P1: 状态管理 (KV) 示例
  └─ P1: 文件传输示例

第三阶段：增强功能
  ├─ P2: Notify (Go/Python)
  ├─ P2: 备份功能示例
  └─ P2: Python SDK 状态管理

第四阶段：完善（按需）
  └─ P3: C++ SDK 和示例
```

## 计划详情

### P0: 恢复 C# Client.cs

**文件**: `2024-12-26-restore-csharp-client.md`

**任务数量**: 13

**主要内容**:
- 创建单元测试项目
- 实现连接管理（TLS）
- 实现 RPC Call 方法
- 实现 Publish/Subscribe 方法
- 实现状态管理 (KV) 方法
- 实现文件传输方法
- 更新文档

**预计时间**: 2-3 小时

**依赖**: 无

---

### P1: Caller (Python)

**文件**: `2024-12-26-p1-caller-python.md`

**任务数量**: 6

**主要内容**:
- 创建 Python Caller 项目
- 实现依赖检查
- 实现 RPC 调用
- 编写文档

**预计时间**: 1 小时

**依赖**: 无

---

### P1: Caller (C#)

**文件**: `2024-12-26-p1-caller-csharp.md`

**任务数量**: 6

**主要内容**:
- 创建 C# Caller 项目
- 实现依赖检查
- 实现 RPC 调用
- 编写文档

**预计时间**: 1 小时

**依赖**: P0（C# Client.cs）

---

### P1: 文件传输示例

**文件**: `2024-12-26-p1-file-transfer.md`

**任务数量**: 5

**主要内容**:
- Go 文件传输示例
- C# 文件传输示例（依赖 P0）
- Python 文件传输示例

**预计时间**: 1.5 小时

**依赖**: P0（C# Client.cs）

---

### P1: 状态管理 (KV) 示例

**文件**: `2024-12-26-p1-state-kv.md`

**任务数量**: 5

**主要内容**:
- Go 状态管理示例
- C# 状态管理示例（依赖 P0）
- Python 状态管理示例

**预计时间**: 1.5 小时

**依赖**: P0（C# Client.cs）

---

### P2: 备份功能示例

**文件**: `2024-12-26-p2-backup.md`

**任务数量**: 5

**主要内容**:
- 创建 Backup Agent
- 创建 Backup Client
- 实现版本管理
- 实现增量备份

**预计时间**: 2 小时

**依赖**: 无

---

### P2: Notify (Go/Python)

**文件**: `2024-12-26-p2-notify-pubsub.md`

**任务数量**: 4

**主要内容**:
- Go 发布订阅示例
- Python 发布订阅示例
- 跨语言通信演示

**预计时间**: 1 小时

**依赖**: 无

---

### P2: Python SDK 状态管理

**文件**: `2024-12-26-p2-python-sdk-state.md`

**任务数量**: 8

**主要内容**:
- 添加 KV 辅助方法
- 实现 set_state()
- 实现 get_state()
- 实现 watch_state()
- 实现 delete_state()
- 单元测试
- 文档更新

**预计时间**: 2 小时

**依赖**: 无

---

### P3: C++ SDK 和示例

**文件**: `2024-12-26-p3-cpp.md`

**任务数量**: 7

**主要内容**:
- 完善 C++ SDK 头文件 (client.hpp, service.hpp, types.hpp)
- 完成 Service 类实现
- 创建 C++ Provider 示例 (math-service)
- 创建 C++ Caller 示例
- 创建 C++ PubSub 示例
- 更新文档

**预计时间**: 3-4 小时

**依赖**: 无

---

## 使用说明

每个计划文件都包含：

1. **目标描述** - 清晰说明要实现什么
2. **架构说明** - 技术方案概述
3. **详细任务** - 按步骤分解，每步 2-5 分钟
4. **测试策略** - 如何验证功能
5. **依赖说明** - 前置要求

### 执行单个计划

```bash
# 查看计划详情
cat docs/plans/2024-12-26-restore-csharp-client.md

# 使用 executing-plans skill 执行
# (在新的会话中)
```

### 批量执行

可以按优先级顺序依次执行：
1. 先完成 P0
2. 再完成所有 P1
3. 然后完成 P2
4. 最后按需完成 P3

---

## 相关文档

- [LightLink 项目规则](../../CLAUDE.md)
- [示例目录结构](../../light_link_platform/examples/README.md)
- [功能与示例覆盖分析](../../docs/plans/)

---

## 更新记录

| 日期 | 变更 |
|------|------|
| 2024-12-26 | 创建所有计划文档 |
