package main

import (
    "fmt"

    "github.com/LiteHomeLab/light_link/sdk/go/client"
)

func main() {
    cli, _ := client.NewClient("nats://localhost:4222", nil)
    defer cli.Close()

    // Upload file
    fileID, err := cli.UploadFile("./test.txt", "test.txt")
    if err != nil {
        fmt.Println("Upload error:", err)
        return
    }
    fmt.Println("Uploaded, file ID:", fileID)

    // Download file
    err = cli.DownloadFile(fileID, "./downloaded.txt")
    if err != nil {
        fmt.Println("Download error:", err)
        return
    }
    fmt.Println("Downloaded successfully")
}
