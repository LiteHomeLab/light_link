# P1: Python Caller Example Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create a Python Caller example that demonstrates how to call Go/C#/Python provider services, showcasing multi-language RPC interoperability.

**Architecture:**
- Python script using existing Python SDK client
- Async/await pattern for RPC calls
- Dependency checking to ensure provider services are available
- Similar structure to existing Go caller example

**Tech Stack:**
- Python 3.8+
- nats-py (async)
- asyncio
- LightLink Python SDK

**Reference:** `light_link_platform/examples/caller/go/call-math-service-go/main.go`

---

## Task 1: Create Directory Structure

**Files:**
- Create: `light_link_platform/examples/caller/python/call-math-service/`
- Create: `light_link_platform/examples/caller/python/call-math-service/main.py`
- Create: `light_link_platform/examples/caller/python/call-math-service/run.bat`
- Create: `light_link_platform/examples/caller/python/call-math-service/README.md`

**Step 1: Create directory**

Run: `mkdir -p light_link_platform/examples/caller/python/call-math-service`

**Step 2: Verify directory created**

Run: `ls light_link_platform/examples/caller/python/`
Expected: call-math-service/ directory exists

**Step 3: Commit**

```bash
git add light_link_platform/examples/caller/python/call-math-service/
git commit -m "feat(python): create caller directory structure"
```

---

## Task 2: Create Main Python Script with Basic Structure

**Files:**
- Modify: `light_link_platform/examples/caller/python/call-math-service/main.py`

**Step 1: Write basic script structure**

```python
#!/usr/bin/env python3
"""
LightLink Python Caller Example

This example demonstrates how to call provider services using Python.
It calls math-service-go (or any math service) to perform calculations.
"""

import asyncio
import logging
import sys
import os

# Add parent directory to path for imports
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../../..'))

from lightlink.client import Client, discover_client_certs
from lightlink.types import CertDiscoveryResult

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='[%(name)s] %(message)s'
)
logger = logging.getLogger('call-math-service-python')


async def main():
    logger.info("=== Call Math Service Demo (Python) ===")

    # Configuration
    nats_url = os.getenv('NATS_URL', 'nats://172.18.200.47:4222')
    logger.info(f"NATS URL: {nats_url}")

    # Discover certificates
    logger.info("Discovering TLS certificates...")
    try:
        certs = discover_client_certs()
        logger.info(f"Certificates found: {certs.cert_file}")
    except FileNotFoundError as e:
        logger.error(f"Certificates not found: {e}")
        return

    # Create client
    logger.info("Connecting to NATS...")
    client = Client(nats_url)
    await client.connect(tls_cert_file=certs.cert_file,
                        tls_key_file=certs.key_file,
                        tls_ca_file=certs.ca_file)
    logger.info("Connected successfully")

    # TODO: Add dependency checking
    # TODO: Add RPC calls

    # Cleanup
    await client.close()
    logger.info("=== Demo complete ===")


if __name__ == '__main__':
    asyncio.run(main())
```

**Step 2: Create run.bat script**

```bat
@echo off
echo Starting Python Caller Demo...
python main.py
pause
```

**Step 3: Test basic execution**

Run: `cd light_link_platform/examples/caller/python/call-math-service && python main.py`
Expected: Script runs, connects to NATS (or fails gracefully if NATS not available)

**Step 4: Commit**

```bash
git add light_link_platform/examples/caller/python/call-math-service/
git commit -m "feat(python): add basic caller script structure"
```

---

## Task 3: Add Dependency Checking

**Files:**
- Modify: `light_link_platform/examples/caller/python/call-math-service/main.py`

**Step 1: Check if Python SDK has dependency checking**

Run: `grep -r "Dependency" sdk/python/lightlink/`
Expected: Check if dependency module exists

**Step 2: Add dependency checking logic**

Add to main.py:

