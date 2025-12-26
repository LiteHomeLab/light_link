# P2: Python SDK State Management Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement state management (KV) functionality in the Python SDK, enabling Python services to store, retrieve, and watch state changes using NATS JetStream KV.

**Architecture:**
- Add state management methods to Python SDK's client.py
- Methods: `set_state()`, `get_state()`, `watch_state()`, `delete_state()`
- Uses NATS JetStream KV for persistence
- Follows the same patterns as Go SDK

**Tech Stack:**
- Python 3.8+
- nats-py with JetStream support
- asyncio

**Reference:** `sdk/go/client/state.go`, `sdk/python/lightlink/client.py`

---

## Task 1: Review Python SDK Structure

**Files:**
- Read: `sdk/python/lightlink/client.py`
- Read: `sdk/go/client/state.go`

**Step 1: Examine current Python client structure**

Run: `head -100 sdk/python/lightlink/client.py`
Expected: Understand current implementation

**Step 2: Examine Go state implementation for reference**

Run: `cat sdk/go/client/state.go`
Expected: Understand the state management pattern

**Step 3: Check JetStream support in Python SDK**

Run: `grep -n "jetstream\|JetStream" sdk/python/lightlink/client.py`
Expected: Check if JetStream context is already initialized

---

## Task 2: Add KV Helper Methods

**Files:**
- Modify: `sdk/python/lightlink/client.py`

**Step 1: Add GetOrCreateKeyValue helper method**

Add to Client class:

```python
async def _get_or_create_kv(self, bucket_name: str) -> nats.js.KV:
    """
    Get or create a KV bucket.

    Args:
        bucket_name: Name of the KV bucket

    Returns:
        KV bucket instance
    """
    if self._js is None:
        raise RuntimeError("JetStream not initialized. Call connect() first.")

    try:
        # Try to get existing bucket
        kv = await self._js.key_value(bucket_name)
        return kv
    except nats.errors.NotFoundError:
        # Create new bucket
        try:
            kv = await self._js.create_key_value(
                bucket_name=bucket_name,
            )
            return kv
        except Exception as e:
            raise RuntimeError(f"Failed to create KV bucket '{bucket_name}': {e}")
    except Exception as e:
        raise RuntimeError(f"Failed to get KV bucket '{bucket_name}': {e}")
```

**Step 2: Update connect() to initialize JetStream**

Modify connect() method to include:

```python
# Initialize JetStream
try:
    self._js = self.nc.jetstream()
except Exception as e:
    raise RuntimeError(f"Failed to initialize JetStream: {e}")
```

**Step 3: Add _js field to __init__**

```python
def __init__(self, nats_url: str):
    self.nc = None
    self._js = None  # JetStream context
    self._subs = []
    self._url = nats_url
```

**Step 4: Commit**

```bash
git add sdk/python/lightlink/client.py
git commit -m "feat(python-sdk): add JetStream KV helper methods"
```

---

## Task 3: Implement SetState Method

**Files:**
- Modify: `sdk/python/lightlink/client.py`

**Step 1: Add set_state method**

```python
async def set_state(self, key: str, value: dict) -> None:
    """
    Set a state value in the KV store.

    Args:
        key: State key (e.g., "config", "user:123")
        value: State value as a dictionary

    Raises:
        RuntimeError: If not connected or operation fails
    """
    if self.nc is None:
        raise RuntimeError("Not connected. Call connect() first.")

    kv = await self._get_or_create_kv("light_link_state")

    # Serialize value to JSON
    import json
    data = json.dumps(value).encode('utf-8')

    # Store in KV
    try:
        await kv.put(key, data)
    except Exception as e:
        raise RuntimeError(f"Failed to set state '{key}': {e}")
```

**Step 2: Write unit test**

Create `sdk/python/tests/test_state.py`:

```python
import pytest
import asyncio
from lightlink.client import Client

@pytest.mark.asyncio
async def test_set_state():
    """Test setting state values"""
    client = Client("nats://localhost:4222")

    try:
        await client.connect()

        # Set a simple state
        await client.set_state("test_key", {"value": 42})

        # If no exception, test passes
        assert True
    except Exception as e:
        pytest.fail(f"set_state failed: {e}")
    finally:
        await client.close()

@pytest.mark.asyncio
async def test_set_state_complex():
    """Test setting complex state values"""
    client = Client("nats://localhost:4222")

    try:
        await client.connect()

        # Set complex state
        state = {
            "config": {
                "debug": True,
                "version": "1.0.0"
            },
            "users": ["alice", "bob"],
            "count": 100
        }
        await client.set_state("complex", state)

        assert True
    except Exception as e:
        pytest.fail(f"set_state complex failed: {e}")
    finally:
        await client.close()
```

**Step 3: Run test**

Run: `cd sdk/python && python -m pytest tests/test_state.py::test_set_state -v`
Expected: May fail if NATS not running, but code compiles

**Step 4: Commit**

```bash
git add sdk/python/lightlink/client.py sdk/python/tests/test_state.py
git commit -m "feat(python-sdk): add set_state method"
```

