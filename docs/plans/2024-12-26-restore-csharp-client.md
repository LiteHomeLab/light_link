# Restore C# Client.cs Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Restore the C# SDK Client.cs file to enable C# services to act as callers, use file transfer, and state management features.

**Architecture:**
- Client.cs provides the client-side counterpart to Service.cs
- Uses NATS.Client 1.1.8 for NATS connection and JetStream operations
- Integrates with existing TLSConfig.cs for automatic certificate discovery
- Follows the same patterns as Go SDK (connection, RPC, pub/sub, state, file transfer)

**Tech Stack:**
- C# .NET 8.0
- NATS.Client 1.1.8
- System.Text.Json 8.0.5
- xUnit for testing

---

## Task 1: Create Unit Test Project Structure

**Files:**
- Create: `sdk/csharp/LightLink.Tests/ClientTests.cs`
- Create: `sdk/csharp/LightLink.Tests/LightLink.Tests.csproj` (if not exists)

**Step 1: Verify test project exists**

Run: `ls sdk/csharp/LightLink.Tests/`
Expected: Existing test project with ServiceTests.cs

**Step 2: Create ClientTests.cs with empty test class**

```csharp
using Xunit;
using System;
using System.Collections.Generic;
using System.Threading.Tasks;

namespace LightLink.Tests
{
    public class ClientTests : IDisposable
    {
        // Tests will be added in subsequent tasks
        public void Dispose()
        {
            // Cleanup code
        }
    }
}
```

**Step 3: Build test project to verify compilation**

Run: `cd sdk/csharp/LightLink.Tests && dotnet build`
Expected: BUILD SUCCESS

**Step 4: Commit**

```bash
git add sdk/csharp/LightLink.Tests/ClientTests.cs
git commit -m "test(csharp): add empty ClientTests class"
```

---

## Task 2: Write Failing Test for Client Connection

**Files:**
- Modify: `sdk/csharp/LightLink.Tests/ClientTests.cs`

**Step 1: Write failing test for client creation and connection**

```csharp
[Fact]
public void Client_Connect_WithoutTLS_ShouldConnect()
{
    // Arrange
    var client = new Client("nats://localhost:4222");

    // Act
    client.Connect();

    // Assert
    Assert.True(client.IsConnected);

    // Cleanup
    client.Close();
}

[Fact]
public async Task Client_ConnectAsync_WithoutTLS_ShouldConnect()
{
    // Arrange
    var client = new Client("nats://localhost:4222");

    // Act
    await client.ConnectAsync();

    // Assert
    Assert.True(client.IsConnected);

    // Cleanup
    client.Close();
}
```

**Step 2: Run test to verify it fails**

Run: `cd sdk/csharp/LightLink.Tests && dotnet test ClientTests.cs -f net8.0`
Expected: FAIL with "Client does not exist"

**Step 3: Commit**

```bash
git add sdk/csharp/LightLink.Tests/ClientTests.cs
git commit -m "test(csharp): add failing connection tests"
```

---

## Task 3: Create Client.cs with Basic Structure

**Files:**
- Create: `sdk/csharp/LightLink/Client.cs`

**Step 1: Create minimal Client class**

