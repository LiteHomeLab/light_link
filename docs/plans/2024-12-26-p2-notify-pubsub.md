# P2: Notify (Go/Python) PubSub Examples Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create publish/subscribe examples in Go and Python that demonstrate message broadcasting and real-time communication.

**Architecture:**
- Go PubSub example with publisher and subscriber
- Python PubSub example with publisher and subscriber
- Demonstrates:
  - Publishing messages
  - Subscribing to subjects
  - Message handling
  - Multiple subscribers

**Tech Stack:**
- Go 1.21+
- Python 3.8+
- NATS core messaging

**Reference:** `examples/pubsub-demo/main.go`, `light_link_platform/examples/notify/csharp/PubSubDemo/`

---

## Task 1: Create Directory Structure

**Files:**
- Create: `light_link_platform/examples/notify/go/pubsub-demo/`

**Step 1: Create directory**

Run: `mkdir -p light_link_platform/examples/notify/go/pubsub-demo`

**Step 2: Commit**

```bash
git add light_link_platform/examples/notify/go/pubsub-demo/
git commit -m "feat(notify): create Go pubsub directory"
```

---

## Task 2: Implement Go PubSub Example

**Files:**
- Create: `light_link_platform/examples/notify/go/pubsub-demo/main.go`
- Create: `light_link_platform/examples/notify/go/pubsub-demo/run.bat`
- Create: `light_link_platform/examples/notify/go/pubsub-demo/README.md`

**Step 1: Write Go pubsub example**

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LiteHomeLab/light_link/examples"
	"github.com/LiteHomeLab/light_link/sdk/go/client"
	"github.com/WQGroup/logger"
)

func main() {
	logger.SetLoggerName("pubsub-demo-go")
	logger.Info("=== Publish/Subscribe Demo (Go) ===")

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

	// Get subject from args or use default
	subject := "demo.go.pubsub"
	if len(os.Args) > 1 {
		subject = os.Args[1]
	}

	// Subscribe to messages
	logger.Infof("\nSubscribing to: %s", subject)
	sub, err := cli.Subscribe(subject, func(data map[string]interface{}) {
		handleMessage(data)
	})
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}
	defer sub.Unsubscribe()

	// Publish some test messages
	logger.Info("\nPublishing test messages...")
	go publishMessages(cli, subject)

	// Wait for interrupt
	logger.Info("\nRunning... (Press Ctrl+C to stop)")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("\n=== Demo complete ===")
}

func handleMessage(data map[string]interface{}) {
	// Parse message
	msgType, _ := data["type"].(string)
	timestamp, _ := data["timestamp"].(string)

	logger.Infof("Received message:")
	logger.Infof("  Type: %s", msgType)
	logger.Infof("  Timestamp: %s", timestamp)

	switch msgType {
	case "greeting":
		message, _ := data["message"].(string)
		count, _ := data["count"].(int)
		logger.Infof("  Message: %s", message)
		logger.Infof("  Count: %d", count)

	case "status":
		status, _ := data["status"].(string)
		progress, _ := data["progress"].(float64)
		logger.Infof("  Status: %s", status)
		logger.Infof("  Progress: %.0f%%", progress)

	case "alert":
		level, _ := data["level"].(string)
		message, _ := data["message"].(string)
		logger.Infof("  Level: %s", level)
		logger.Infof("  Message: %s", message)

	default:
		// Print all fields
		for k, v := range data {
			logger.Infof("  %s: %v", k, v)
		}
	}

	fmt.Println() // Empty line for readability
}

func publishMessages(cli *client.Client, subject string) {
	time.Sleep(500 * time.Millisecond) // Wait for subscription to be ready

	// Message 1: Greeting
	msg1 := map[string]interface{}{
		"type":      "greeting",
		"timestamp": time.Now().Format(time.RFC3339),
		"message":   "Hello from Go!",
		"count":     42,
	}
	publish(cli, subject, msg1)
	time.Sleep(1 * time.Second)

	// Message 2: Status update
	msg2 := map[string]interface{}{
		"type":      "status",
		"timestamp": time.Now().Format(time.RFC3339),
		"status":    "processing",
		"progress":  75.5,
	}
	publish(cli, subject, msg2)
	time.Sleep(1 * time.Second)

	// Message 3: Alert
	msg3 := map[string]interface{}{
		"type":      "alert",
		"timestamp": time.Now().Format(time.RFC3339),
		"level":     "warning",
		"message":   "Task completed successfully",
	}
	publish(cli, subject, msg3)
	time.Sleep(1 * time.Second)

	// Message 4: Custom data
	msg4 := map[string]interface{}{
		"type":      "custom",
		"timestamp": time.Now().Format(time.RFC3339),
		"user":      "alice",
		"action":    "login",
		"ip":        "192.168.1.100",
	}
	publish(cli, subject, msg4)
}