---

## Task 4: Implement GetState Method

**Files:**
- Modify: `sdk/python/lightlink/client.py`

**Step 1: Add get_state method**

```python
async def get_state(self, key: str) -> dict:
    """
    Get a state value from the KV store.

    Args:
        key: State key to retrieve

    Returns:
        State value as a dictionary

    Raises:
        RuntimeError: If not connected or key not found
    """
    if self.nc is None:
        raise RuntimeError("Not connected. Call connect() first.")

    kv = await self._get_or_create_kv("light_link_state")

    # Get from KV
    try:
        entry = await kv.get(key)
    except nats.errors.NotFoundError:
        raise KeyError(f"State key '{key}' not found")
    except Exception as e:
        raise RuntimeError(f"Failed to get state '{key}': {e}")

    # Deserialize value from JSON
    import json
    try:
        value = json.loads(entry.value.decode('utf-8'))
        return value
    except json.JSONDecodeError as e:
        raise RuntimeError(f"Failed to deserialize state '{key}': {e}")
```

**Step 2: Add unit test**

Add to `sdk/python/tests/test_state.py`:

```python
@pytest.mark.asyncio
async def test_get_state():
    """Test getting state values"""
    client = Client("nats://localhost:4222")

    try:
        await client.connect()

        # First set a value
        await client.set_state("test_get", {"name": "test", "value": 123})

        # Then get it back
        state = await client.get_state("test_get")

        assert state["name"] == "test"
        assert state["value"] == 123
    except Exception as e:
        pytest.fail(f"get_state failed: {e}")
    finally:
        await client.close()

@pytest.mark.asyncio
async def test_get_state_not_found():
    """Test getting non-existent state"""
    client = Client("nats://localhost:4222")

    try:
        await client.connect()

        # Try to get non-existent key
        with pytest.raises(KeyError):
            await client.get_state("nonexistent_key")
    finally:
        await client.close()
```

**Step 3: Run tests**

Run: `cd sdk/python && python -m pytest tests/test_state.py::test_get_state -v`
Expected: Tests pass (with NATS running)

**Step 4: Commit**

```bash
git add sdk/python/lightlink/client.py sdk/python/tests/test_state.py
git commit -m "feat(python-sdk): add get_state method"
```

---

## Task 5: Implement WatchState Method

**Files:**
- Modify: `sdk/python/lightlink/client.py`

**Step 1: Add watch_state method**

```python
async def watch_state(self, key: str, handler):
    """
    Watch for changes to a state key.

    Args:
        key: State key to watch
        handler: Async callback function that receives the updated state

    Returns:
        Watcher object that can be used to stop watching

    Raises:
        RuntimeError: If not connected or watch fails
    """
    if self.nc is None:
        raise RuntimeError("Not connected. Call connect() first.")

    kv = await self._get_or_create_kv("light_link_state")

    # Create a watch iterator
    try:
        watcher = await kv.watch(key)
    except Exception as e:
        raise RuntimeError(f"Failed to watch state '{key}': {e}")

    # Start watching in background
    async def watch_loop():
        try:
            async for entry in watcher:
                # Deserialize value
                import json
                try:
                    value = json.loads(entry.value.decode('utf-8'))
                    await handler(value)
                except json.JSONDecodeError:
                    # Pass raw value if JSON decode fails
                    await handler(entry.value.decode('utf-8'))
        except Exception as e:
            # Watch ended (e.g., key deleted)
            pass

    # Start watch loop as background task
    import asyncio
    task = asyncio.create_task(watch_loop())

    # Return function to stop watching
    def stop():
        task.cancel()

    return stop
```

**Step 2: Add unit test**

Add to `sdk/python/tests/test_state.py`:

```python
@pytest.mark.asyncio
async def test_watch_state():
    """Test watching state changes"""
    client = Client("nats://localhost:4222")

    try:
        await client.connect()

        # Track received updates
        received = []

        async def handler(value):
            received.append(value)

        # Start watching
        stop = await client.watch_state("test_watch", handler)

        # Give watch time to establish
        await asyncio.sleep(0.5)

        # Make some updates
        await client.set_state("test_watch", {"version": 1})
        await asyncio.sleep(0.2)

        await client.set_state("test_watch", {"version": 2})
        await asyncio.sleep(0.2)

        # Stop watching
        stop()

        # Verify we received updates
        assert len(received) >= 2
        assert received[-1]["version"] == 2
    except Exception as e:
        pytest.fail(f"watch_state failed: {e}")
    finally:
        await client.close()
```

**Step 3: Run tests**

Run: `cd sdk/python && python -m pytest tests/test_state.py::test_watch_state -v`
Expected: Tests pass (with NATS running)

**Step 4: Commit**

```bash
git add sdk/python/lightlink/client.py sdk/python/tests/test_state.py
git commit -m "feat(python-sdk): add watch_state method"
```

---

## Task 6: Implement DeleteState Method

**Files:**
- Modify: `sdk/python/lightlink/client.py`

