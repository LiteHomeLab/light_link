# TLS Certificate Management

## Directory Structure

```
deploy/nats/tls/
├── ca.key, ca.crt           # CA Certificate (Root of Trust)
├── server.key, server.crt   # NATS Server Certificate
├── generate-certs.bat       # Generate all certificates
├── generate-service-cert.bat # Generate single service certificate
├── cert-manifest.txt        # List of all services
├── demo-service.*           # Demo service certificates
├── test-service.*           # Test service certificates
├── client-app.*             # Client app certificates
└── clients/                 # Client distribution packages
    ├── demo-service/
    │   ├── ca.crt
    │   ├── demo-service.crt
    │   ├── demo-service.key
    │   └── README.md
    ├── test-service/
    │   └── ...
    └── {service-name}/
        └── ...
```

## Generate New Service Certificate

### Quick Start

```batch
cd deploy/nats/tls
generate-service-cert.bat my-service
```

This creates:
1. Certificate files in current directory: `my-service.key`, `my-service.crt`
2. Client distribution package: `clients/my-service/`
   - `ca.crt` - CA Root Certificate
   - `my-service.crt` - Service Certificate
   - `my-service.key` - Service Private Key
   - `README.md` - Deployment instructions

### Deploy to Service

```batch
# Copy client package to service directory
xcopy /E /I deploy\nats\tls\clients\my-service my-service\tls\

# Or manually
mkdir my-service\tls
copy deploy\nats\tls\clients\my-service\*.* my-service\tls\
```

## Certificate Files

| File | Purpose | Distribution |
|------|---------|--------------|
| `ca.crt` | CA Root Certificate | All clients (public) |
| `ca.key` | CA Private Key | CA server only (SECRET) |
| `server.crt` | NATS Server Certificate | NATS server only |
| `server.key` | NATS Server Private Key | NATS server only (SECRET) |
| `{service}.crt` | Service Certificate | Service deployment |
| `{service}.key` | Service Private Key | Service deployment (SECRET) |

## Default TLS Configuration

After copying to `tls/` directory, use these default paths:

### Go SDK

```go
tlsConfig := &client.TLSConfig{
    CaFile:     "tls/ca.crt",
    CertFile:   "tls/my-service.crt",
    KeyFile:    "tls/my-service.key",
    ServerName: "nats-server",
}
client, err := client.NewClient("tls://172.18.200.47:4222", tlsConfig)
```

### Python SDK

```python
tls_config = TLSConfig(
    ca_file="tls/ca.crt",
    cert_file="tls/my-service.crt",
    key_file="tls/my-service.key",
    server_name="nats-server"
)
client = Client(tls_config=tls_config)
```

### Environment Variables (Optional Override)

```bash
set NATS_URL=tls://172.18.200.47:4222
set TLS_CA=tls/ca.crt
set TLS_CERT=tls/my-service.crt
set TLS_KEY=tls/my-service.key
set TLS_SERVER_NAME=nats-server
```

## Security Notes

- **Private Keys (.key files) must be kept secure!**
- Never commit .key files to version control
- Use secure channels to distribute client packages
- CA certificate (ca.crt) is public and can be freely distributed

## Certificate Manifest

See `cert-manifest.txt` for a list of all generated services.
