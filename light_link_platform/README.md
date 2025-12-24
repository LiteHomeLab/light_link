# Platform Examples

This directory contains the management platform and example services for the LightLink framework.

## Directory Structure

```
platform_examples/
├── manager_base/       # Management Platform (Backend + Frontend)
│   ├── server/        # Go backend server
│   ├── web/           # Vue 3 frontend
│   ├── data/          # Database and storage
│   └── README.md      # Platform documentation
│
├── go/                # Go Example Services
│   ├── metadata-demo/      # Metadata registration demo
│   └── metadata-client/    # Metadata query client demo
│
├── csharp/            # C# Example Services
│   ├── MetadataDemo/       # Metadata registration demo
│   ├── TextServiceDemo/    # Text processing service
│   ├── RpcDemo/            # RPC demo
│   └── PubSubDemo.cs       # PubSub demo
│
└── python/            # Python Example Services
    ├── metadata_demo.py    # Metadata registration demo
    ├── data_service.py     # Data processing service
    ├── rpc_service.py      # RPC service demo
    └── rpc_service_tls.py  # RPC service with TLS
```

## Quick Start

### 1. Start NATS Server

```bash
nats-server -config ../deploy/nats/nats-server.conf
```

### 2. Start Management Platform

```bash
cd manager_base/server
go run main.go
```

Then open browser to `http://localhost:8080`

### 3. Run Example Services

**Go Math Service:**
```bash
cd go/metadata-demo
go run main.go
```

**C# Text Service:**
```bash
cd csharp/TextServiceDemo
dotnet run
```

**Python Data Service:**
```bash
cd python
python data_service.py
```

## Services Overview

| Service | Language | Methods |
|---------|----------|---------|
| math-service | Go | add, multiply, power, divide |
| csharp-text-service | C# | reverse, uppercase, wordcount |
| python-data-service | Python | filter, transform, aggregate |

## Development Notes

- All services register metadata with the management platform automatically
- Services send heartbeat every 30 seconds
- Platform shows service status, methods, and allows RPC calls
- TLS support available (see `rpc_service_tls.py`)