```csharp
using System;
using NATS.Client;

namespace LightLink
{
    /// <summary>
    /// LightLink C# Client
    /// Provides RPC, Pub/Sub, State Management, and File Transfer capabilities
    /// </summary>
    public class Client : IDisposable
    {
        private string _url;
        private TLSConfig? _tlsConfig;
        private IConnection? _nc;

        /// <summary>
        /// Create a new client
        /// </summary>
        /// <param name="url">NATS server URL (default: nats://localhost:4222)</param>
        /// <param name="tlsConfig">Optional TLS configuration</param>
        public Client(string url = "nats://localhost:4222", TLSConfig? tlsConfig = null)
        {
            _url = url;
            _tlsConfig = tlsConfig;
        }

        /// <summary>
        /// Connect to NATS server
        /// </summary>
        public void Connect()
        {
            var opts = ConnectionFactory.GetDefaultOptions();
            opts.Url = _url;
            opts.Name = "LightLink C# Client";
            opts.MaxReconnect = 10;
            opts.ReconnectWait = 2000;

            // Configure TLS if provided
            if (_tlsConfig != null)
            {
                ConfigureTLS(opts);
            }

            _nc = new ConnectionFactory().CreateConnection(opts);
        }

        /// <summary>
        /// Connect asynchronously
        /// </summary>
        public System.Threading.Tasks.Task ConnectAsync()
        {
            return System.Threading.Tasks.Task.Run(() => Connect());
        }

        /// <summary>
        /// Close connection
        /// </summary>
        public void Close()
        {
            _nc?.Close();
            _nc = null;
        }

        /// <summary>
        /// Check if connected
        /// </summary>
        public bool IsConnected => _nc != null && _nc.State == ConnState.CONNECTED;

        /// <summary>
        /// Dispose
        /// </summary>
        public void Dispose()
        {
            Close();
        }

        private void ConfigureTLS(Options opts)
        {
            if (_tlsConfig == null) return;

            // Use PFX certificate if available
            if (!string.IsNullOrEmpty(_tlsConfig.PfxFile) &&
                System.IO.File.Exists(_tlsConfig.PfxFile))
            {
                var cert = new System.Security.Cryptography.X509Certificates.X509Certificate2(
                    _tlsConfig.PfxFile,
                    _tlsConfig.PfxPassword);

                opts.AddCertificate(cert);
            }
            // Fall back to cert/key files
            else if (!string.IsNullOrEmpty(_tlsConfig.CertFile) &&
                     System.IO.File.Exists(_tlsConfig.CertFile))
            {
                var cert = new System.Security.Cryptography.X509Certificates.X509Certificate2(
                    _tlsConfig.CertFile);

                opts.AddCertificate(cert);
            }

            // Configure SSL/TLS
            opts.Secure = true;

            // Skip server name verification for self-signed certs
            if (_tlsConfig.InsecureSkipVerify)
            {
                opts.TLSRemoteCertificationValidationCallback =
                    (sender, certificate, chain, sslPolicyErrors) => true;
            }
        }
    }
}
```

**Step 2: Build to verify compilation**

Run: `cd sdk/csharp/LightLink && dotnet build`
Expected: BUILD SUCCESS

**Step 3: Run test to verify it passes**

Run: `cd sdk/csharp/LightLink.Tests && dotnet test ClientTests.cs -f net8.0`
Expected: PASS (requires NATS server running or skip test)

**Step 4: Commit**

```bash
git add sdk/csharp/LightLink/Client.cs
git commit -m "feat(csharp): add Client class with basic connection"
```

---

## Task 4: Write Failing Test for RPC Call

**Files:**
- Modify: `sdk/csharp/LightLink.Tests/ClientTests.cs`

**Step 1: Write failing test for RPC call**

```csharp
[Fact]
public async Task Client_Call_ShouldInvokeRemoteMethod()
{
    // Arrange
    var client = new Client("nats://localhost:4222");
    await client.ConnectAsync();

    // Act
    var result = client.Call("math-service", "add",
        new Dictionary<string, object>
        {
            { "a", 10 },
            { "b", 20 }
        });

    // Assert
    Assert.NotNull(result);
    Assert.True(result.ContainsKey("result"));
    Assert.Equal(30, (int)result["result"].GetInt32());

    // Cleanup
    client.Close();
}
```

**Step 2: Run test to verify it fails**

Run: `cd sdk/csharp/LightLink.Tests && dotnet test -f net8.0`
Expected: FAIL with "Call does not exist"

**Step 3: Commit**

```bash
git add sdk/csharp/LightLink.Tests/ClientTests.cs
git commit -m "test(csharp): add failing RPC call test"
```