func publish(cli *client.Client, subject string, data map[string]interface{}) {
	jsonData, _ := json.Marshal(data)
	if err := cli.Publish(subject, data); err != nil {
		logger.Errorf("Failed to publish: %v", err)
		return
	}
	logger.Infof("Published: %s", string(jsonData))
}
```

**Step 2: Create run.bat**

```bat
@echo off
echo Starting Go PubSub Demo...
go run main.go %1
pause
```

**Step 3: Create README.md**

```markdown
# Go Publish/Subscribe Demo

This example demonstrates publish/subscribe messaging with LightLink.

## Running

```bash
# Default subject
go run main.go

# Custom subject
go run main.go "my.custom.subject"
```

## How It Works

1. **Subscribe** - Client subscribes to a subject
2. **Publish** - Client publishes messages to the same subject
3. **Receive** - Subscriber receives and processes messages

## Message Types

The demo publishes different message types:

- **greeting** - Simple text message
- **status** - Status update with progress
- **alert** - Alert/notification message
- **custom** - Custom data structure

## Multiple Subscribers

You can run multiple instances to demonstrate broadcast messaging:

```bash
# Terminal 1
go run main.go

# Terminal 2
go run main.go

# Terminal 3 - publish only
go run main.go
```

All subscribers will receive the same messages.

## Expected Output

```
[pubsub-demo-go] === Publish/Subscribe Demo (Go) ===
[pubsub-demo-go] Connected successfully

[pubsub-demo-go] Subscribing to: demo.go.pubsub
[pubsub-demo-go] Publishing test messages...
[pubsub-demo-go] Published: {"count":42,"message":"Hello from Go!","timestamp":"...","type":"greeting"}

[pubsub-demo-go] Received message:
[pubsub-demo-go]   Type: greeting
[pubsub-demo-go]   Timestamp: 2024-12-26T10:30:00Z
[pubsub-demo-go]   Message: Hello from Go!
[pubsub-demo-go]   Count: 42
```
```

**Step 4: Test the example**

Run: `go run main.go`
Expected: Publishes and receives messages

**Step 5: Commit**

```bash
git add light_link_platform/examples/notify/go/pubsub-demo/
git commit -m "feat(notify): add Go pubsub example"
```

---

## Task 3: Implement Python PubSub Example

**Files:**
- Create: `light_link_platform/examples/notify/python/pubsub_demo/`
- Create: `light_link_platform/examples/notify/python/pubsub_demo/main.py`
- Create: `light_link_platform/examples/notify/python/pubsub_demo/run.bat`
- Create: `light_link_platform/examples/notify/python/pubsub_demo/README.md`

**Step 1: Write Python pubsub example**

```python
#!/usr/bin/env python3
"""
LightLink Python Publish/Subscribe Demo

Demonstrates pub/sub messaging with LightLink.
"""

import asyncio
import json
import logging
import os
import sys
import time
from datetime import datetime

# Add parent directory to path for imports
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../../..'))

from lightlink.client import Client, discover_client_certs

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='[%(name)s] %(message)s'
)
logger = logging.getLogger('pubsub-demo-python')


async def handle_message(data):
    """Handle received message"""
    msg_type = data.get('type', 'unknown')
    timestamp = data.get('timestamp', '')

    logger.info("Received message:")
    logger.info(f"  Type: {msg_type}")
    logger.info(f"  Timestamp: {timestamp}")

    if msg_type == 'greeting':
        message = data.get('message', '')
        count = data.get('count', 0)
        logger.info(f"  Message: {message}")
        logger.info(f"  Count: {count}")

    elif msg_type == 'status':
        status = data.get('status', '')
        progress = data.get('progress', 0)
        logger.info(f"  Status: {status}")
        logger.info(f"  Progress: {progress}%")

    elif msg_type == 'alert':
        level = data.get('level', '')
        message = data.get('message', '')
        logger.info(f"  Level: {level}")
        logger.info(f"  Message: {message}")

    else:
        # Print all fields
        for k, v in data.items():
            logger.info(f"  {k}: {v}")

    print()  # Empty line for readability


