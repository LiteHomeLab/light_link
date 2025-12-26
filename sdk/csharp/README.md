# LightLink C# SDK

LightLink C# SDK provides client and server functionality for NATS-based microservices communication.

## Features

- **RPC Client/Service**: Remote procedure calls with timeout support
- **Pub/Sub**: Message publishing and subscribing
- **State Management**: KV-based state storage (simplified in-memory implementation)
- **File Transfer**: Upload/download files (simplified in-memory implementation)
- **TLS Support**: Automatic certificate discovery and configuration
- **Metadata Registration**: Service method metadata for discovery

> **Note**: The current implementation uses simplified in-memory storage for state management and file transfer. For production use with NATS JetStream KV and Object Store, upgrade to NATS.Net.Client package.

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

### RPC Service

```csharp
using LightLink;

var svc = new Service("my-service", "nats://localhost:4222");

// Register method
svc.RegisterRPC("add", async (args) => {
    int a = args.ContainsKey("a") ? Convert.ToInt32(args["a"]) : 0;
    int b = args.ContainsKey("b") ? Convert.ToInt32(args["b"]) : 0;
    return new Dictionary<string, object> { { "result", a + b } };
});

svc.Start();
```

### Metadata Registration

```csharp
var meta = new MethodMetadata {
    Name = "add",
    Description = "Add two numbers",
    Params = new List<ParameterMetadata> {
        new() { Name = "a", Type = "number", Required = true, Description = "First number" },
        new() { Name = "b", Type = "number", Required = true, Description = "Second number" }
    }
};

svc.RegisterMethodWithMetadata("add", handler, meta);
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
Console.WriteLine($"Enabled: {config["enabled"]}");
```

### File Transfer

```csharp
// Upload
string fileId = client.UploadFile("local.txt", "remote.txt");
Console.WriteLine($"File uploaded: {fileId}");

// Download
client.DownloadFile(fileId, "downloaded.txt");
```

## TLS Configuration

The SDK automatically discovers certificates in the following locations:

- `./client`
- `../client`
- `../../client`
- `../../../client`
- `../../../../client`
- `../../../../../client`

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
- NATS server with JetStream enabled (optional, for advanced features)

## Protocol Compatibility

- JSON format is fully compatible with Go SDK
- Supports NATS JetStream (requires NATS.Net.Client upgrade)
- Supports parameter validation via metadata

## Limitations

Current implementation (using NATS.Client 1.1.8):
- State management uses in-memory cache with pub/sub synchronization
- File transfer uses in-memory store
- No JetStream KV or Object Store support

For production use with full JetStream support:
1. Upgrade to `NATS.Net.Client` (modern NATS C# client)
2. Update state management to use JetStream KV
3. Update file transfer to use JetStream Object Store

## Examples

See `light_link_platform/examples/` for full examples:
- `provider/csharp/MathService` - RPC service example
- `caller/csharp/CallMathService` - RPC client example
- `examples/PubSubDemo` - Publish/subscribe example
- `examples/MetadataDemo` - Metadata registration example
- `examples/TextServiceDemo` - Text processing service example

## License

MIT License