**Step 1: Add delete_state method**

```python
async def delete_state(self, key: str) -> bool:
    """
    Delete a state key from the KV store.

    Args:
        key: State key to delete

    Returns:
        True if deleted, False if key didn't exist

    Raises:
        RuntimeError: If not connected or operation fails
    """
    if self.nc is None:
        raise RuntimeError("Not connected. Call connect() first.")

    kv = await self._get_or_create_kv("light_link_state")

    try:
        await kv.delete(key)
        return True
    except nats.errors.NotFoundError:
        return False
    except Exception as e:
        raise RuntimeError(f"Failed to delete state '{key}': {e}")
```

**Step 2: Add unit test**

Add to `sdk/python/tests/test_state.py`:

```python
@pytest.mark.asyncio
async def test_delete_state():
    """Test deleting state values"""
    client = Client("nats://localhost:4222")

    try:
        await client.connect()

        # Set a value
        await client.set_state("test_delete", {"temp": True})

        # Delete it
        result = await client.delete_state("test_delete")
        assert result is True

        # Verify it's gone
        with pytest.raises(KeyError):
            await client.get_state("test_delete")

        # Delete non-existent key
        result = await client.delete_state("nonexistent")
        assert result is False
    except Exception as e:
        pytest.fail(f"delete_state failed: {e}")
    finally:
        await client.close()
```

**Step 3: Run tests**

Run: `cd sdk/python && python -m pytest tests/test_state.py::test_delete_state -v`
Expected: Tests pass (with NATS running)

**Step 4: Commit**

```bash
git add sdk/python/lightlink/client.py sdk/python/tests/test_state.py
git commit -m "feat(python-sdk): add delete_state method"
```

---

## Task 7: Update SDK Documentation

**Files:**
- Modify: `sdk/python/README.md`

**Step 1: Add state management section**

```markdown
## State Management

The Python SDK supports state management using NATS JetStream KV store.

### Set State

```python
await client.set_state("config", {
    "version": "1.0.0",
    "debug": True
})
```

### Get State

```python
config = await client.get_state("config")
print(config["version"])  # "1.0.0"
```

### Watch State Changes

```python
async def on_state_change(new_state):
    print(f"State changed: {new_state}")

stop_watching = await client.watch_state("config", on_state_change)

# Later: stop watching
stop_watching()
```

### Delete State

```python
await client.delete_state("config")
```

## API Reference

### Client Methods

- `connect()` - Connect to NATS server
- `close()` - Close connection
- `call(service, method, args)` - RPC call
- `publish(subject, data)` - Publish message
- `subscribe(subject, handler)` - Subscribe to messages
- `set_state(key, value)` - Set state value
- `get_state(key)` - Get state value
- `watch_state(key, handler)` - Watch for state changes
- `delete_state(key)` - Delete state value
```

**Step 2: Commit**

```bash
git add sdk/python/README.md
git commit -m "docs(python-sdk): add state management documentation"
```

---

## Task 8: Update Existing Examples to Use State Management

**Files:**
- Modify: `light_link_platform/examples/notify/python/state_demo/main.py`

**Step 1: Update state_demo to use implemented methods**

Remove the "not implemented" checks since methods now exist.

**Step 2: Test the updated example**

Run: `python main.py`
Expected: Full state management demo works

**Step 3: Commit**

```bash
git add light_link_platform/examples/notify/python/state_demo/main.py
git commit -m "fix(python): update state_demo to use new SDK methods"
```

---

## Testing Strategy

### Prerequisites
- NATS server with JetStream enabled
- TLS certificates

### Run All Tests

```bash
cd sdk/python
python -m pytest tests/test_state.py -v
```

### Manual Testing

```python
import asyncio
from lightlink.client import Client, discover_client_certs

async def test():
    certs = discover_client_certs()
    client = Client("nats://localhost:4222")
    await client.connect(
        tls_cert_file=certs.cert_file,
        tls_key_file=certs.key_file,
        tls_ca_file=certs.ca_file
    )

    # Set, get, watch, delete
    await client.set_state("test", {"value": 42})
    state = await client.get_state("test")
    print(state)  # {'value': 42}
    await client.delete_state("test")
    await client.close()

asyncio.run(test())
```

---

## Dependencies

- Python 3.8+
- nats-py with JetStream support
- NATS server with JetStream enabled

---

## Related Plans

- P1 State KV Examples: `2024-12-26-p1-state-kv.md` (Uses this SDK implementation)

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
2. Navigate to the State/KV section
3. Verify Python SDK state operations:
   - `set_state()` creates entries visible in UI
   - `get_state()` retrieves correct values
   - `watch_state()` triggers UI updates

### Step 4: Test Python SDK Integration

1. Run Python state management example
2. Perform state operations via example
3. Verify in management platform:
   - State changes appear immediately
   - KV bucket `light_link_state` contains data
   - Watch handlers receive updates

### Step 5: Capture Evidence

Take screenshots showing Python SDK state operations reflected in management platform.