---

## Task 5: Implement RPC Call Method

**Files:**
- Modify: `sdk/csharp/LightLink/Client.cs`

**Step 1: Add RPC request/response classes and Call method**

Add to Client.cs (before Dispose):

```csharp
/// <summary>
/// RPC call (synchronous)
/// </summary>
/// <param name="service">Service name</param>
/// <param name="method">Method name</param>
/// <param name="args">Arguments dictionary</param>
/// <param name="timeoutMs">Timeout in milliseconds (default: 5000)</param>
/// <returns>Result dictionary</returns>
public Dictionary<string, JsonElement> Call(string service, string method,
    Dictionary<string, object> args, int timeoutMs = 5000)
{
    if (_nc == null)
        throw new InvalidOperationException("Not connected. Call Connect() first.");

    string subject = $"$SRV.{service}.{method}";

    var request = new RPCRequest
    {
        Id = Guid.NewGuid().ToString(),
        Method = method,
        Args = args
    };

    string requestJson = JsonSerializer.Serialize(request);
    byte[] requestData = System.Text.Encoding.UTF8.GetBytes(requestJson);

    try
    {
        Msg msg = _nc.Request(subject, requestData, timeoutMs);
        string responseJson = System.Text.Encoding.UTF8.GetString(msg.Data);

        var response = JsonSerializer.Deserialize<RPCResponse>(responseJson);
        if (response == null || !response.Success)
        {
            throw new Exception(response?.Error ?? "RPC call failed");
        }

        return response.Result ?? new Dictionary<string, JsonElement>();
    }
    catch (NATS.Client.NATSTimeoutException)
    {
        throw new TimeoutException($"RPC call to {service}.{method} timed out");
    }
}

/// <summary>
/// RPC call (asynchronous)
/// </summary>
public async Task<Dictionary<string, JsonElement>> CallAsync(string service, string method,
    Dictionary<string, object> args, int timeoutMs = 5000)
{
    return await Task.Run(() => Call(service, method, args, timeoutMs));
}
```

**Step 2: Add internal RPC classes at end of file**

```csharp
// Internal classes for JSON serialization

internal class RPCRequest
{
    public string Id { get; set; } = string.Empty;
    public string Method { get; set; } = string.Empty;
    public Dictionary<string, object> Args { get; set; } = new();
}

internal class RPCResponse
{
    public string Id { get; set; } = string.Empty;
    public bool Success { get; set; }
    public Dictionary<string, JsonElement>? Result { get; set; }
    public string Error { get; set; } = string.Empty;
}
```

**Step 3: Build to verify compilation**

Run: `cd sdk/csharp/LightLink && dotnet build`
Expected: BUILD SUCCESS

**Step 4: Run tests**

Run: `cd sdk/csharp/LightLink.Tests && dotnet test -f net8.0`
Expected: Tests may fail if math-service not running, but compilation passes

**Step 5: Commit**

```bash
git add sdk/csharp/LightLink/Client.cs
git commit -m "feat(csharp): add RPC call methods"
```

---

## Task 6: Write Failing Test for Publish/Subscribe

**Files:**
- Modify: `sdk/csharp/LightLink.Tests/ClientTests.cs`

**Step 1: Write failing test for publish/subscribe**

```csharp
[Fact]
public async Task Client_PublishSubscribe_ShouldReceiveMessage()
{
    // Arrange
    var client = new Client("nats://localhost:4222");
    await client.ConnectAsync();

    var subject = "test.pubsub";
    var received = false;
    var receivedData = new Dictionary<string, JsonElement>();

    // Subscribe first
    using (var sub = client.Subscribe(subject, (data) =>
    {
        received = true;
        receivedData = data;
    }))
    {
        // Wait for subscription to be ready
        await Task.Delay(100);

        // Act
        client.Publish(subject, new Dictionary<string, object>
        {
            { "message", "Hello, LightLink!" },
            { "count", 42 }
        });

        // Wait for message
        await Task.Delay(200);

        // Assert
        Assert.True(received);
        Assert.True(receivedData.ContainsKey("message"));
        Assert.Equal("Hello, LightLink!", receivedData["message"].GetString());
        Assert.Equal(42, receivedData["count"].GetInt32());
    }

    // Cleanup
    client.Close();
}
```

