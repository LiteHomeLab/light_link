# TLS Certificate Automation Implementation Plan

**Date:** 2025-12-24
**Status:** READY FOR IMPLEMENTATION

---

## Executive Summary

This plan implements TLS certificate automation for LightLink deployments:

1. **SDK Layer**: Add `ServerName` field for proper certificate verification (removes `InsecureSkipVerify`)
2. **Application Layer**: Define default certificate paths with override support
3. **Certificate Generation**: Create script that generates certificates and manifest only (no SDK code modification)

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Application Layer                          │
│  manager_base, examples/ - Default path: tls/ directory          │
│  Deployment: Copy clients/{service-name}/ to service's tls/      │
├─────────────────────────────────────────────────────────────────┤
│                        SDK Layer                                │
│  sdk/go, sdk/python, sdk/csharp - ServerName field only         │
│  No default paths defined at SDK level                           │
├─────────────────────────────────────────────────────────────────┤
│                   Certificate Generation                        │
│  generate-service-cert.bat - Generate certs + client bundle     │
│  Output: clients/{service-name}/ folder (ready to deploy)        │
└─────────────────────────────────────────────────────────────────┘

Client Deployment Flow:
┌─────────────────────────────────────────────────────────────────┐
│  1. Generate: generate-service-cert.bat my-service              │
│                                                                   │
│  2. Output folder created:                                       │
│     deploy/nats/tls/clients/my-service/                         │
│     ├── ca.crt               (CA certificate)                   │
│     ├── my-service.crt       (Service certificate)              │
│     ├── my-service.key       (Service private key)              │
│     └── README.md            (Usage instructions)                │
│                                                                   │
│  3. Deploy to client service:                                    │
│     Copy deploy/nats/tls/clients/my-service/*                    │
│     To: {service-directory}/tls/                                 │
│                                                                   │
│  4. Client config (default):                                     │
│     ca: tls/ca.crt                                              │
│     cert: tls/my-service.crt                                    │
│     key: tls/my-service.key                                     │
└─────────────────────────────────────────────────────────────────┘
```

---

## Phase 1: SDK Layer - Add ServerName Field

### Task 1.1: Go SDK

**Files:**
- `sdk/go/types/types.go` - Add ServerName to TLSConfig struct
- `sdk/go/client/connection.go` - Use ServerName, remove InsecureSkipVerify
- `sdk/go/client/*_test.go` - Update tests

**Changes:**

```go
// types.go
type TLSConfig struct {
    CaFile     string `json:"ca_file"`
    CertFile   string `json:"cert_file"`
    KeyFile    string `json:"key_file"`
    ServerName string `json:"server_name,omitempty"` // NEW
}

// connection.go
func CreateTLSOption(config *TLSConfig) (nats.Option, error) {
    cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
    // ...

    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
        RootCAs:      pool,
        MinVersion:   tls.VersionTLS12,
        ServerName:   config.ServerName, // NEW: No InsecureSkipVerify
    }

    return nats.Secure(tlsConfig), nil
}
```

### Task 1.2: Python SDK

**Files:**
- `sdk/python/lightlink/client.py`

**Changes:**

```python
class TLSConfig:
    def __init__(self, ca_file, cert_file, key_file, server_name=None):
        self.ca_file = ca_file
        self.cert_file = cert_file
        self.key_file = key_file
        self.server_name = server_name  # NEW

async def connect(self):
    if self.tls_config:
        ssl_ctx = ssl.create_default_context(ssl.Purpose.SERVER_AUTH)
        # ...
        if self.tls_config.server_name:
            ssl_ctx.server_hostname = self.tls_config.server_name
```

---

## Phase 2: Application Layer - Default Certificate Paths

### Task 2.1: Manager Base Configuration

**Files:**
- `light_link_platform/manager_base/server/config/config.go`
- `light_link_platform/manager_base/server/main.go`

**Changes:**

```go
// config.go
type TLSConfig struct {
    Enabled    bool   `yaml:"enabled"`
    Cert       string `yaml:"cert"`
    Key        string `yaml:"key"`
    CA         string `yaml:"ca"`
    ServerName string `yaml:"server_name,omitempty"` // NEW
}

func GetDefaultConfig() *Config {
    return &Config{
        NATS: NATSConfig{
            TLS: TLSConfig{
                Enabled:    false,
                CA:         "tls/ca.crt",       // Default: tls/ directory
                Cert:       "tls/manager.crt",  // Default: tls/ directory
                Key:        "tls/manager.key",  // Default: tls/ directory
                ServerName: "nats-server",
            },
        },
    }
}
```

**Deployment:**
1. Generate certificate: `generate-service-cert.bat manager`
2. Copy `deploy/nats/tls/clients/manager/*` to `manager_base/server/tls/`
3. Enable TLS in console.yaml

### Task 2.2: Examples - Environment Variable Support

**Files:**
- `examples/config.go` (or create new)
- `light_link_platform/examples/go/*/main.go`
- `light_link_platform/examples/python/*.py`

```go
// Go examples
config := &Config{
    NATSURL:  os.Getenv("NATS_URL"),
    CAFile:   os.Getenv("TLS_CA"),
    CertFile: os.Getenv("TLS_CERT"),
    KeyFile:  os.Getenv("TLS_KEY"),
    ServerName: os.Getenv("TLS_SERVER_NAME"),
}
```

---

## Phase 3: Certificate Generation Script

### Task 3.1: Create generate-service-cert.bat

**File:** `deploy/nats/tls/generate-service-cert.bat`

```batch
@echo off
REM Service Certificate Generator for LightLink
REM Usage: generate-service-cert.bat <service-name>

setlocal enabledelayedexpansion

if "%~1"=="" (
    echo Usage: generate-service-cert.bat ^<service-name^>
    echo Example: generate-service-cert.bat my-service
    exit /b 1
)

set "SERVICE_NAME=%~1"
set "SCRIPT_DIR=%~dp0"
cd /d "%SCRIPT_DIR%"

echo ========================================
echo LightLink Service Certificate Generator
echo ========================================
echo.
echo Service Name: %SERVICE_NAME%
echo.

REM Check OpenSSL
where openssl >nul 2>&1
if errorlevel 1 (
    echo ERROR: OpenSSL not found in PATH
    echo Please install OpenSSL or add it to PATH
    pause
    exit /b 1
)

REM Check CA exists
if not exist "ca.crt" (
    echo ERROR: CA certificate not found
    echo Please run generate-certs.bat first to generate CA certificate
    pause
    exit /b 1
)

if not exist "ca.key" (
    echo ERROR: CA private key not found
    echo Please run generate-certs.bat first to generate CA certificate
    pause
    exit /b 1
)

REM Generate private key
echo [1/4] Generating private key for %SERVICE_NAME%...
openssl genrsa -out %SERVICE_NAME%.key 2048
if errorlevel 1 goto error
echo Private key generated successfully

REM Generate CSR
echo [2/4] Generating certificate signing request...
openssl req -new -key %SERVICE_NAME%.key -out %SERVICE_NAME%.csr -subj "/CN=%SERVICE_NAME%"
if errorlevel 1 goto error
echo CSR generated successfully

REM Sign with CA
echo [3/4] Signing certificate with CA...
openssl x509 -req -days 10950 -in %SERVICE_NAME%.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out %SERVICE_NAME%.crt
if errorlevel 1 goto error
echo Certificate signed successfully

REM Cleanup temporary files
echo.
echo Cleaning up temporary files...
del %SERVICE_NAME%.csr 2>nul
del ca.srl 2>nul

REM Create client distribution directory
echo [4/4] Creating client distribution package...
if not exist "clients" mkdir clients
if not exist "clients\%SERVICE_NAME%" mkdir clients\%SERVICE_NAME%

REM Copy files to client distribution directory
copy /Y ca.crt clients\%SERVICE_NAME%\ >nul
copy /Y %SERVICE_NAME%.crt clients\%SERVICE_NAME%\ >nul
copy /Y %SERVICE_NAME%.key clients\%SERVICE_NAME%\ >nul
echo Client package created: clients\%SERVICE_NAME%\

REM Generate README for client
echo # TLS Certificate Package for %SERVICE_NAME% > clients\%SERVICE_NAME%\README.md
echo. >> clients\%SERVICE_NAME%\README.md
echo Generated: %date% %time% >> clients\%SERVICE_NAME%\README.md
echo. >> clients\%SERVICE_NAME%\README.md
echo ## Files >> clients\%SERVICE_NAME%\README.md
echo. >> clients\%SERVICE_NAME%\README.md
echo - ca.crt - CA Root Certificate (trusted certificate) >> clients\%SERVICE_NAME%\README.md
echo - %SERVICE_NAME%.crt - Service Certificate >> clients\%SERVICE_NAME%\README.md
echo - %SERVICE_NAME%.key - Service Private Key (KEEP SECRET!) >> clients\%SERVICE_NAME%\README.md
echo. >> clients\%SERVICE_NAME%\README.md
echo ## Deployment >> clients\%SERVICE_NAME%\README.md
echo. >> clients\%SERVICE_NAME%\README.md
echo Copy all files in this folder to your service directory: >> clients\%SERVICE_NAME%\README.md
echo. >> clients\%SERVICE_NAME%\README.md
echo ```bash >> clients\%SERVICE_NAME%\README.md
echo # Copy to service directory >> clients\%SERVICE_NAME%\README.md
echo mkdir tls >> clients\%SERVICE_NAME%\README.md
echo copy *.* tls\ >> clients\%SERVICE_NAME%\README.md
echo ``` >> clients\%SERVICE_NAME%\README.md
echo. >> clients\%SERVICE_NAME%\README.md
echo ## Configuration (Default Paths) >> clients\%SERVICE_NAME%\README.md
echo. >> clients\%SERVICE_NAME%\README.md
echo After copying to tls/ directory, use these default paths: >> clients\%SERVICE_NAME%\README.md
echo. >> clients\%SERVICE_NAME%\README.md
echo ### Go SDK >> clients\%SERVICE_NAME%\README.md
echo ```go >> clients\%SERVICE_NAME%\README.md
echo tlsConfig := ^&client.TLSConfig{ >> clients\%SERVICE_NAME%\README.md
echo     CaFile:     "tls/ca.crt", >> clients\%SERVICE_NAME%\README.md
echo     CertFile:   "tls/%SERVICE_NAME%.crt", >> clients\%SERVICE_NAME%\README.md
echo     KeyFile:    "tls/%SERVICE_NAME%.key", >> clients\%SERVICE_NAME%\README.md
echo     ServerName: "nats-server", >> clients\%SERVICE_NAME%\README.md
echo } >> clients\%SERVICE_NAME%\README.md
echo ``` >> clients\%SERVICE_NAME%\README.md
echo. >> clients\%SERVICE_NAME%\README.md
echo ### Python SDK >> clients\%SERVICE_NAME%\README.md
echo ```python >> clients\%SERVICE_NAME%\README.md
echo tls_config = TLSConfig( >> clients\%SERVICE_NAME%\README.md
echo     ca_file="tls/ca.crt", >> clients\%SERVICE_NAME%\README.md
echo     cert_file="tls/%SERVICE_NAME%.crt", >> clients\%SERVICE_NAME%\README.md
echo     key_file="tls/%SERVICE_NAME%.key", >> clients\%SERVICE_NAME%\README.md
echo     server_name="nats-server" >> clients\%SERVICE_NAME%\README.md
echo ) >> clients\%SERVICE_NAME%\README.md
echo ``` >> clients\%SERVICE_NAME%\README.md
echo. >> clients\%SERVICE_NAME%\README.md
echo ### Environment Variables (Optional) >> clients\%SERVICE_NAME%\README.md
echo ```bash >> clients\%SERVICE_NAME%\README.md
echo set NATS_URL=tls://172.18.200.47:4222 >> clients\%SERVICE_NAME%\README.md
echo set TLS_CA=tls/ca.crt >> clients\%SERVICE_NAME%\README.md
echo set TLS_CERT=tls/%SERVICE_NAME%.crt >> clients\%SERVICE_NAME%\README.md
echo set TLS_KEY=tls/%SERVICE_NAME%.key >> clients\%SERVICE_NAME%\README.md
echo set TLS_SERVER_NAME=nats-server >> clients\%SERVICE_NAME%\README.md
echo ``` >> clients\%SERVICE_NAME%\README.md

echo.
echo ========================================
echo Certificate generation completed!
echo ========================================
echo.
echo Generated files in current directory:
echo - %SERVICE_NAME%.key (Private Key - KEEP SECRET!)
echo - %SERVICE_NAME%.crt (Certificate)
echo.
echo Client distribution package: clients\%SERVICE_NAME%\
echo   - ca.crt
echo   - %SERVICE_NAME%.crt
echo   - %SERVICE_NAME%.key
echo   - README.md
echo.
echo To deploy:
echo   1. Copy clients\%SERVICE_NAME%\* to your service directory\tls\
echo   2. Use default paths: tls/ca.crt, tls/%SERVICE_NAME%.crt, tls/%SERVICE_NAME%.key
echo   3. Set ServerName to "nats-server" for certificate verification
echo.

REM Append to manifest
echo %SERVICE_NAME%: clients\%SERVICE_NAME% >> cert-manifest.txt

goto end

:error
echo.
echo ========================================
echo ERROR: Certificate generation failed
echo ========================================
echo.
pause
exit /b 1

:end
pause
```

### Task 3.2: Create cert-manifest.txt

**File:** `deploy/nats/tls/cert-manifest.txt`

```
# LightLink Certificate Manifest
# Format: <service-name>: <client-distribution-path>

# NATS Server (server-side only, no client package)
nats-server: server.crt, server.key

# Demo Services
demo-service: clients/demo-service
test-service: clients/test-service
client-app: clients/client-app

# Manager Services
manager: clients/manager
```

### Task 3.3: Update README.md

**File:** `deploy/nats/tls/README.md`

```markdown
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
```

---

## Phase 4: Testing

### Task 4.1: Update All Tests

Add `ServerName` field to all TLSConfig instances in test files.

### Task 4.2: Integration Testing

```bash
# 1. Generate new certificate
cd deploy/nats/tls
generate-service-cert.bat test-new-service

# Output creates:
# - test-new-service.key, test-new-service.crt (in current dir)
# - clients/test-new-service/ (distribution package)

# 2. Deploy to Go example service
cd light_link_platform/examples/go/metadata-demo
mkdir tls
xcopy /E /I ..\..\..\deploy\nats\tls\clients\test-new-service tls\

# Run with default paths (uses tls/ directory)
set TLS_SERVER_NAME=nats-server
go run main.go

# 3. Deploy to Python example service
cd light_link_platform/examples/python
mkdir tls
xcopy /E /I ..\..\deploy\nats\tls\clients\test-new-service tls\

# Run with default paths
set TLS_SERVER_NAME=nats-server
python data_service.py

# 4. Verify connection works with ServerName (no InsecureSkipVerify)
```

---

## File Checklist

### Files to Create

| File | Purpose |
|------|---------|
| `deploy/nats/tls/generate-service-cert.bat` | Certificate generation script |
| `deploy/nats/tls/cert-manifest.txt` | Certificate manifest file |
| `deploy/nats/tls/clients/.gitkeep` | Client distribution directory |

### Files to Modify

| File | Changes |
|------|---------|
| `sdk/go/types/types.go` | Add ServerName to TLSConfig |
| `sdk/go/client/connection.go` | Use ServerName, remove InsecureSkipVerify |
| `sdk/python/lightlink/client.py` | Add server_name field |
| `light_link_platform/manager_base/server/config/config.go` | Add ServerName, default tls/ paths |
| `light_link_platform/manager_base/server/main.go` | Use ServerName in TLS setup |
| `sdk/go/client/*_test.go` | Add ServerName to test configs |
| `light_link_platform/examples/*/main.go` | Use tls/ default paths |
| `deploy/nats/tls/README.md` | Update with client package instructions |
| `CLAUDE.md` | Document TLS configuration |

---

## Success Criteria

- [ ] generate-service-cert.bat creates valid certificates
- [ ] Client distribution package created in `clients/{service-name}/`
- [ ] Client package contains: ca.crt, {service}.crt, {service}.key, README.md
- [ ] cert-manifest.txt tracks all generated services
- [ ] Go SDK connects without InsecureSkipVerify
- [ ] Python SDK uses server_name for verification
- [ ] Manager Base supports TLS with default tls/ paths
- [ ] Copying client package to service tls/ directory works out-of-the-box
- [ ] All tests pass with ServerName configured
- [ ] Documentation is complete

---

## Rollback Strategy

If issues occur:
- Set `tls.enabled: false` in console.yaml
- Keep ServerName field but default to empty (optional verification)
- Certificate generation failures: use existing generate-certs.bat manually

---

## Critical Files (Top 5)

1. `sdk/go/client/connection.go` - Core TLS logic
2. `sdk/go/types/types.go` - TLSConfig definition
3. `deploy/nats/tls/generate-service-cert.bat` - Certificate generation
4. `light_link_platform/manager_base/server/config/config.go` - App config
5. `sdk/python/lightlink/client.py` - Python TLS support
