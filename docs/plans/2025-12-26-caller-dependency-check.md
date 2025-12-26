# Caller Dependency Check Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 实现一个带依赖检查功能的 caller 示例程序，启动时验证依赖服务和方法可用后再执行计算

**Architecture:**
1. 在 SDK 中添加依赖检查功能 (`dependency.go`) - 监听 NATS 服务注册消息验证依赖
2. SDK 集成 `github.com/WQGroup/logger` 日志库
3. 创建 `call-math-service-go` 示例程序展示依赖检查和自动调用功能

**Tech Stack:** Go 1.24, NATS, WQGroup/logger, Playwright (测试验证)

---

## Task 1: Add logger dependency to go.mod

**Files:**
- Modify: `go.mod`

**Step 1: Add logger dependency**

```bash
cd C:/WorkSpace/Go2Hell/src/github.com/LiteHomeLab/light_link
go get github.com/WQGroup/logger@latest
```

Expected: `go.mod` updated with logger dependency

**Step 2: Run go mod tidy**

```bash
go mod tidy
```

Expected: Dependencies resolved, `go.sum` updated

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "feat(sdk): add WQGroup/logger dependency"
```

---

## Task 2: Create dependency.go with core types and checker

**Files:**
- Create: `sdk/go/client/dependency.go`
- Test: `sdk/go/client/dependency_test.go`

**Step 1: Write the test for Dependency type**

Create `sdk/go/client/dependency_test.go`:

```go
package client

import (
	"testing"

	"github.com/LiteHomeLab/light_link/sdk/go/types"
	"github.com/stretchr/testify/assert"
)

func TestDependencyType(t *testing.T) {
	dep := Dependency{
		ServiceName: "test-service",
		Methods:     []string{"method1", "method2"},
	}

	assert.Equal(t, "test-service", dep.ServiceName)
	assert.Equal(t, []string{"method1", "method2"}, dep.Methods)
}

func TestDependencyCheckResult(t *testing.T) {
	result := &DependencyCheckResult{
		ServiceName:      "test-service",
		ServiceFound:     true,
		AvailableMethods: []string{"method1"},
		MissingMethods:   []string{"method2"},
		AllSatisfied:     false,
	}

	assert.Equal(t, "test-service", result.ServiceName)
	assert.True(t, result.ServiceFound)
	assert.False(t, result.AllSatisfied)
}
```

**Step 2: Run test to verify it fails**

```bash
cd C:/WorkSpace/Go2Hell/src/github.com/LiteHomeLab/light_link
go test -v ./sdk/go/client -run TestDependency
```

Expected: FAIL with "undefined: Dependency"

**Step 3: Write minimal implementation**

Create `sdk/go/client/dependency.go`:

```go
package client

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/LiteHomeLab/light_link/sdk/go/types"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

// Dependency 依赖定义
type Dependency struct {
	ServiceName string   // 服务名称
	Methods     []string // 必需的方法列表
}

// DependencyCheckResult 依赖检查结果
type DependencyCheckResult struct {
	ServiceName      string   // 服务名称
	ServiceFound     bool     // 服务是否已注册
	AvailableMethods []string // 可用的方法
	MissingMethods   []string // 缺失的方法
	AllSatisfied     bool     // 是否全部满足
}

// DependencyChecker 依赖检查器
type DependencyChecker struct {
	nc         *nats.Conn
	deps       []Dependency
	registered map[string]*types.ServiceMetadata // 已注册的服务
	mu         sync.RWMutex
	sub        *nats.Subscription
	logger     *logrus.Logger
}

// NewDependencyChecker 创建依赖检查器
func NewDependencyChecker(nc *nats.Conn, deps []Dependency) *DependencyChecker {
	return &DependencyChecker{
		nc:         nc,
		deps:       deps,
		registered: make(map[string]*types.ServiceMetadata),
		logger:     logrus.New(),
	}
}