**Step 2: Run test to verify it fails**

Run: `cd sdk/csharp/LightLink.Tests && dotnet test -f net8.0`
Expected: FAIL with "Publish/Subscribe does not exist"

**Step 3: Commit**

```bash
git add sdk/csharp/LightLink.Tests/ClientTests.cs
git commit -m "test(csharp): add failing pub/sub tests"
```

---

## Task 7: Implement Publish/Subscribe Methods

**Files:**
- Modify: `sdk/csharp/LightLink/Client.cs`

**Step 1: Add Publish and Subscribe methods**

Add to Client.cs:

```csharp
/// <summary>
/// Publish message
/// </summary>
public void Publish(string subject, Dictionary<string, object> data)
{
    if (_nc == null)
        throw new InvalidOperationException("Not connected. Call Connect() first.");

    string json = JsonSerializer.Serialize(data);
    byte[] msgData = System.Text.Encoding.UTF8.GetBytes(json);
    _nc.Publish(subject, msgData);
}

/// <summary>
/// Publish message asynchronously
/// </summary>
public async Task PublishAsync(string subject, Dictionary<string, object> data)
{
    await Task.Run(() => Publish(subject, data));
}

/// <summary>
/// Subscribe to messages
/// </summary>
public ISubscription Subscribe(string subject, Action<Dictionary<string, JsonElement>> handler)
{
    if (_nc == null)
        throw new InvalidOperationException("Not connected. Call Connect() first.");

    return _nc.SubscribeAsync(subject, (msg) =>
    {
        try
        {
            string json = System.Text.Encoding.UTF8.GetString(msg.Data);
            var data = JsonSerializer.Deserialize<Dictionary<string, JsonElement>>(json);
            if (data != null)
            {
                handler(data);
            }
        }
        catch (Exception)
        {
            // Ignore JSON deserialization errors
        }
    });
}
```

**Step 2: Build and test**

Run: `cd sdk/csharp/LightLink && dotnet build && cd ../LightLink.Tests && dotnet test -f net8.0`
Expected: Tests pass

**Step 3: Commit**

```bash
git add sdk/csharp/LightLink/Client.cs
git commit -m "feat(csharp): add publish/subscribe methods"
```

---

## Task 8: Write Failing Test for State Management (KV)

**Files:**
- Modify: `sdk/csharp/LightLink.Tests/ClientTests.cs`

**Step 1: Write failing test for state management**

```csharp
[Fact]
public async Task Client_SetGetState_ShouldStoreAndRetrieve()
{
    // Arrange
    var client = new Client("nats://localhost:4222");
    await client.ConnectAsync();

    var key = "test.state";
    var value = new Dictionary<string, object>
    {
        { "status", "active" },
        { "count", 100 },
        { "enabled", true }
    };

    // Act - Set state
    client.SetState(key, value);
    await Task.Delay(100); // Allow time for KV update

    // Act - Get state
    var retrieved = client.GetState(key);

    // Assert
    Assert.NotNull(retrieved);
    Assert.True(retrieved.ContainsKey("status"));
    Assert.Equal("active", retrieved["status"].GetString());
    Assert.True(retrieved.ContainsKey("count"));
    Assert.Equal(100, retrieved["count"].GetInt32());
    Assert.True(retrieved.ContainsKey("enabled"));
    Assert.True(retrieved["enabled"].GetBoolean());

    // Cleanup
    client.Close();
}
```

**Step 2: Run test to verify it fails**

