# SDK API Comparison

This document documents the differences between SDK implementations.

## Connection

| Language | Method | Signature |
|----------|--------|-----------|
| Go | `NewClient` | `func NewClient(url string, tlsConfig *TLSConfig) (*Client, error)` |
| Python | `Client.connect()` | `async def connect(self)` |
| C# | `Client.Connect()` | `async Task Connect()` |

## RPC Call

| Language | Method | Signature |
|----------|--------|-----------|
| Go | `Call` | `func (c *Client) Call(service, method string, args map[string]interface{}) (map[string]interface{}, error)` |
| Python | `call` | `async def call(self, service, method, args, timeout=5.0)` |
| C# | `Call` | `async Task<Dictionary<string, JsonElement>> Call(string service, string method, Dictionary<string, object> args, int timeoutMs = 5000)` |

**Note:** C# uses `Dictionary<string, JsonElement>` while others use native map/dict types. This is a known limitation for cross-language compatibility.

## Publish/Subscribe

| Language | Publish Method | Subscribe Method |
|----------|----------------|------------------|
| Go | `Publish(subject, data)` | `Subscribe(subject, handler)` |
| Python | `publish(subject, data)` | `subscribe(subject, handler)` |
| C# | `Publish(subject, data)` | `Subscribe(subject, handler)` |

## State Management

| Language | Set State | Get State | Watch State |
|----------|-----------|-----------|-------------|
| Go | `SetState(key, value)` | `GetState(key)` | `WatchState(key, handler)` |
| Python | `set_state(key, value)` | `get_state(key)` | `watch_state(key, handler)` |
| C# | `SetState(key, value)` | `GetState(key)` | `WatchState(key, handler)` |

## File Transfer

| Language | Upload | Download |
|----------|--------|----------|
| Go | `UploadFile(path, name)` | `DownloadFile(id, path)` |
| Python | `upload_file(path, name)` | `download_file(id, path)` |
| C# | `UploadFile(path, name)` | `DownloadFile(id, path)` |

**Note:** File transfer implementations differ internally:
- Go: Uses NATS Object Store API
- Python/C#: Manual chunk-based implementation