// GetCheckResults 获取当前检查结果
func (dc *DependencyChecker) GetCheckResults() []*DependencyCheckResult {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	results := make([]*DependencyCheckResult, 0, len(dc.deps))

	for _, dep := range dc.deps {
		result := &DependencyCheckResult{
			ServiceName: dep.ServiceName,
		}

		metadata, exists := dc.registered[dep.ServiceName]
		if exists {
			result.ServiceFound = true

			// 检查每个方法
			available := make([]string, 0)
			missing := make([]string, 0)

			methodMap := make(map[string]bool)
			for _, m := range metadata.Methods {
				methodMap[m.Name] = true
			}

			for _, method := range dep.Methods {
				if methodMap[method] {
					available = append(available, method)
				} else {
					missing = append(missing, method)
				}
			}

			result.AvailableMethods = available
			result.MissingMethods = missing
			result.AllSatisfied = len(missing) == 0
		} else {
			result.ServiceFound = false
			result.MissingMethods = dep.Methods
			result.AllSatisfied = false
		}

		results = append(results, result)
	}

	return results
}

// allSatisfied 检查所有依赖是否满足
func (dc *DependencyChecker) allSatisfied() bool {
	results := dc.GetCheckResults()
	for _, r := range results {
		if !r.AllSatisfied {
			return false
		}
	}
	return true
}
```

**Step 4: Run test to verify it passes**

```bash
go test -v ./sdk/go/client -run TestDependency
```

Expected: PASS

**Step 5: Commit**

```bash
git add sdk/go/client/dependency.go sdk/go/client/dependency_test.go
git commit -m "feat(sdk): add DependencyChecker core types and methods"
```

---

## Task 3: Implement WaitForDependencies with NATS subscription

**Files:**
- Modify: `sdk/go/client/dependency.go`

**Step 1: Write test for WaitForDependencies**

Add to `sdk/go/client/dependency_test.go`:

```go
package client

import (
	"context"
	"testing"
	"time"

	"github.com/LiteHomeLab/light_link/sdk/go/types"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDependencyChecker_WaitForDependencies(t *testing.T) {
	// Create test NATS server
	nc, err := nats.Connect(nats.DefaultURL)
	require.NoError(t, err)
	defer nc.Close()

	deps := []Dependency{
		{ServiceName: "test-service", Methods: []string{"method1"}},
	}

	checker := NewDependencyChecker(nc, deps)

	// Start waiting in background
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	doneCh := make(chan error)
	go func() {
		doneCh <- checker.WaitForDependencies(ctx)
	}()

	// Give time for subscription to set up
	time.Sleep(100 * time.Millisecond)

	// Publish registration message
	metadata := &types.ServiceMetadata{
		Name:    "test-service",
		Version: "1.0.0",
		Methods: []types.MethodMetadata{
			{Name: "method1"},
		},
	}

	// Simulate registration message
	registerMsg := &types.RegisterMessage{
		Service:   "test-service",
		Version:   "1.0.0",
		Metadata:  *metadata,
		Timestamp: time.Now().Unix(),
	}

	data, _ := json.Marshal(registerMsg)
	nc.Publish("$LL.register.test-service", data)

	// Should complete without error
	err = <-doneCh
	assert.NoError(t, err)
}
```

**Step 2: Run test to verify it fails**

```bash
go test -v ./sdk/go/client -run TestDependencyChecker_WaitForDependencies
```

Expected: FAIL with "undefined: WaitForDependencies"

**Step 3: Implement WaitForDependencies**

Add to `sdk/go/client/dependency.go`:

```go
import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/LiteHomeLab/light_link/sdk/go/types"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/WQGroup/logger"
)

// ... (existing code)

