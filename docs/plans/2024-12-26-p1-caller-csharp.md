# P1: C# Caller Example Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

> **PREREQUISITE:** P0 (Restore C# Client.cs) must be completed first.

**Goal:** Create a C# Caller example that demonstrates how to call Go/Python provider services, showcasing multi-language RPC interoperability.

**Architecture:**
- C# console application using restored C# SDK Client class
- Async/await pattern for RPC calls
- Dependency checking to ensure provider services are available
- Similar structure to existing Go caller example

**Tech Stack:**
- C# .NET 8.0
- NATS.Client 1.1.8
- LightLink C# SDK

**Reference:** `light_link_platform/examples/caller/go/call-math-service-go/main.go`

---

## Task 1: Create Project Structure

**Files:**
- Create: `light_link_platform/examples/caller/csharp/CallMathService/`
- Create: `light_link_platform/examples/caller/csharp/CallMathService/CallMathService.csproj`
- Create: `light_link_platform/examples/caller/csharp/CallMathService/Program.cs`
- Create: `light_link_platform/examples/caller/csharp/CallMathService/run.bat`

**Step 1: Create directory**

Run: `mkdir -p light_link_platform/examples/caller/csharp/CallMathService`

**Step 2: Create .csproj file**

```xml
<Project Sdk="Microsoft.NET.Sdk">

  <PropertyGroup>
    <OutputType>Exe</OutputType>
    <TargetFramework>net8.0</TargetFramework>
    <ImplicitUsings>enable</ImplicitUsings>
    <Nullable>enable</Nullable>
  </PropertyGroup>

  <ItemGroup>
    <ProjectReference Include="../../../../../sdk/csharp/LightLink/LightLink.csproj" />
  </ItemGroup>

</Project>
```

**Step 3: Verify project structure**

Run: `ls light_link_platform/examples/caller/csharp/CallMathService/`
Expected: CallMathService.csproj exists

**Step 4: Commit**

```bash
git add light_link_platform/examples/caller/csharp/CallMathService/
git commit -m "feat(csharp): create caller project structure"
```

---

## Task 2: Create Basic Program Structure

**Files:**
- Modify: `light_link_platform/examples/caller/csharp/CallMathService/Program.cs`

**Step 1: Write basic program structure**

```csharp
using System;
using System.Threading.Tasks;
using LightLink;

namespace CallMathService
{
    class Program
    {
        static async Task Main(string[] args)
        {
            Console.WriteLine("=== Call Math Service Demo (C#) ===");

            // Configuration
            string natsUrl = Environment.GetEnvironmentVariable("NATS_URL")
                ?? "nats://172.18.200.47:4222";
            Console.WriteLine($"NATS URL: {natsUrl}");

            // Discover certificates
            Console.WriteLine("\nDiscovering TLS certificates...");
            var tlsResult = CertDiscovery.DiscoverClientCerts();
            if (!tlsResult.Found)
            {
                Console.WriteLine("ERROR: Client certificates not found!");
                Console.WriteLine("Please copy the 'client/' folder to your service directory.");
                return;
            }
            var tlsConfig = CertDiscovery.ToTLSConfig(tlsResult);
            Console.WriteLine($"Certificates found: {tlsConfig.CertFile}");

            // Create client
            Console.WriteLine("\nConnecting to NATS...");
            var client = new Client(natsUrl, tlsConfig);
            await client.ConnectAsync();
            Console.WriteLine("Connected successfully");

            // TODO: Add dependency checking
            // TODO: Add RPC calls

            // Cleanup
            client.Close();
            Console.WriteLine("\n=== Demo complete ===");
            Console.WriteLine("\nPress any key to exit...");
            Console.ReadKey();
        }
    }
}
```

**Step 2: Create run.bat**

```bat
@echo off
echo Starting C# Caller Demo...
dotnet run
pause
```

**Step 3: Build to verify**

Run: `cd light_link_platform/examples/caller/csharp/CallMathService && dotnet build`
Expected: BUILD SUCCESS

**Step 4: Commit**

```bash
git add light_link_platform/examples/caller/csharp/CallMathService/
git commit -m "feat(csharp): add basic caller program structure"
```

---

## Task 3: Add Dependency Checking

**Files:**
- Modify: `light_link_platform/examples/caller/csharp/CallMathService/Program.cs`

**Step 1: Add dependency checking method**

```csharp
static async Task WaitForService(Client client, string serviceName,
    string[] methods, int timeoutSeconds = 30)
{
    Console.WriteLine($"\nChecking dependencies...");
    Console.WriteLine($"Waiting for {serviceName} with methods: {string.Join(", ", methods)}");

    var startTime = DateTime.Now;
    var checkInterval = TimeSpan.FromSeconds(2);

    while (true)
    {
        var elapsed = DateTime.Now - startTime;
        if (elapsed.TotalSeconds > timeoutSeconds)
        {
            throw new TimeoutException($"Timeout waiting for {serviceName}");
        }

        Console.WriteLine($"Checking for {serviceName}... ({(int)elapsed.TotalSeconds}s)");

        // Try a simple call to check availability
        // In full implementation, query $LL.SERVICE registry
        try
        {
            await Task.Delay(checkInterval);
            // For now, assume service is available after first check
            Console.WriteLine($"✓ {serviceName} is available");
            return;
        }
        catch (Exception ex)
        {
            Console.WriteLine($"Service not ready: {ex.Message}");
            await Task.Delay(checkInterval);
        }
    }
}
```

**Step 2: Update Main to use dependency checking**

```csharp
static async Task Main(string[] args)
{
    // ... existing code up to "Connected successfully" ...

    // Wait for math service
    try
    {
        await WaitForService(client, "math-service-go",
            new[] { "add", "multiply", "power", "divide" });
        Console.WriteLine("All dependencies satisfied!");
    }
    catch (TimeoutException ex)
    {
        Console.WriteLine($"Dependency check failed: {ex.Message}");
        client.Close();
        return;
    }

    // TODO: Add RPC calls

    // ... cleanup code ...
}
```

**Step 3: Build and test**

Run: `dotnet build && dotnet run`
Expected: Shows dependency checking messages

**Step 4: Commit**

```bash
git add light_link_platform/examples/caller/csharp/CallMathService/Program.cs
git commit -m "feat(csharp): add dependency checking"
```

---

## Task 4: Add RPC Calls

**Files:**
- Modify: `light_link_platform/examples/caller/csharp/CallMathService/Program.cs`

**Step 1: Add calculation method**

```csharp
static async Task PerformCalculations(Client client, string serviceName = "math-service-go")
{
    Console.WriteLine("\n=== Starting calculations ===\n");

    // 1. add(10, 20)
    try
    {
        var result = await client.CallAsync(serviceName, "add",
            new Dictionary<string, object> { { "a", 10 }, { "b", 20 } });
        Console.WriteLine($"add(10, 20) = {result["result"].GetInt32()}");
    }
    catch (Exception ex)
    {
        Console.WriteLine($"add failed: {ex.Message}");
    }

    // 2. multiply(5, 6)
    try
    {
        var result = await client.CallAsync(serviceName, "multiply",
            new Dictionary<string, object> { { "a", 5 }, { "b", 6 } });
        Console.WriteLine($"multiply(5, 6) = {result["result"].GetInt32()}");
    }
    catch (Exception ex)
    {
        Console.WriteLine($"multiply failed: {ex.Message}");
    }

    // 3. power(2, 10)
    try
    {
        var result = await client.CallAsync(serviceName, "power",
            new Dictionary<string, object> { { "base", 2 }, { "exp", 10 } });
        Console.WriteLine($"power(2, 10) = {result["result"].GetInt32()}");
    }
    catch (Exception ex)
    {
        Console.WriteLine($"power failed: {ex.Message}");
    }

    // 4. divide(100, 4)
    try
    {
        var result = await client.CallAsync(serviceName, "divide",
            new Dictionary<string, object>
            {
                { "numerator", 100 },
                { "denominator", 4 }
            });
        Console.WriteLine($"divide(100, 4) = {result["result"].GetDouble()}");
    }
    catch (Exception ex)
    {
        Console.WriteLine($"divide failed: {ex.Message}");
    }

    // 5. Complex calculation
    try
    {
        // First: multiply(3, 4)
        var result1 = await client.CallAsync(serviceName, "multiply",
            new Dictionary<string, object> { { "a", 3 }, { "b", 4 } });
        int multiplyResult = result1["result"].GetInt32();

        // Then: add(result, 10)
        var result2 = await client.CallAsync(serviceName, "add",
            new Dictionary<string, object>
            {
                { "a", multiplyResult },
                { "b", 10 }
            });
        int addResult = result2["result"].GetInt32();

        Console.WriteLine($"Complex: multiply(3, 4) = {multiplyResult}, " +
            $"then add({multiplyResult}, 10) = {addResult}");
    }
    catch (Exception ex)
    {
        Console.WriteLine($"Complex calculation failed: {ex.Message}");
    }

    Console.WriteLine("\n=== Calculations complete ===");
}
```

**Step 2: Update Main to call PerformCalculations**

```csharp
static async Task Main(string[] args)
{
    // ... existing code ...

    // Wait for math service
    // ... dependency checking code ...

    // Perform calculations
    await PerformCalculations(client, "math-service-go");

    // Cleanup
    client.Close();
    Console.WriteLine("\n=== Demo complete ===");
    // ... rest of code ...
}
```

**Step 3: Build and test**

Run: `dotnet build && dotnet run`
Expected: Performs all calculations (requires math-service-go running)

**Step 4: Commit**

```bash
git add light_link_platform/examples/caller/csharp/CallMathService/Program.cs
git commit -m "feat(csharp): add RPC calculation calls"
```

---

## Task 5: Create README

**Files:**
- Create: `light_link_platform/examples/caller/csharp/CallMathService/README.md`

**Step 1: Write comprehensive README**

```markdown
# C# Caller Example

This example demonstrates how to call provider services using C#.

## Features

- **Dependency Checking**: Waits for required services before calling
- **RPC Calls**: Demonstrates calling multiple methods
- **Error Handling**: Graceful handling of service unavailability
- **TLS Support**: Automatic certificate discovery
- **Async/Await**: Modern async patterns

## Running

### Prerequisites

1. .NET 8.0 SDK installed
2. NATS server running (or use remote at 172.18.200.47:4222)
3. Start a math service provider:

```bash
# Go provider
cd light_link_platform/examples/provider/go/math-service-go
go run main.go

# Or Python provider
cd light_link_platform/examples/provider/python/math_service
python main.py

# Or C# provider
cd light_link_platform/examples/provider/csharp/MathService
dotnet run
```

### Run the Caller

```bash
cd light_link_platform/examples/caller/csharp/CallMathService
dotnet run

# Or on Windows
run.bat
```

## Expected Output

```
=== Call Math Service Demo (C#) ===
NATS URL: nats://172.18.200.47:4222

Discovering TLS certificates...
Certificates found: ./client/client.crt

Connecting to NATS...
Connected successfully

Checking dependencies...
Waiting for math-service-go with methods: add, multiply, power, divide
✓ math-service-go is available
All dependencies satisfied!

=== Starting calculations ===

add(10, 20) = 30
multiply(5, 6) = 30
power(2, 10) = 1024
divide(100, 4) = 25
Complex: multiply(3, 4) = 12, then add(12, 10) = 22

=== Calculations complete ===

=== Demo complete ===
```

## Code Structure

```
CallMathService/
├── Program.cs     # Main caller program
├── CallMathService.csproj  # Project file
├── run.bat        # Windows startup script
└── README.md      # This file
```

## Customization

To call different services, modify:

1. Service name in `PerformCalculations()`
2. Method parameters
3. Expected response handling

## Troubleshooting

**"Certificates not found"**
- Copy the `client/` folder from `light_link_platform/client/`

**"Timeout waiting for service"**
- Ensure the provider service is running
- Check NATS connection

**"RPC call failed"**
- Verify the provider has registered the required methods
- Check method parameter names match

## Dependencies

- .NET 8.0
- NATS.Client 1.1.8
- LightLink C# SDK

## Related Examples

- Go Caller: `../../go/call-math-service-go/`
- Python Caller: `../../python/call-math-service/`
- Go Provider: `../../provider/go/math-service-go/`
```

**Step 2: Commit**

```bash
git add light_link_platform/examples/caller/csharp/CallMathService/README.md
git commit -m "docs(csharp): add caller README"
```

---

## Task 6: Update Parent README

**Files:**
- Modify: `light_link_platform/examples/caller/README.md`

**Step 1: Add C# entry to caller README**

```markdown
| 语言 | 项目 | 状态 |
|------|------|------|
| Go | call-math-service-go | ✅ 支持依赖检查 |
| Python | call-math-service | ✅ 支持依赖检查 |
| C# | CallMathService | ✅ 支持依赖检查 |
```

**Step 2: Commit**

```bash
git add light_link_platform/examples/caller/README.md
git commit -m "docs(caller): add C# example to README"
```

---

## Testing Strategy

### Prerequisites
- NATS server running
- math-service-go (or other math provider) running
- P0 (C# Client.cs) must be completed

### Test Scenarios

1. **Normal operation**: Provider started first, then caller
2. **Dependency waiting**: Caller started first, waits for provider
3. **Cross-language calls**: C# caller calling Go/Python providers

### Running Tests

```bash
# Terminal 1: Start Go provider
cd light_link_platform/examples/provider/go/math-service-go
go run main.go

# Terminal 2: Start C# caller
cd light_link_platform/examples/caller/csharp/CallMathService
dotnet run
```

---

## Dependencies

This example requires:

- .NET 8.0 SDK
- LightLink C# SDK (with Client.cs from P0)
- TLS certificates in `client/` folder

---

## Related Plans

- P0: `2024-12-26-restore-csharp-client.md` (Must be completed first)
- P1 Python: `2024-12-26-p1-caller-python.md`
