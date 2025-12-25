# Remote NATS Configuration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Update all NATS connection configurations to use remote server at 172.18.200.47:4222

**Architecture:** Simple configuration update - replace all hardcoded `localhost:4222` with `172.18.200.47:4222` across all SDKs, examples, and platform components. TLS certificate paths remain unchanged as relative paths.

**Tech Stack:** Go, Python, C#, TypeScript (frontend), YAML config files

---

## Task 1: Update CLAUDE.md documentation

**Files:**
- Modify: `CLAUDE.md`

**Step 1: Add remote NATS server information to NATS service configuration section**

Find the `## NATS 服务配置` section and add:
```markdown
## NATS 服务配置

- 远程服务器地址：`172.18.200.47:4222` (已部署，无需本地启动)
- 配置文件：`deploy/nats/nats-server.conf`
- TLS 证书目录：`deploy/nats/tls/`
- 默认端口：4222
- 需要 JetStream 支持（KV 和 Object Store）
```

**Step 2: Update quick start section - remove local NATS startup**

Find the `### 启动 NATS 服务器` section and replace with:
```markdown
### 连接 NATS 服务器

NATS 服务器已部署在远程地址 `172.18.200.47:4222`，本地调试无需启动。
如需使用环境变量覆盖：`set NATS_URL=nats://custom-address:4222`
```

**Step 3: Verify changes**

Run: `grep "172.18.200.47" CLAUDE.md`
Expected: Shows the remote server address in documentation

**Step 4: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: update CLAUDE.md with remote NATS server address"
```

---

## Task 2: Update Go SDK default config (examples)

**Files:**
- Modify: `examples/config.go:14`

**Step 1: Modify default NATS URL**

Replace line 14:
```go
// Before:
url = "nats://localhost:4222"

// After:
url = "nats://172.18.200.47:4222"
```

**Step 2: Verify changes**

Run: `grep "172.18.200.47" examples/config.go`
Expected: Shows the new default URL

**Step 3: Commit**

```bash
git add examples/config.go
git commit -m "feat: update Go SDK default NATS URL to remote server"
```

---

## Task 3: Update Python SDK client default

**Files:**
- Modify: `sdk/python/lightlink/client.py:20`

**Step 1: Modify Client constructor default parameter**

Replace line 20:
```python
# Before:
def __init__(self, url="nats://localhost:4222", tls_config=None):

# After:
def __init__(self, url="nats://172.18.200.47:4222", tls_config=None):
```

**Step 2: Verify changes**

Run: `grep "172.18.200.47" sdk/python/lightlink/client.py`
Expected: Shows the new default URL

**Step 3: Commit**

```bash
git add sdk/python/lightlink/client.py
git commit -m "feat: update Python SDK default NATS URL to remote server"
```

---

## Task 4: Update manager platform YAML config

**Files:**
- Modify: `light_link_platform/manager_base/server/console.yaml:6`

**Step 1: Modify nats.url in YAML config**

Replace line 6:
```yaml
# Before:
url: "nats://localhost:4222"

# After:
url: "nats://172.18.200.47:4222"
```

**Step 2: Verify changes**

Run: `grep "172.18.200.47" light_link_platform/manager_base/server/console.yaml`
Expected: Shows the new URL

**Step 3: Commit**

```bash
git add light_link_platform/manager_base/server/console.yaml
git commit -m "feat: update manager platform YAML config to remote NATS"
```

---

## Task 5: Update manager platform Go config default

**Files:**
- Modify: `light_link_platform/manager_base/server/config/config.go:115`

**Step 1: Modify GetDefaultConfig NATS URL**

Replace line 115:
```go
// Before:
URL: "nats://localhost:4222",

// After:
URL: "nats://172.18.200.47:4222",
```

**Step 2: Verify changes**

Run: `grep "172.18.200.47" light_link_platform/manager_base/server/config/config.go`
Expected: Shows the new URL

**Step 3: Commit**

```bash
git add light_link_platform/manager_base/server/config/config.go
git commit -m "feat: update manager platform Go config default NATS URL"
```

---

## Task 6: Update Python example service

**Files:**
- Modify: `sdk/python/examples/rpc_service.py:24`
- Modify: `light_link_platform/examples/python/data_service.py:50`

**Step 1: Modify rpc_service.py default URL**

Replace line 24 in `sdk/python/examples/rpc_service.py`:
```python
# Before:
svc = Service("math-service", "nats://localhost:4222")

# After:
svc = Service("math-service", "nats://172.18.200.47:4222")
```

**Step 2: Modify data_service.py default URL**

Replace line 50 in `light_link_platform/examples/python/data_service.py`:
```python
# Before:
nats_url = os.getenv("NATS_URL", "nats://localhost:4222")

# After:
nats_url = os.getenv("NATS_URL", "nats://172.18.200.47:4222")
```

**Step 3: Verify changes**

Run: `grep "172.18.200.47" sdk/python/examples/rpc_service.py light_link_platform/examples/python/data_service.py`
Expected: Shows both files updated

**Step 4: Commit**

```bash
git add sdk/python/examples/rpc_service.py light_link_platform/examples/python/data_service.py
git commit -m "feat: update Python example services to remote NATS"
```

---

## Task 7: Verify C# examples (already updated)

**Files:**
- Check: `light_link_platform/examples/csharp/RpcDemo.cs`
- Check: `light_link_platform/examples/csharp/PubSubDemo.cs`

**Step 1: Verify C# files already use remote URL**

Run: `grep "172.18.200.47" light_link_platform/examples/csharp/RpcDemo.cs light_link_platform/examples/csharp/PubSubDemo.cs`
Expected: Both files already show `"nats://172.18.200.47:4222"` as default

**Note:** These files are already configured correctly - no changes needed.

---

## Task 8: Test the configuration changes

**Step 1: Run Go tests**

Run: `go test ./... -v -short`
Expected: All tests pass

**Step 2: Test manager platform connection**

Run: `cd light_link_platform/manager_base/server && go run main.go`
Expected: Server starts and connects to NATS successfully
Check logs for: "Connected to NATS" or similar success message

**Step 3: Test Python example**

Run: `cd light_link_platform/examples/python && python data_service.py`
Expected: Service starts and registers with NATS successfully

**Step 4: Commit**

```bash
git add -A
git commit -m "test: verify all NATS connections work with remote server"
```

---

## Summary

Files modified:
1. `CLAUDE.md` - Documentation update
2. `examples/config.go` - Go SDK default config
3. `sdk/python/lightlink/client.py` - Python SDK client
4. `light_link_platform/manager_base/server/console.yaml` - Manager YAML config
5. `light_link_platform/manager_base/server/config/config.go` - Manager Go config
6. `sdk/python/examples/rpc_service.py` - Python RPC example
7. `light_link_platform/examples/python/data_service.py` - Python data service

Files already correct (no change needed):
- `light_link_platform/examples/csharp/RpcDemo.cs`
- `light_link_platform/examples/csharp/PubSubDemo.cs`

Environment variable override support is preserved across all files using `NATS_URL`.