// WaitForDependencies 等待所有依赖满足
func (dc *DependencyChecker) WaitForDependencies(ctx context.Context) error {
	// Subscribe to registration messages
	sub, err := dc.nc.Subscribe("$LL.register.>", func(msg *nats.Msg) {
		dc.handleRegisterMessage(msg)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}
	dc.sub = sub
	defer sub.Unsubscribe()

	// Initial progress print
	dc.printInitialRequirements()

	// Wait loop
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			dc.PrintProgress()
		}

		if dc.allSatisfied() {
			dc.printAllSatisfied()
			return nil
		}
	}
}

// handleRegisterMessage 处理注册消息
func (dc *DependencyChecker) handleRegisterMessage(msg *nats.Msg) {
	var registerMsg types.RegisterMessage
	if err := json.Unmarshal(msg.Data, &registerMsg); err != nil {
		dc.logger.Warnf("Failed to unmarshal register message: %v", err)
		return
	}

	dc.mu.Lock()
	dc.registered[registerMsg.Service] = &registerMsg.Metadata
	dc.mu.Unlock()

	// Print progress update
	dc.PrintProgress()
}

// printInitialRequirements 打印初始依赖需求
func (dc *DependencyChecker) printInitialRequirements() {
	logger.Info("=== Waiting for dependencies ===")
	logger.Infof("Required services: %d", len(dc.deps))

	totalMethods := 0
	for _, dep := range dc.deps {
		logger.Infof("  - %s (%d methods)", dep.ServiceName, len(dep.Methods))
		totalMethods += len(dep.Methods)
	}

	logger.Infof("Total methods required: %d", totalMethods)
	logger.Info("")
}

// PrintProgress 打印当前进度
func (dc *DependencyChecker) PrintProgress() {
	results := dc.GetCheckResults()

	totalServices := len(results)
	readyServices := 0
	totalMethods := 0
	availableMethods := 0

	for _, r := range results {
		totalMethods += len(r.AvailableMethods) + len(r.MissingMethods)
		availableMethods += len(r.AvailableMethods)
		if r.AllSatisfied {
			readyServices++
		}
	}

	if totalMethods == 0 {
		return
	}

	// Print overall progress
	logger.Infof("Overall progress: %d/%d methods available (%d/%d services ready)",
		availableMethods, totalMethods, readyServices, totalServices)
	logger.Info("")

	// Print each service status
	for _, r := range results {
		logger.Infof("--- %s ---", r.ServiceName)

		if !r.ServiceFound {
			logger.Infof("  Status: Service not found")
			for _, m := range r.MissingMethods {
				logger.Infof("    ✗ %s (service not registered)", m)
			}
		} else {
			logger.Infof("  Status: %d/%d methods available",
				len(r.AvailableMethods), len(r.AvailableMethods)+len(r.MissingMethods))

			for _, m := range r.AvailableMethods {
				logger.Infof("  ✓ %s", m)
			}
			for _, m := range r.MissingMethods {
				logger.Infof("  ✗ %s (not found)", m)
			}
		}
		logger.Info("")
	}
}

// printAllSatisfied 打印所有依赖满足信息
func (dc *DependencyChecker) printAllSatisfied() {
	logger.Info("=== All dependencies satisfied! ===")
	logger.Info("")
	logger.Info("Available services:")
	logger.Info("")

	results := dc.GetCheckResults()
	for _, r := range results {
		logger.Infof("%s (%d/%d methods)", r.ServiceName,
			len(r.AvailableMethods), len(r.AvailableMethods))

		metadata := dc.registered[r.ServiceName]
		for i, m := range metadata.Methods {
			prefix := "  └─"
			if i == len(metadata.Methods)-1 {
				prefix = "  └─"
			}

			returns := "void"
			if len(m.Returns) > 0 {
				retList := make([]string, 0, len(m.Returns))
				for _, ret := range m.Returns {
					retList = append(retList, fmt.Sprintf("%s: %s", ret.Name, ret.Type))
				}
				returns = strings.Join(retList, ", ")
			}

			logger.Infof("%s%s %s", prefix, m.Name, returns)
		}
		logger.Info("")
	}
}