Run: `cd sdk/csharp/LightLink.Tests && dotnet test -f net8.0`
Expected: FAIL with "SetState/GetState does not exist"

**Step 3: Commit**

```bash
git add sdk/csharp/LightLink.Tests/ClientTests.cs
git commit -m "test(csharp): add failing state management tests"
```

---

## Task 9: Implement State Management Methods

**Files:**
- Modify: `sdk/csharp/LightLink/Client.cs`

**Step 1: Add JetStream context field and helper methods**

Add to Client class fields:

```csharp
private NATS.Client.JetStream.IJetStream? _js;
```

Add to Connect() method after creating connection:

```csharp
_nc = new ConnectionFactory().CreateConnection(opts);
_js = _nc.CreateJetStreamContext();
```

Add before Dispose:

```csharp
private NATS.Client.JetStream.IKeyValue GetOrCreateKeyValue(string bucketName)
{
    if (_js == null)
        throw new InvalidOperationException("JetStream not initialized");

    try
    {
        return _js.GetKeyValue(bucketName);
    }
    catch
    {
        var config = NATS.Client.JetStream.KeyValueConfiguration.Builder()
            .WithName(bucketName)
            .Build();
        return _js.CreateKeyValue(config);
    }
}
```

**Step 2: Add SetState and GetState methods**

```csharp
/// <summary>
/// Set state value
/// </summary>
public void SetState(string key, Dictionary<string, object> value)
{
    var kv = GetOrCreateKeyValue("light_link_state");
    string json = JsonSerializer.Serialize(value);
    byte[] data = System.Text.Encoding.UTF8.GetBytes(json);
    kv.Put(key, data);
}

/// <summary>
/// Get state value
/// </summary>
public Dictionary<string, JsonElement> GetState(string key)
{
    var kv = GetOrCreateKeyValue("light_link_state");
    var entry = kv.Get(key);
    string json = System.Text.Encoding.UTF8.GetString(entry.Value);
    var result = JsonSerializer.Deserialize<Dictionary<string, JsonElement>>(json);
    return result ?? new Dictionary<string, JsonElement>();
}
```

**Step 3: Build and test**

Run: `cd sdk/csharp/LightLink && dotnet build && cd ../LightLink.Tests && dotnet test -f net8.0`
Expected: Tests pass

**Step 4: Commit**

```bash
git add sdk/csharp/LightLink/Client.cs
git commit -m "feat(csharp): add state management (KV) methods"
```

---

## Task 10: Write Failing Test for File Transfer

**Files:**
- Modify: `sdk/csharp/LightLink.Tests/ClientTests.cs`

**Step 1: Write failing test for file upload/download**

```csharp
[Fact]
public async Task Client_UploadDownloadFile_ShouldTransferFile()
{
    // Arrange
    var client = new Client("nats://localhost:4222");
    await client.ConnectAsync();

    // Create test file
    var testFile = "test_upload.txt";
    var downloadFile = "test_download.txt";
    var testContent = "Hello, LightLink File Transfer!";
    await System.IO.File.WriteAllTextAsync(testFile, testContent);

    try
    {
        // Act - Upload
        var fileId = client.UploadFile(testFile, "remote.txt");
        Assert.NotNull(fileId);
        Assert.NotEmpty(fileId);

        await Task.Delay(100); // Allow time for upload

        // Act - Download
        client.DownloadFile(fileId, downloadFile);

        await Task.Delay(100); // Allow time for download

        // Assert
        var downloadedContent = await System.IO.File.ReadAllTextAsync(downloadFile);
        Assert.Equal(testContent, downloadedContent);
    }
    finally
    {
        // Cleanup
        if (System.IO.File.Exists(testFile)) System.IO.File.Delete(testFile);
        if (System.IO.File.Exists(downloadFile)) System.IO.File.Delete(downloadFile);
        client.Close();
    }
}
```

**Step 2: Run test to verify it fails**