async def publish_messages(client, subject):
    """Publish test messages"""
    await asyncio.sleep(0.5)  # Wait for subscription to be ready

    # Message 1: Greeting
    msg1 = {
        'type': 'greeting',
        'timestamp': datetime.now().isoformat(),
        'message': 'Hello from Python!',
        'count': 100
    }
    await publish(client, subject, msg1)
    await asyncio.sleep(1)

    # Message 2: Status update
    msg2 = {
        'type': 'status',
        'timestamp': datetime.now().isoformat(),
        'status': 'processing',
        'progress': 50.0
    }
    await publish(client, subject, msg2)
    await asyncio.sleep(1)

    # Message 3: Alert
    msg3 = {
        'type': 'alert',
        'timestamp': datetime.now().isoformat(),
        'level': 'info',
        'message': 'Python demo running'
    }
    await publish(client, subject, msg3)
    await asyncio.sleep(1)

    # Message 4: Custom data
    msg4 = {
        'type': 'custom',
        'timestamp': datetime.now().isoformat(),
        'user': 'bob',
        'action': 'logout',
        'session': 'abc123'
    }
    await publish(client, subject, msg4)


async def publish(client, subject, data):
    """Publish a message"""
    try:
        await client.publish(subject, data)
        logger.info(f"Published: {json.dumps(data)}")
    except Exception as e:
        logger.error(f"Failed to publish: {e}")


async def main():
    logger.info("=== Publish/Subscribe Demo (Python) ===")

    # Configuration
    nats_url = os.getenv('NATS_URL', 'nats://172.18.200.47:4222')
    subject = sys.argv[1] if len(sys.argv) > 1 else 'demo.python.pubsub'
    logger.info(f"NATS URL: {nats_url}")

    # Discover certificates
    logger.info("\nDiscovering TLS certificates...")
    try:
        certs = discover_client_certs()
        logger.info(f"Certificates found: {certs.cert_file}")
    except FileNotFoundError as e:
        logger.error(f"Certificates not found: {e}")
        return

    # Create client
    logger.info("\nConnecting to NATS...")
    client = Client(nats_url)
    await client.connect(
        tls_cert_file=certs.cert_file,
        tls_key_file=certs.key_file,
        tls_ca_file=certs.ca_file
    )
    logger.info("Connected successfully")

    # Subscribe to messages
    logger.info(f"\nSubscribing to: {subject}")
    try:
        sub = await client.subscribe(subject, handle_message)
    except Exception as e:
        logger.error(f"Failed to subscribe: {e}")
        await client.close()
        return

    # Publish test messages
    logger.info("\nPublishing test messages...")
    asyncio.create_task(publish_messages(client, subject))

    # Run for a while
    logger.info("\nRunning... (Press Ctrl+C to stop)")
    try:
        await asyncio.sleep(10)
    except KeyboardInterrupt:
        pass

    # Cleanup
    await sub.unsubscribe()
    await client.close()
    logger.info("\n=== Demo complete ===")


if __name__ == '__main__':
    asyncio.run(main())
```

**Step 2: Create run.bat**

```bat
@echo off
echo Starting Python PubSub Demo...
python main.py %1
pause
```

**Step 3: Create README.md**

```markdown
# Python Publish/Subscribe Demo

This example demonstrates publish/subscribe messaging with LightLink.

## Running

```bash
# Default subject
python main.py

# Custom subject
python main.py "my.custom.subject"
```

## How It Works

1. **Subscribe** - Client subscribes to a subject
2. **Publish** - Client publishes messages to the same subject
3. **Receive** - Subscriber receives and processes messages

## Message Types

- **greeting** - Simple text message
- **status** - Status update with progress
- **alert** - Alert/notification message
- **custom** - Custom data structure

## Cross-Language Demo

You can run Go and Python demos together to demonstrate cross-language messaging:

```bash
# Terminal 1 - Go subscriber
cd light_link_platform/examples/notify/go/pubsub-demo
go run main.go "demo.cross"

# Terminal 2 - Python publisher
cd light_link_platform/examples/notify/python/pubsub_demo
python main.py "demo.cross"
```

The Go subscriber will receive messages from the Python publisher.

## Expected Output