// Close 关闭检查器
func (dc *DependencyChecker) Close() {
	if dc.sub != nil {
		dc.sub.Unsubscribe()
	}
}
```

**Step 4: Run test to verify it passes**

```bash
go test -v ./sdk/go/client -run TestDependencyChecker_WaitForDependencies
```

Expected: PASS

**Step 5: Commit**

```bash
git add sdk/go/client/dependency.go sdk/go/client/dependency_test.go
git commit -m "feat(sdk): implement WaitForDependencies with progress logging"
```

---

## Task 4: Integrate logger into SDK client files

**Files:**
- Modify: `sdk/go/client/connection.go`
- Modify: `sdk/go/client/rpc.go`

**Step 1: Modify connection.go to use logger**

Read `sdk/go/client/connection.go` first to see what needs changing.

Then modify to add logger import and replace println statements:

```go
package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/LiteHomeLab/light_link/sdk/go/types"
	"github.com/WQGroup/logger"
)

// ... (existing code)

// In NewClient, add logger initialization:
func NewClient(url string, opts ...Option) (*Client, error) {
	// Initialize logger
	logger.SetLoggerName("LightLink-Client")

	// ... rest of existing code

	natsOpts := []nats.Option{
		nats.Name(client.name),
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(10),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				logger.Errorf("Disconnected: %s", err.Error())
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logger.Infof("Reconnected to %s", nc.ConnectedUrl())
		}),
	}

	// ... rest of existing code
}
```

**Step 2: Modify rpc.go to use logger**

```go
package client

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/LiteHomeLab/light_link/sdk/go/types"
	"github.com/WQGroup/logger"
)

// Call makes a synchronous RPC call
func (c *Client) Call(service, method string, args map[string]interface{}) (map[string]interface{}, error) {
	return c.CallWithTimeout(service, method, args, 5*time.Second)
}