Run: `cd sdk/csharp/LightLink.Tests && dotnet test -f net8.0`
Expected: FAIL with "UploadFile/DownloadFile does not exist"

**Step 3: Commit**

```bash
git add sdk/csharp/LightLink.Tests/ClientTests.cs
git commit -m "test(csharp): add failing file transfer tests"
```

---

## Task 11: Implement File Transfer Methods

**Files:**
- Modify: `sdk/csharp/LightLink/Client.cs`

**Step 1: Add ObjectStore helper method**

Add before Dispose:

```csharp
private NATS.Client.JetStream.IObjectStore GetOrCreateObjectStore(string bucketName)
{
    if (_js == null)
        throw new InvalidOperationException("JetStream not initialized");

    try
    {
        return _js.GetObjectStore(bucketName);
    }
    catch
    {
        var config = NATS.Client.JetStream.ObjectStoreConfiguration.Builder()
            .WithName(bucketName)
            .Build();
        return _js.CreateObjectStore(config);
    }
}
```

**Step 2: Add UploadFile and DownloadFile methods**

```csharp
/// <summary>
/// Upload file to Object Store
/// </summary>
public string UploadFile(string filePath, string remoteName)
{
    var objStore = GetOrCreateObjectStore("light_link_files");
    string fileId = Guid.NewGuid().ToString();

    byte[] fileData = System.IO.File.ReadAllBytes(filePath);

    // Upload entire file as one object (like Go SDK)
    objStore.Put(fileId, fileData);

    // Publish metadata notification
    var metadata = new Dictionary<string, object>
    {
        { "file_id", fileId },
        { "file_name", remoteName },
        { "file_size", fileData.Length }
    };

    try
    {
        Publish("file.uploaded", metadata);
    }
    catch
    {
        // Ignore publish errors
    }

    return fileId;
}

/// <summary>
/// Upload file asynchronously
/// </summary>
public async Task<string> UploadFileAsync(string filePath, string remoteName)
{
    return await Task.Run(() => UploadFile(filePath, remoteName));
}

/// <summary>
/// Download file from Object Store
/// </summary>
public void DownloadFile(string fileId, string localPath)
{
    var objStore = GetOrCreateObjectStore("light_link_files");
    var fileData = objStore.GetBytes(fileId);

    System.IO.File.WriteAllBytes(localPath, fileData);
}

/// <summary>
/// Download file asynchronously
/// </summary>
public async Task DownloadFileAsync(string fileId, string localPath)
{
    await Task.Run(() => DownloadFile(fileId, localPath));
}
```

**Step 3: Build and test**

Run: `cd sdk/csharp/LightLink && dotnet build && cd ../LightLink.Tests && dotnet test -f net8.0`
Expected: Tests pass

**Step 4: Commit**

```bash
git add sdk/csharp/LightLink/Client.cs
git commit -m "feat(csharp): add file transfer methods"
```

---

## Task 12: Clean Up - Remove .bak File

**Files:**
- Delete: `sdk/csharp/LightLink/Client.cs.bak`

**Step 1: Remove backup file**

Run: `rm sdk/csharp/LightLink/Client.cs.bak`

**Step 2: Verify build still works**

Run: `cd sdk/csharp/LightLink && dotnet build`
Expected: BUILD SUCCESS

**Step 3: Run all tests**

Run: `cd sdk/csharp/LightLink.Tests && dotnet test -f net8.0`
Expected: All tests pass

**Step 4: Commit**

```bash
git add sdk/csharp/LightLink/Client.cs.bak
git commit -m "chore(csharp): remove Client.cs.bak backup file"
```

---

## Task 13: Update Project Documentation

**Files:**
- Modify: `sdk/csharp/README.md` (create if not exists)

**Step 1: Create SDK README**

Create `sdk/csharp/README.md`:

