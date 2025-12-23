package client

import (
    "io/ioutil"
    "os"
    "testing"
)

func TestFileTransfer(t *testing.T) {
    client, err := NewClient("nats://localhost:4222", nil)
    if err != nil {
        t.Skip("Need running NATS server with JetStream:", err)
    }
    defer client.Close()

    // Create test file
    tmpFile, err := ioutil.TempFile("", "test-*.txt")
    if err != nil {
        t.Fatal(err)
    }
    defer os.Remove(tmpFile.Name())

    testData := []byte("Hello, LightLink File Transfer!")
    tmpFile.Write(testData)
    tmpFile.Close()

    // Upload file
    fileID, err := client.UploadFile(tmpFile.Name(), "test.txt")
    if err != nil {
        t.Fatalf("UploadFile failed: %v", err)
    }

    // Download file
    outFile, err := ioutil.TempFile("", "download-*.txt")
    if err != nil {
        t.Fatal(err)
    }
    defer os.Remove(outFile.Name())
    outFile.Close()

    err = client.DownloadFile(fileID, outFile.Name())
    if err != nil {
        t.Fatalf("DownloadFile failed: %v", err)
    }

    // Verify content
    downloaded, _ := ioutil.ReadFile(outFile.Name())
    if string(downloaded) != string(testData) {
        t.Error("File content mismatch")
    }
}
