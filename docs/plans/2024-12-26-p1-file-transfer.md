# P1: File Transfer Examples Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create file transfer examples in Go, C#, and Python that demonstrate how to upload and download files using the LightLink SDK.

**Architecture:**
- Three separate examples (Go, C#, Python) following the same pattern
- Each example demonstrates:
  - Creating a test file
  - Uploading to Object Store
  - Downloading from Object Store
  - Verifying file integrity

**Tech Stack:**
- Go 1.21+
- C# .NET 8.0
- Python 3.8+
- NATS JetStream Object Store

**Reference:** `examples/file-transfer-demo/main.go` (SDK example)

---

## Task 1: Create Directory Structure

**Files:**
- Create: `light_link_platform/examples/file-transfer/go/`
- Create: `light_link_platform/examples/file-transfer/csharp/`
- Create: `light_link_platform/examples/file-transfer/python/`

**Step 1: Create directories**

Run:
```bash
mkdir -p light_link_platform/examples/file-transfer/go/file-transfer-demo
mkdir -p light_link_platform/examples/file-transfer/csharp/FileTransferDemo
mkdir -p light_link_platform/examples/file-transfer/python/file_transfer_demo
```

**Step 2: Verify directories created**

Run: `ls light_link_platform/examples/file-transfer/`
Expected: go/, csharp/, python/ directories exist

**Step 3: Create main README**

```markdown
# File Transfer Examples

This directory contains examples demonstrating file transfer capabilities using LightLink SDK.

## Features

- **Upload files** to NATS Object Store
- **Download files** from Object Store
- **File integrity verification** using checksums
- **Large file support** (files are automatically chunked)

## Examples by Language

| Language | Project | Description |
|----------|---------|-------------|
| Go | file-transfer-demo | Upload and download files |
| C# | FileTransferDemo | Upload and download files |
| Python | file_transfer_demo | Upload and download files |

## Running the Examples

### Prerequisites

1. NATS server running with JetStream enabled
2. TLS certificates in `client/` folder (for all examples)

### Go Example

```bash
cd light_link_platform/examples/file-transfer/go/file-transfer-demo
go run main.go
```

### C# Example

```bash
cd light_link_platform/examples/file-transfer/csharp/FileTransferDemo
dotnet run
```

### Python Example

```bash
cd light_link_platform/examples/file-transfer/python/file_transfer_demo
python main.py
```

## How It Works

1. **Create test file** - A sample file is created locally
2. **Upload** - File is uploaded to NATS Object Store
3. **Get File ID** - A unique file ID is returned
4. **Download** - File is downloaded using the file ID
5. **Verify** - Downloaded file is compared with original

## Troubleshooting

**"JetStream not enabled"**
- Ensure NATS server is started with JetStream: `nats-server -js`

**"Upload failed"**
- Check NATS connection
- Verify TLS certificates are in place

**"Download failed"**
- Ensure file ID is correct
- Check Object Store bucket exists
```

**Step 4: Commit**

```bash
git add light_link_platform/examples/file-transfer/
git commit -m "feat(file-transfer): create directory structure and main README"
```

---

## Task 2: Implement Go File Transfer Example

**Files:**
- Create: `light_link_platform/examples/file-transfer/go/file-transfer-demo/main.go`
- Create: `light_link_platform/examples/file-transfer/go/file-transfer-demo/run.bat`

**Step 1: Write Go example**

```go
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/LiteHomeLab/light_link/examples"
	"github.com/LiteHomeLab/light_link/sdk/go/client"
	"github.com/WQGroup/logger"
)

func main() {
	logger.SetLoggerName("file-transfer-go")
	logger.Info("=== File Transfer Demo (Go) ===")

	config := examples.GetConfig()
	logger.Infof("NATS URL: %s", config.NATSURL)

	// Create client
	logger.Info("Connecting to NATS...")
	cli, err := client.NewClient(config.NATSURL, client.WithAutoTLS())
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer cli.Close()
	logger.Info("Connected successfully")

	// Create test file
	testFile := "test_upload.txt"
	downloadFile := "test_download.txt"
	testContent := "Hello, LightLink File Transfer from Go!\n" +
		"This is a test file for demonstrating file upload and download."

	logger.Info("\n[1/4] Creating test file...")
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		log.Fatalf("Failed to create test file: %v", err)
	}
	defer func() {
		os.Remove(testFile)
		os.Remove(downloadFile)
	}()
	logger.Infof("Test file created: %s", testFile)

	// Upload file
	logger.Info("\n[2/4] Uploading file...")
	fileID, err := cli.UploadFile(testFile, testFile)
	if err != nil {
		log.Fatalf("Upload failed: %v", err)
	}
	logger.Infof("File uploaded successfully! File ID: %s", fileID)

	// Download file
	logger.Info("\n[3/4] Downloading file...")
	err = cli.DownloadFile(fileID, downloadFile)
	if err != nil {
		log.Fatalf("Download failed: %v", err)
	}
	logger.Infof("File downloaded successfully: %s", downloadFile)

	// Verify file
	logger.Info("\n[4/4] Verifying file...")
	downloadedContent, err := os.ReadFile(downloadFile)
	if err != nil {
		log.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(downloadedContent) == testContent {
		logger.Info("✓ File verification PASSED - Content matches!")
	} else {
		logger.Error("✗ File verification FAILED - Content mismatch!")
	}

	logger.Info("\n=== Demo complete ===")
}
```

**Step 2: Create run.bat**

```bat
@echo off
echo Starting Go File Transfer Demo...
go run main.go
pause
```

**Step 3: Test the example**

Run: `go run main.go`
Expected: Creates, uploads, downloads, and verifies a test file

**Step 4: Commit**

```bash
git add light_link_platform/examples/file-transfer/go/file-transfer-demo/
git commit -m "feat(file-transfer): add Go example"
```

---

## Task 3: Implement C# File Transfer Example

**Files:**
- Create: `light_link_platform/examples/file-transfer/csharp/FileTransferDemo/FileTransferDemo.csproj`
- Create: `light_link_platform/examples/file-transfer/csharp/FileTransferDemo/Program.cs`
- Create: `light_link_platform/examples/file-transfer/csharp/FileTransferDemo/run.bat`

**Step 1: Create .csproj file**

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

**Step 2: Write Program.cs**

```csharp
using System;
using System.IO;
using System.Threading.Tasks;
using LightLink;

namespace FileTransferDemo
{
    class Program
    {
        static async Task Main(string[] args)
        {
            Console.WriteLine("=== File Transfer Demo (C#) ===");

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
                return;
            }
            var tlsConfig = CertDiscovery.ToTLSConfig(tlsResult);

            // Create client
            Console.WriteLine("\nConnecting to NATS...");
            var client = new Client(natsUrl, tlsConfig);
            await client.ConnectAsync();
            Console.WriteLine("Connected successfully");

            // File paths
            string testFile = "test_upload.txt";
            string downloadFile = "test_download.txt";
            string testContent = "Hello, LightLink File Transfer from C#!\n" +
                "This is a test file for demonstrating file upload and download.";

            try
            {
                // Create test file
                Console.WriteLine("\n[1/4] Creating test file...");
                File.WriteAllText(testFile, testContent);
                Console.WriteLine($"Test file created: {testFile}");

                // Upload file
                Console.WriteLine("\n[2/4] Uploading file...");
                string fileId = await client.UploadFileAsync(testFile, testFile);
                Console.WriteLine($"File uploaded successfully! File ID: {fileId}");

                // Download file
                Console.WriteLine("\n[3/4] Downloading file...");
                await client.DownloadFileAsync(fileId, downloadFile);
                Console.WriteLine($"File downloaded successfully: {downloadFile}");

                // Verify file
                Console.WriteLine("\n[4/4] Verifying file...");
                string downloadedContent = File.ReadAllText(downloadFile);

                if (downloadedContent == testContent)
                {
                    Console.WriteLine("✓ File verification PASSED - Content matches!");
                }
                else
                {
                    Console.WriteLine("✗ File verification FAILED - Content mismatch!");
                }

                Console.WriteLine("\n=== Demo complete ===");
            }
            finally
            {
                // Cleanup
                if (File.Exists(testFile)) File.Delete(testFile);
                if (File.Exists(downloadFile)) File.Delete(downloadFile);
                client.Close();
            }

            Console.WriteLine("\nPress any key to exit...");
            Console.ReadKey();
        }
    }
}
```

**Step 3: Create run.bat**

```bat
@echo off
echo Starting C# File Transfer Demo...
dotnet run
pause
```

**Step 4: Build and test**

Run: `dotnet build && dotnet run`
Expected: Creates, uploads, downloads, and verifies a test file

**Step 5: Commit**

```bash
git add light_link_platform/examples/file-transfer/csharp/FileTransferDemo/
git commit -m "feat(file-transfer): add C# example"
```

---

## Task 4: Implement Python File Transfer Example

**Files:**
- Create: `light_link_platform/examples/file-transfer/python/file_transfer_demo/main.py`
- Create: `light_link_platform/examples/file-transfer/python/file_transfer_demo/run.bat`

**Step 1: Check if Python SDK has file transfer methods**

Run: `grep -n "def.*[Uu]pload\|def.*[Dd]ownload" sdk/python/lightlink/client.py`
Expected: Check if methods exist or need to be implemented

**Step 2: Write Python example**

```python
#!/usr/bin/env python3
"""
LightLink Python File Transfer Example

Demonstrates uploading and downloading files using LightLink SDK.
"""

import asyncio
import logging
import os
import sys

# Add parent directory to path for imports
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../../..'))

from lightlink.client import Client, discover_client_certs

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='[%(name)s] %(message)s'
)
logger = logging.getLogger('file-transfer-python')


async def main():
    logger.info("=== File Transfer Demo (Python) ===")

    # Configuration
    nats_url = os.getenv('NATS_URL', 'nats://172.18.200.47:4222')
    logger.info(f"NATS URL: {nats_url}")

    # Discover certificates
    logger.info("\nDiscovering TLS certificates...")
    try:
        certs = discover_client_certs()
        logger.info(f"Certificates found: {certs.cert_file}")
    except FileNotFoundError as e:
        logger.error(f"Certificates not found: {e}")
        return

    # Create client
    logger.info("\nConnecting to NATS...")
    client = Client(nats_url)
    await client.connect(
        tls_cert_file=certs.cert_file,
        tls_key_file=certs.key_file,
        tls_ca_file=certs.ca_file
    )
    logger.info("Connected successfully")

    # File paths
    test_file = "test_upload.txt"
    download_file = "test_download.txt"
    test_content = ("Hello, LightLink File Transfer from Python!\n"
                   "This is a test file for demonstrating file upload and download.")

    try:
        # Create test file
        logger.info("\n[1/4] Creating test file...")
        with open(test_file, 'w') as f:
            f.write(test_content)
        logger.info(f"Test file created: {test_file}")

        # Upload file
        logger.info("\n[2/4] Uploading file...")
        # Note: This assumes upload_file method exists in Python SDK
        # If not, this will need to be implemented first
        try:
            file_id = await client.upload_file(test_file, test_file)
            logger.info(f"File uploaded successfully! File ID: {file_id}")
        except AttributeError:
            logger.error("upload_file method not found in Python SDK")
            logger.info("Please implement file transfer in Python SDK first")
            return

        # Download file
        logger.info("\n[3/4] Downloading file...")
        await client.download_file(file_id, download_file)
        logger.info(f"File downloaded successfully: {download_file}")

        # Verify file
        logger.info("\n[4/4] Verifying file...")
        with open(download_file, 'r') as f:
            downloaded_content = f.read()

        if downloaded_content == test_content:
            logger.info("✓ File verification PASSED - Content matches!")
        else:
            logger.error("✗ File verification FAILED - Content mismatch!")

        logger.info("\n=== Demo complete ===")

    finally:
        # Cleanup
        if os.path.exists(test_file):
            os.remove(test_file)
        if os.path.exists(download_file):
            os.remove(download_file)
        await client.close()


if __name__ == '__main__':
    asyncio.run(main())
```

**Step 3: Create run.bat**

```bat
@echo off
echo Starting Python File Transfer Demo...
python main.py
pause
```

**Step 4: Test the example**

Run: `python main.py`
Expected: Creates, uploads, downloads, and verifies a test file
Note: May fail if Python SDK doesn't have upload/download methods

**Step 5: Commit**

```bash
git add light_link_platform/examples/file-transfer/python/file_transfer_demo/
git commit -m "feat(file-transfer): add Python example"
```

---

## Task 5: Update Parent Examples README

**Files:**
- Modify: `light_link_platform/examples/README.md`

**Step 1: Add file-transfer section**

```markdown
## Examples by Function

| 功能 | Go | C# | Python | 说明 |
|------|----|----|--------|------|
| Provider | ✅ | ✅ | ✅ | RPC 服务提供者 |
| Caller | ✅ | ✅ | ✅ | RPC 服务调用者 |
| Notify (PubSub) | ✅ | ✅ | ✅ | 发布订阅 |
| File Transfer | ✅ | ✅ | ⚠️ | 文件传输 |
| Backup | ✅ | 🔄 | 🔄 | 备份功能 |

*Note: Python file transfer requires SDK implementation*
```

**Step 2: Commit**

```bash
git add light_link_platform/examples/README.md
git commit -m "docs(examples): add file transfer to README"
```

---

## Testing Strategy

### Prerequisites
- NATS server with JetStream enabled
- TLS certificates in each example's `client/` folder

### Test Each Language

```bash
# Go
cd light_link_platform/examples/file-transfer/go/file-transfer-demo
go run main.go

# C#
cd light_link_platform/examples/file-transfer/csharp/FileTransferDemo
dotnet run

# Python
cd light_link_platform/examples/file-transfer/python/file_transfer_demo
python main.py
```

### Expected Output

All examples should:
1. Create a test file
2. Upload it to Object Store
3. Display the file ID
4. Download the file
5. Verify content matches

---

## Dependencies

### Go
- Go 1.21+
- LightLink Go SDK

### C#
- .NET 8.0
- LightLink C# SDK (with Client.cs from P0)

### Python
- Python 3.8+
- nats-py
- LightLink Python SDK
- **Note**: File transfer methods may need to be implemented in Python SDK first

---

## Related Plans

- P0: `2024-12-26-restore-csharp-client.md` (C# SDK Client required)
- P2 Python SDK State: May include file transfer implementation

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
2. Navigate to the Files section
3. Verify that:
   - Uploaded files appear in the file list
   - File metadata (size, type, timestamp) is displayed
   - Download functionality works

### Step 4: Test File Transfer Flow

1. Run a file transfer example (Go/C#/Python)
2. Observe in the management platform:
   - File upload progress
   - File appearing in Object Store
   - Successful download verification

### Step 5: Capture Evidence

Take screenshots showing the complete file transfer workflow in the management platform.