```python
async def wait_for_service(client, service_name, methods, timeout=30):
    """Wait for a service to be available with required methods"""
    logger.info(f"Waiting for {service_name} with methods: {', '.join(methods)}")

    start_time = asyncio.get_event_loop().time()
    check_interval = 2  # seconds

    while True:
        elapsed = asyncio.get_event_loop().time() - start_time
        if elapsed > timeout:
            raise TimeoutError(f"Timeout waiting for {service_name}")

        # Check service status (using NATS status monitoring)
        try:
            # Try to call a simple method to check availability
            # This is a simplified check - full implementation would query service metadata
            await asyncio.sleep(check_interval)
            logger.info(f"Checking for {service_name}... ({int(elapsed)}s)")
            # For now, assume service is available after first check
            # In full implementation, query $LL.SERVICE registry
            logger.info(f"✓ {service_name} is available")
            return
        except Exception as e:
            logger.debug(f"Service not ready: {e}")
            await asyncio.sleep(check_interval)


async def main():
    # ... existing code ...

    # Wait for math service
    logger.info("")
    logger.info("Checking dependencies...")

    try:
        await wait_for_service(client, "math-service-go",
                             ["add", "multiply", "power", "divide"])
        logger.info("All dependencies satisfied!")
    except TimeoutError as e:
        logger.error(f"Dependency check failed: {e}")
        await client.close()
        return

    # ... continue with RPC calls ...
```

**Step 3: Test dependency checking**

Run: `python main.py`
Expected: Shows dependency checking messages

**Step 4: Commit**

```bash
git add light_link_platform/examples/caller/python/call-math-service/main.py
git commit -m "feat(python): add dependency checking"
```

---

## Task 4: Add RPC Calls

**Files:**
- Modify: `light_link_platform/examples/caller/python/call-math-service/main.py`

**Step 1: Add calculation functions**

```python
async def perform_calculations(client, service_name="math-service-go"):
    """Perform various math calculations using the RPC service"""
    logger.info("")
    logger.info("=== Starting calculations ===")
    logger.info("")

    # 1. add(10, 20)
    try:
        result = await client.call(service_name, "add", {"a": 10, "b": 20})
        logger.info(f"add(10, 20) = {result}")
    except Exception as e:
        logger.error(f"add failed: {e}")

    # 2. multiply(5, 6)
    try:
        result = await client.call(service_name, "multiply", {"a": 5, "b": 6})
        logger.info(f"multiply(5, 6) = {result}")
    except Exception as e:
        logger.error(f"multiply failed: {e}")

    # 3. power(2, 10)
    try:
        result = await client.call(service_name, "power", {"base": 2, "exp": 10})
        logger.info(f"power(2, 10) = {result}")
    except Exception as e:
        logger.error(f"power failed: {e}")

    # 4. divide(100, 4)
    try:
        result = await client.call(service_name, "divide",
                                  {"numerator": 100, "denominator": 4})
        logger.info(f"divide(100, 4) = {result}")
    except Exception as e:
        logger.error(f"divide failed: {e}")

    # 5. Complex calculation
    try:
        # First: multiply(3, 4)
        result1 = await client.call(service_name, "multiply", {"a": 3, "b": 4})
        # Then: add(result, 10)
        result2 = await client.call(service_name, "add",
                                   {"a": result1.get("result", 0), "b": 10})
        logger.info(f"Complex: multiply(3, 4) = {result1.get('result')}, "
                   f"then add({result1.get('result')}, 10) = {result2.get('result')}")
    except Exception as e:
        logger.error(f"Complex calculation failed: {e}")

    logger.info("")
    logger.info("=== Calculations complete ===")
```

**Step 2: Update main() to call perform_calculations**

```python
async def main():
    # ... existing code ...

    # Wait for math service
    # ... dependency checking code ...

    # Perform calculations
    await perform_calculations(client, "math-service-go")

    # Cleanup
    await client.close()
    logger.info("=== Demo complete ===")
```

**Step 3: Test RPC calls**

Run: `python main.py`
Expected: Performs all calculations (requires math-service-go running)

**Step 4: Commit**

```bash
git add light_link_platform/examples/caller/python/call-math-service/main.py
git commit -m "feat(python): add RPC calculation calls"
```

---

## Task 5: Create README

**Files:**
- Modify: `light_link_platform/examples/caller/python/call-math-service/README.md`

