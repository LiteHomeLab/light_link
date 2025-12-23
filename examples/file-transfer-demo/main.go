package main

import (
    "fmt"
    "log"
    "os"

    "github.com/LiteHomeLab/light_link/sdk/go/client"
    "github.com/LiteHomeLab/light_link/examples"
)

func main() {
    config := examples.GetConfig()

    fmt.Println("=== File Transfer Demo ===")
    fmt.Println("NATS URL:", config.NATSURL)

    cli, err := client.NewClient(config.NATSURL, nil)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer cli.Close()

    // Create test file
    testFile := "test.txt"
    content := []byte("Hello, LightLink! This is a test file for file transfer.\n")
    err = os.WriteFile(testFile, content, 0644)
    if err != nil {
        log.Fatalf("Failed to create test file: %v", err)
    }
    defer os.Remove(testFile)

    fmt.Println("\n[1/2] Uploading file...")
    fileID, err := cli.UploadFile(testFile, testFile)
    if err != nil {
        fmt.Println("Upload error:", err)
        return
    }
    fmt.Printf("Uploaded successfully, file ID: %s\n", fileID)

    fmt.Println("\n[2/2] Downloading file...")
    downloadFile := "downloaded.txt"
    defer os.Remove(downloadFile)

    err = cli.DownloadFile(fileID, downloadFile)
    if err != nil {
        fmt.Println("Download error:", err)
        return
    }
    fmt.Println("Downloaded successfully")

    // Verify content
    downloadedContent, err := os.ReadFile(downloadFile)
    if err != nil {
        fmt.Printf("Failed to read downloaded file: %v\n", err)
        return
    }

    if string(downloadedContent) == string(content) {
        fmt.Println("Content verification: PASSED")
    } else {
        fmt.Println("Content verification: FAILED")
    }

    fmt.Println("\n=== File Transfer Demo Complete ===")
}