```
[pubsub-demo-python] === Publish/Subscribe Demo (Python) ===
[pubsub-demo-python] Connected successfully

[pubsub-demo-python] Subscribing to: demo.python.pubsub
[pubsub-demo-python] Publishing test messages...
[pubsub-demo-python] Published: {"count":100,"message":"Hello from Python!"...}

[pubsub-demo-python] Received message:
[pubsub-demo-python]   Type: greeting
[pubsub-demo-python]   Timestamp: 2024-12-26T10:30:00Z
[pubsub-demo-python]   Message: Hello from Python!
[pubsub-demo-python]   Count: 100
```
```

**Step 4: Test the example**

Run: `python main.py`
Expected: Publishes and receives messages

**Step 5: Commit**

```bash
git add light_link_platform/examples/notify/python/pubsub_demo/
git commit -m "feat(notify): add Python pubsub example"
```

---

## Task 4: Update Notify README

**Files:**
- Modify: `light_link_platform/examples/notify/README.md`

**Step 1: Update with pubsub examples**

```markdown
# Notify - 消息通知与状态管理

本目录包含使用消息发布订阅和状态管理的示例程序。

## 示例项目

| 语言 | 发布订阅 | 状态管理 (KV) | 说明 |
|------|----------|--------------|------|
| Go | ✅ pubsub-demo | ✅ state-demo | 发布订阅和 KV 存储 |
| C# | ✅ PubSubDemo | ✅ StateDemo | 发布订阅和 KV 存储 |
| Python | ✅ pubsub_demo | ✅ state_demo | 发布订阅和 KV 存储 |

## 发布/订阅功能

### 1. 发布消息
向指定主题发送消息，所有订阅者都会收到

### 2. 订阅消息
接收指定主题的消息，支持多个订阅者

### 3. 消息类型
支持任意 JSON 格式的消息数据

### 4. 多语言互通
Go、C#、Python 可以相互通信

## 运行示例

### Go 发布订阅
```bash
cd light_link_platform/examples/notify/go/pubsub-demo
go run main.go
```

### C# 发布订阅
```bash
cd light_link_platform/examples/notify/csharp/PubSubDemo
dotnet run
```

### Python 发布订阅
```bash
cd light_link_platform/examples/notify/python/pubsub_demo
python main.py
```

### 跨语言测试
```bash
# 终端 1: Go 订阅者
cd light_link_platform/examples/notify/go/pubsub-demo
go run main.go "test.cross"

# 终端 2: Python 发布者
cd light_link_platform/examples/notify/python/pubsub_demo
python main.py "test.cross"
```
```

**Step 2: Commit**

```bash
git add light_link_platform/examples/notify/README.md
git commit -m "docs(notify): update README with pubsub examples"
```

---

## Testing Strategy

### Prerequisites
- NATS server running
- TLS certificates in `client/` folder

### Test Each Language

```bash
# Go
cd light_link_platform/examples/notify/go/pubsub-demo
go run main.go

# C#
cd light_link_platform/examples/notify/csharp/PubSubDemo
dotnet run

# Python
cd light_link_platform/examples/notify/python/pubsub_demo
python main.py
```

### Cross-Language Test

1. Start Go subscriber on `test.cross`
2. Start Python publisher on `test.cross`
3. Verify Go receives Python's messages

---

## Dependencies

### Go
- Go 1.21+
- LightLink Go SDK

### Python
- Python 3.8+
- nats-py
- LightLink Python SDK

---

## Acceptance Testing via Management Platform

**IMPORTANT:** All development plans must be verified through the management platform.

### Step 1: Start Management Platform Backend

```bash
cd light_link_platform/manager_base/server
go run main.go
```

Wait for the backend server to start.

### Step 2: Start Management Platform Frontend

```bash
cd light_link_platform/manager_base/web
npm run dev
```

Wait for the frontend to start.

### Step 3: Open Browser and Verify

1. Open browser to the frontend URL
2. Navigate to the Messages/PubSub section
3. Verify that:
   - Published messages appear in real-time
   - Message subjects are organized
   - Message content is displayed correctly

### Step 4: Test PubSub Flow

1. Run pubsub examples (Go and/or Python)
2. Observe in the management platform:
   - Messages being published
   - Subscribers receiving messages
   - Cross-language message exchange

### Step 5: Capture Evidence

Take screenshots showing message flow in the management platform.