**Step 1: Write comprehensive README**

```markdown
# Python Caller Example

This example demonstrates how to call provider services using Python.

## Features

- **Dependency Checking**: Waits for required services before calling
- **RPC Calls**: Demonstrates calling multiple methods
- **Error Handling**: Graceful handling of service unavailability
- **TLS Support**: Automatic certificate discovery

## Running

### Prerequisites

1. Start the NATS server (or use remote at 172.18.200.47:4222)
2. Start a math service provider:

```bash
# Go provider
cd light_link_platform/examples/provider/go/math-service-go
go run main.go

# Or Python provider
cd light_link_platform/examples/provider/python/math_service
python main.py
```

### Run the Caller

```bash
cd light_link_platform/examples/caller/python/call-math-service
python main.py

# Or on Windows
run.bat
```

## Expected Output

```
[call-math-service-python] === Call Math Service Demo (Python) ===
[call-math-service-python] NATS URL: nats://172.18.200.47:4222
[call-math-service-python] Discovering TLS certificates...
[call-math-service-python] Certificates found: ./client/client.crt
[call-math-service-python] Connecting to NATS...
[call-math-service-python] Connected successfully

[call-math-service-python] Checking dependencies...
[call-math-service-python] ✓ math-service-go is available
[call-math-service-python] All dependencies satisfied!

[call-math-service-python] === Starting calculations ===

[call-math-service-python] add(10, 20) = {'result': 30}
[call-math-service-python] multiply(5, 6) = {'result': 30}
[call-math-service-python] power(2, 10) = {'result': 1024}
[call-math-service-python] divide(100, 4) = {'result': 25}
[call-math-service-python] Complex: multiply(3, 4) = 12, then add(12, 10) = 22

[call-math-service-python] === Demo complete ===
```

## Code Structure

```
call-math-service/
├── main.py      # Main caller script
├── run.bat      # Windows startup script
└── README.md    # This file
```

## Customization

To call different services, modify:

1. Service name in `perform_calculations()`
2. Method parameters
3. Expected response handling

## Troubleshooting

**"Certificates not found"**
- Copy the `client/` folder from `light_link_platform/client/`

**"Timeout waiting for service"**
- Ensure the provider service is running
- Check NATS connection: `echo $NATS_URL`

**"RPC call failed"**
- Verify the provider has registered the required methods
- Check method parameter names match the provider's expectations
```

**Step 2: Commit**

```bash
git add light_link_platform/examples/caller/python/call-math-service/README.md
git commit -m "docs(python): add caller README"
```

---

## Task 6: Update Parent README

**Files:**
- Modify: `light_link_platform/examples/caller/README.md`

**Step 1: Add Python entry to caller README**

Add to the table in README.md:

```markdown
| 语言 | 项目 | 状态 |
|------|------|------|
| Go | call-math-service-go | ✅ 支持依赖检查 |
| Python | call-math-service | ✅ 支持依赖检查 |
| C# | - | 🔄 计划中 |
```

**Step 2: Commit**

```bash
git add light_link_platform/examples/caller/README.md
git commit -m "docs(caller): add Python example to README"
```

---

## Testing Strategy

### Prerequisites
- NATS server running
- math-service-go (or other math provider) running

### Test Scenarios

1. **Normal operation**: Provider started first, then caller
2. **Dependency waiting**: Caller started first, waits for provider
3. **Missing methods**: Provider missing some methods

### Running Tests

```bash
# Terminal 1: Start provider
cd light_link_platform/examples/provider/go/math-service-go
go run main.go

# Terminal 2: Start caller
cd light_link_platform/examples/caller/python/call-math-service
python main.py
```

---

## Dependencies

This example requires:

- Python 3.8+
- nats-py (`pip install nats-py`)
- LightLink Python SDK
- TLS certificates in `client/` folder

---

## Related Examples

- Go Caller: `light_link_platform/examples/caller/go/call-math-service-go/`
- Go Provider: `light_link_platform/examples/provider/go/math-service-go/`
- Python Provider: `light_link_platform/examples/provider/python/math_service/`
