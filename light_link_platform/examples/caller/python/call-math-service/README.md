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
[call-math-service-python] Checking for math-service-go... (2s)
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