```markdown
# LightLink C# SDK

LightLink C# SDK provides client and server functionality for NATS-based microservices communication.

## Features

- **RPC Client/Service**: Remote procedure calls with timeout support
- **Pub/Sub**: Message publishing and subscribing
- **State Management**: KV-based state storage
- **File Transfer**: Upload/download files via Object Store
- **TLS Support**: Automatic certificate discovery and configuration

## Installation

```bash
dotnet add package LightLink
```

## Quick Start

### RPC Client

```csharp
using LightLink;

// Create client
var client = new Client("nats://localhost:4222");
client.Connect();

// Call remote service
var result = client.Call("math-service", "add",
    new Dictionary<string, object> { { "a", 10 }, { "b", 20 } });

Console.WriteLine($"Result: {result["result"]}");

client.Close();
```

### Service

```csharp
using LightLink;

// Create service
var svc = new Service("my-service", "nats://localhost:4222");

// Register method
svc.RegisterMethodWithMetadata("add",
    (args) =>
    {
        int a = args.ContainsKey("a") ? int.Parse(args["a"].ToString()) : 0;
        int b = args.ContainsKey("b") ? int.Parse(args["b"].ToString()) : 0;
        return Task.FromResult(
            new Dictionary<string, object> { { "result", a + b } });
    },
    new MethodMetadata
    {
        Description = "Add two numbers",
        Parameters = new Dictionary<string, ParameterInfo>
        {
            { "a", new ParameterInfo { Type = "int", Description = "First number" } },
            { "b", new ParameterInfo { Type = "int", Description = "Second number" } }
        }
    });

svc.Start();
```

### Publish/Subscribe

```csharp
// Subscribe
client.Subscribe("my.subject", (data) =>
{
    Console.WriteLine($"Received: {data["message"]}");
});

// Publish
client.Publish("my.subject",
    new Dictionary<string, object> { { "message", "Hello!" } });
```

### State Management

```csharp
// Set state
client.SetState("config",
    new Dictionary<string, object>
    {
        { "enabled", true },
        { "threshold", 100 }
    });

// Get state
var config = client.GetState("config");
```

### File Transfer

```csharp
// Upload
string fileId = client.UploadFile("local.txt", "remote.txt");

// Download
client.DownloadFile(fileId, "downloaded.txt");
```

## TLS Configuration

The SDK automatically discovers certificates in the following locations:

- `./client`
- `../client`
- `../../client`

Or manually configure:

```csharp
var tlsConfig = new TLSConfig
{
    PfxFile = "./client/client.pfx",
    PfxPassword = "lightlink",
    InsecureSkipVerify = true
};

var client = new Client("nats://localhost:4222", tlsConfig);
```

## Requirements

- .NET 8.0 or higher
- NATS.Client 1.1.8
- NATS server with JetStream enabled

## License

MIT License
```

**Step 2: Commit**

```bash
git add sdk/csharp/README.md
git commit -m "docs(csharp): add SDK documentation"
```

---

## Testing Strategy

### Prerequisites
- NATS server running on `nats://localhost:4222` with JetStream enabled
- For RPC tests: `math-service` provider should be running

### Running Tests

```bash
# Run all tests
cd sdk/csharp/LightLink.Tests
dotnet test

# Run specific test
dotnet test -f net8.0 --filter "FullyQualifiedName~Client_Connect"
```

### Test Coverage

| Feature | Tests |
|---------|-------|
| Connection | Connect, ConnectAsync, IsConnected |
| RPC | Call, CallAsync with timeout |
| Pub/Sub | Publish, Subscribe, message handling |
| State Management | SetState, GetState |
| File Transfer | UploadFile, DownloadFile |

---

## Rollback Plan

If implementation fails, restore from backup:

```bash
cp sdk/csharp/LightLink/Client.cs.bak sdk/csharp/LightLink/Client.cs
```

---

## Related Skills

- @superpowers:verification-before-completion - Run verification before final commit
- @superpowers:test-driven-development - TDD methodology reference
