# LightLink TLS Certificate Management

## Overview

This directory contains the TLS certificate templates and references for LightLink NATS communication.

The certificate generation scripts have been moved to a separate repository: **[create_tls](https://github.com/LiteHomeLab/create_tls)**

## Quick Start

### Generate New Certificates

The certificate generation is now managed by the `create_tls` submodule:

```batch
cd deploy/nats/create_tls
setup-certs.bat
```

This will generate a timestamped folder `certs-YYYY-MM-DD_HH-MM/` containing:

```
certs-YYYY-MM-DD_HH-MM/
├── nats-server/     # Deploy to NATS server
│   ├── ca.crt
│   ├── server.crt
│   ├── server.key
│   └── README.txt
└── client/          # Deploy to ALL client services
    ├── ca.crt
    ├── client.crt
    ├── client.key
    └── README.txt
```

### Deploy Certificates

1. **NATS Server**: Copy `nats-server/` folder to your NATS server
2. **Client Services**: Copy `client/` folder to your service directory

## External Repository

- **Repository**: [git@github.com:LiteHomeLab/create_tls.git](https://github.com/LiteHomeLab/create_tls)
- **Location**: `deploy/nats/create_tls/` (Git submodule)
- **Purpose**: Certificate generation scripts and documentation

## Certificate Architecture

| Certificate | CN (Common Name) | Used By |
|-------------|------------------|---------|
| CA Root | `LightLink CA` | Signs all certificates |
| NATS Server | `nats-server` | NATS server only |
| Client | `lightlink-client` | All client services (shared) |

## Client Connection Configuration

### Go SDK

```go
tlsConfig := &client.TLSConfig{
    CaFile:     "client/ca.crt",
    CertFile:   "client/client.crt",
    KeyFile:    "client/client.key",
    ServerName: "nats-server",  // Must match server certificate CN
}
client, err := client.NewClient("tls://172.18.200.47:4222", tlsConfig)
```

### Python SDK

```python
tls_config = TLSConfig(
    ca_file="client/ca.crt",
    cert_file="client/client.crt",
    key_file="client/client.key",
    server_name="nats-server"  # Must match server certificate CN
)
client = Client(tls_config=tls_config)
```

### C# SDK

```csharp
Options opts = ConnectionFactory.GetDefaultOptions();
opts.Url = "tls://172.18.200.47:4222";
opts.SSL = true;
opts.SetCertificate("client/ca.crt", "client/client.crt", "client/client.key");
```

## Environment Variables (Optional)

```bash
set NATS_URL=tls://172.18.200.47:4222
set TLS_CA=client/ca.crt
set TLS_CERT=client/client.crt
set TLS_KEY=client/client.key
set TLS_SERVER_NAME=nats-server
```

## Security Notes

- **Private keys (.key files) must be kept secure!**
- Never commit .key files to version control
- Use secure channels to distribute certificate packages
- CA certificate (ca.crt) is public and can be freely distributed

## Update Submodule

To update the `create_tls` submodule to the latest version:

```batch
cd deploy/nats/create_tls
git pull origin main
cd ../..
git add deploy/nats/create_tls
git commit -m "chore: update create_tls submodule"
```

## Directory Structure

```
deploy/nats/
├── tls/              # This directory (templates and docs)
│   └── README.md
└── create_tls/       # Git submodule -> git@github.com:LiteHomeLab/create_tls.git
    ├── setup-certs.bat
    ├── find-openssl.bat
    ├── setup-certs-dependency.txt
    └── README.md
```

## Troubleshooting

If `setup-certs.bat` cannot find OpenSSL, run:

```batch
cd deploy/nats/create_tls
find-openssl.bat
```

This diagnostic tool will search for OpenSSL on your system.

## See Also

- [create_tls Repository](https://github.com/LiteHomeLab/create_tls)
- [Project Documentation](../../../docs/)