// CallWithTimeout makes an RPC call with timeout
func (c *Client) CallWithTimeout(service, method string, args map[string]interface{}, timeout time.Duration) (map[string]interface{}, error) {
	logger.Debugf("Calling %s.%s with args: %+v", service, method, args)

	// ... existing RPC call code ...

	if !response.Success {
		logger.Errorf("RPC error: %s", response.Error)
		return nil, fmt.Errorf("RPC error: %s", response.Error)
	}

	logger.Debugf("RPC response: %+v", response.Result)
	return response.Result, nil
}
```

**Step 3: Run tests**

```bash
go test -v ./sdk/go/client
```

Expected: All existing tests still pass

**Step 4: Commit**

```bash
git add sdk/go/client/connection.go sdk/go/client/rpc.go
git commit -m "feat(sdk): integrate WQGroup/logger into client"
```

---

## Task 5: Create call-math-service-go directory structure

**Files:**
- Create: `light_link_platform/examples/caller/go/call-math-service-go/main.go`
- Create: `light_link_platform/examples/caller/go/call-math-service-go/go.mod`

**Step 1: Create directory**

```bash
mkdir -p C:/WorkSpace/Go2Hell/src/github.com/LiteHomeLab/light_link/light_link_platform/examples/caller/go/call-math-service-go
cd C:/WorkSpace/Go2Hell/src/github.com/LiteHomeLab/light_link/light_link_platform/examples/caller/go/call-math-service-go
go mod init github.com/LiteHomeLab/light_link/light_link_platform/examples/caller/go/call-math-service-go
go mod edit -replace github.com/LiteHomeLab/light_link=../../../../../../..
go get github.com/LiteHomeLab/light_link/sdk/go/client
go get github.com/LiteHomeLab/light_link/sdk/go/types
go get github.com/LiteHomeLab/light_link/examples
go mod tidy
```

Expected: `go.mod` and `go.sum` created

**Step 2: Write main.go**

Create `light_link_platform/examples/caller/go/call-math-service-go/main.go`:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/LiteHomeLab/light_link/examples"
	"github.com/LiteHomeLab/light_link/sdk/go/client"
	"github.com/WQGroup/logger"
)

func main() {
	logger.SetLoggerName("call-math-service-go")
	logger.Info("=== Call Math Service Demo ===")

	config := examples.GetConfig()
	logger.Infof("NATS URL: %s", config.NATSURL)

	// Create client
	logger.Info("Connecting to NATS...")
	cli, err := client.NewClient(config.NATSURL, client.WithAutoTLS())
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer cli.Close()
	logger.Info("Connected successfully")

	// Define dependencies
	deps := []client.Dependency{
		{
			ServiceName: "math-service-go",
			Methods:     []string{"add", "multiply", "power", "divide"},
		},
	}

	// Wait for dependencies
	checker := client.NewDependencyChecker(cli.GetNATSConn(), deps)
	if err := checker.WaitForDependencies(context.Background()); err != nil {
		log.Fatalf("Failed to wait for dependencies: %v", err)
	}
	defer checker.Close()

	// Perform calculations
	performCalculations(cli)

	logger.Info("=== Demo complete ===")
}

func performCalculations(cli *client.Client) {
	logger.Info("")
	logger.Info("=== Starting calculations ===")
	logger.Info("")

	// 1. add(10, 20)
	result, err := cli.Call("math-service-go", "add", map[string]interface{}{
		"a": float64(10),
		"b": float64(20),
	})
	if err != nil {
		logger.Errorf("add failed: %v", err)
	} else {
		logger.Infof("add(10, 20) = %v", result)
	}

	// 2. multiply(5, 6)
	result, err = cli.Call("math-service-go", "multiply", map[string]interface{}{
		"a": float64(5),
		"b": float64(6),
	})
	if err != nil {
		logger.Errorf("multiply failed: %v", err)
	} else {
		logger.Infof("multiply(5, 6) = %v", result)
	}

	// 3. power(2, 10)
	result, err = cli.Call("math-service-go", "power", map[string]interface{}{
		"base": float64(2),
		"exp":   float64(10),
	})
	if err != nil {
		logger.Errorf("power failed: %v", err)
	} else {
		logger.Infof("power(2, 10) = %v", result)
	}

	// 4. divide(100, 4)
	result, err = cli.Call("math-service-go", "divide", map[string]interface{}{
		"numerator":   float64(100),
		"denominator": float64(4),
	})
	if err != nil {
		logger.Errorf("divide failed: %v", err)
	} else {
		logger.Infof("divide(100, 4) = %v", result)
	}

	// 5. Complex calculation: add(multiply(3, 4), divide(20, 2))
	// First: multiply(3, 4)
	result, err = cli.Call("math-service-go", "multiply", map[string]interface{}{
		"a": float64(3),
		"b": float64(4),
	})
	if err != nil {
		logger.Errorf("Complex calculation multiply failed: %v", err)
		return
	}
	product := result["product"].(float64)

	// Second: divide(20, 2)
	result, err = cli.Call("math-service-go", "divide", map[string]interface{}{
		"numerator":   float64(20),
		"denominator": float64(2),
	})
	if err != nil {
		logger.Errorf("Complex calculation divide failed: %v", err)
		return
	}
	quotient := result["quotient"].(float64)

	// Third: add(product, quotient)
	result, err = cli.Call("math-service-go", "add", map[string]interface{}{
		"a": product,
		"b": quotient,
	})
	if err != nil {
		logger.Errorf("Complex calculation add failed: %v", err)
	} else {
		logger.Infof("Complex: add(multiply(3, 4), divide(20, 2)) = add(%.0f, %.0f) = %v",
			product, quotient, result)
	}
}
```

**Step 3: Commit**

```bash
git add light_link_platform/examples/caller/go/call-math-service-go/
git commit -m "feat(caller): add call-math-service-go example with dependency check"
```

---

## Task 6: Build and test call-math-service-go

