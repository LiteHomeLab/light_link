package client

import (
    "os"
    "testing"
)

func TestFileTransfer(t *testing.T) {
    config := &TLSConfig{
        CaFile:     "../../../deploy/nats/tls/ca.crt",
        CertFile:   "../../../deploy/nats/tls/demo-service.crt",
        KeyFile:    "../../../deploy/nats/tls/demo-service.key",
        ServerName: "nats-server",
    }
    c, err := NewClient("nats://172.18.200.47:4222", WithTLS(config))
    if err != nil {
        t.Skip("Need running NATS server with JetStream:", err)
    }
    defer c.Close()

    // Create test file
    tmpFile, err := os.CreateTemp("", "test-*.txt")
    if err != nil {
        t.Fatal(err)
    }
    defer os.Remove(tmpFile.Name())

    testData := []byte("Hello, LightLink File Transfer!")
    tmpFile.Write(testData)
    tmpFile.Close()

    // Upload file
    fileID, err := c.UploadFile(tmpFile.Name(), "test.txt")
    if err != nil {
        t.Fatalf("UploadFile failed: %v", err)
    }

    // Download file
    outFile, err := os.CreateTemp("", "download-*.txt")
    if err != nil {
        t.Fatal(err)
    }
    defer os.Remove(outFile.Name())
    outFile.Close()

    err = c.DownloadFile(fileID, outFile.Name())
    if err != nil {
        t.Fatalf("DownloadFile failed: %v", err)
    }

    // Verify content
    downloaded, _ := os.ReadFile(outFile.Name())
    if string(downloaded) != string(testData) {
        t.Error("File content mismatch")
    }
}