**Files:**
- Build: `light_link_platform/examples/caller/go/call-math-service-go/main.go`

**Step 1: Build the program**

```bash
cd C:/WorkSpace/Go2Hell/src/github.com/LiteHomeLab/light_link/light_link_platform/examples/caller/go/call-math-service-go
go build -o call-math-service.exe main.go
```

Expected: `call-math-service.exe` created

**Step 2: Verify build works**

```bash
./call-math-service.exe --help 2>&1 | head -1
```

Expected: Program starts (will wait for dependencies)

**Step 3: Create start script**

Create `light_link_platform/examples/caller/go/call-math-service-go/run.bat`:

```batch
@echo off
cd /d %~dp0
echo Starting call-math-service-go...
go run main.go
```

**Step 4: Commit**

```bash
git add light_link_platform/examples/caller/go/call-math-service-go/
git commit -m "feat(caller): add build and start script for call-math-service-go"
```

---

## Task 7: Integration test with Playwright

**Files:**
- Test: Manual verification using browser

**Step 1: Start manager server**

```bash
cd C:/WorkSpace/Go2Hell/src/github.com/LiteHomeLab/light_link/light_link_platform/manager_base/server
go run main.go
```

Wait for: "Server starting on 0.0.0.0:8080"

**Step 2: Start web frontend**

```bash
cd C:/WorkSpace/Go2Hell/src/github.com/LiteHomeLab/light_link/light_link_platform/manager_base/web
npm run dev
```

Wait for: "Local: http://localhost:5173"

**Step 3: Open browser and check services**

Use Playwright to:
1. Navigate to http://localhost:5173
2. Login as admin
3. Go to Services page
4. Verify math-service-go is NOT listed yet

**Step 4: Start math-service-go provider**

```bash
cd C:/WorkSpace/Go2Hell/src/github.com/LiteHomeLab/light_link/light_link_platform/examples/provider/go/math-service-go
go run main.go
```

**Step 5: Verify service appears in UI**

Use Playwright to refresh Services page and verify:
- math-service-go appears with "在线" status
- Click into detail view
- Verify all 4 methods listed: add, multiply, power, divide

**Step 6: Start call-math-service-go caller**

```bash
cd C:/WorkSpace/Go2Hell/src/github.com/LiteHomeLab/light_link/light_link_platform/examples/caller/go/call-math-service-go
go run main.go
```

Expected log output:
```
[INFO] === Call Math Service Demo ===
[INFO] NATS URL: nats://172.18.200.47:4222
[INFO] Connecting to NATS...
[INFO] Connected successfully
[INFO] === Waiting for dependencies ===
[INFO] Required services: 1
[INFO]   - math-service-go (4 methods)
[INFO] Total methods required: 4
[INFO]
[INFO] Overall progress: 4/4 methods available (1/1 services ready)
[INFO]
[INFO] --- math-service-go ---
[INFO]   Status: 4/4 methods available
[INFO]   ✓ add
[INFO]   ✓ multiply
[INFO]   ✓ power
[INFO]   ✓ divide
[INFO]
[INFO] === All dependencies satisfied! ===
[INFO]
[INFO] Available services:
[INFO]
[INFO] math-service-go (4/4 methods)
[INFO]   add sum: number
[INFO]   multiply product: number
[INFO]   power result: number
[INFO]   divide quotient: number
[INFO]
[INFO] === Starting calculations ===
[INFO]
[INFO] add(10, 20) = map[sum:30]
[INFO] multiply(5, 6) = map[product:30]
[INFO] power(2, 10) = map[result:1024]
[INFO] divide(100, 4) = map[quotient:25]
[INFO] Complex: add(multiply(3, 4), divide(20, 2)) = add(12, 10) = map[sum:22]
[INFO]
[INFO] === Demo complete ===
```

**Step 7: Stop all services and verify caller waits**

1. Stop math-service-go
2. Restart call-math-service-go
3. Verify it shows waiting state with 0/4 methods
4. Start math-service-go
5. Verify caller detects all methods and proceeds

**Step 8: Update caller README**

Update `light_link_platform/examples/caller/README.md`:

```markdown
# Caller - 服务调用者

本目录包含调用 Provider 服务的客户端示例程序。

## 示例项目

### call-math-service-go (Go)

调用 math-service-go 的数学计算服务，展示依赖检查功能。

**特性:**
- 启动时验证依赖服务和方法可用性
- 等待所有依赖满足后自动执行计算
- 实时显示依赖检查进度

**运行:**
\`\`\`bash
cd go/call-math-service-go
go run main.go
\`\`\`

**预期输出:**
\`\`\`
[INFO] === Call Math Service Demo ===
[INFO] Connecting to NATS...
[INFO] Waiting for dependencies: math-service-go (add, multiply, power, divide)
[INFO] All dependencies satisfied!
[INFO] add(10, 20) = 30
[INFO] multiply(5, 6) = 30
...
\`\`\`
```

**Step 9: Commit**

```bash
git add light_link_platform/examples/caller/README.md
git commit -m "docs(caller): update README with call-math-service-go documentation"
```

---

## Task 8: Final verification and documentation

**Files:**
- Create: `docs/caller-dependency-check.md`

**Step 1: Write documentation**

Create `docs/caller-dependency-check.md`:

```markdown
# Caller 依赖检查功能

## 概述

从 SDK v0.x 开始，Client 模块支持依赖检查功能。Caller 程序可以在启动时验证依赖的服务和方法是否可用。

## 使用方法

### 1. 定义依赖

\`\`\`go
deps := []client.Dependency{
    {
        ServiceName: "math-service-go",
        Methods:     []string{"add", "multiply", "power", "divide"},
    },
    {
        ServiceName: "text-service",
        Methods:     []string{"uppercase", "lowercase"},
    },
}
\`\`\`

### 2. 创建检查器并等待

\`\`\`go
checker := client.NewDependencyChecker(nc, deps)
err := checker.WaitForDependencies(context.Background())
if err != nil {
    log.Fatalf("Dependencies not satisfied: %v", err)
}
\`\`\`

### 3. 日志输出

检查器会自动输出依赖检查进度：

\`\`\`
[INFO] === Waiting for dependencies ===
[INFO] Required services: 2
[INFO]   - math-service-go (4 methods)
[INFO]   - text-service (2 methods)
[INFO] Overall progress: 3/6 methods available (0/2 services ready)
[INFO] --- math-service-go ---
[INFO]   Status: 3/4 methods available
[INFO]   ✓ add
[INFO]   ✓ multiply
[INFO]   ✗ power (not found)
...
\`\`\`

## 示例

完整示例请参考：`light_link_platform/examples/caller/go/call-math-service-go`
\`\`\`

**Step 2: Final commit**

```bash
git add docs/caller-dependency-check.md
git commit -m "docs: add caller dependency check documentation"
```

**Step 3: Run full test suite**

```bash
cd C:/WorkSpace/Go2Hell/src/github.com/LiteHomeLab/light_link
go test -v ./...
```

Expected: All tests pass

**Step 4: Create summary commit**

```bash
git add -A
git commit -m "feat(caller): complete dependency check implementation

- Add WQGroup/logger to SDK
- Implement DependencyChecker in client
- Create call-math-service-go example
- Add comprehensive documentation
"
```

---

## Summary

This plan implements a complete caller dependency checking system:

1. **SDK Enhancement**: Add `DependencyChecker` to `sdk/go/client/dependency.go`
2. **Logger Integration**: Use `WQGroup/logger` throughout SDK
3. **Example Program**: `call-math-service-go` demonstrates the feature
4. **Testing**: Integration tests with Playwright for UI verification

**Total tasks**: 8
**Estimated files to create**: 5
**Estimated files to modify**: 4
