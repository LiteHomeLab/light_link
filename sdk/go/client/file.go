package client

import (
    "context"
    "io"
    "os"

    "github.com/google/uuid"
    "github.com/nats-io/nats.go/jetstream"
)

// UploadFile uploads file to Object Store
func (c *Client) UploadFile(filePath, fileName string) (string, error) {
    js, err := jetstream.New(c.nc)
    if err != nil {
        return "", err
    }

    // Get or create Object Store
    store, err := js.ObjectStore(context.Background(), "light_link_files")
    if err != nil {
        store, err = js.CreateObjectStore(context.Background(), jetstream.ObjectStoreConfig{
            Bucket: "light_link_files",
        })
        if err != nil {
            return "", err
        }
    }

    // Read file
    data, err := os.ReadFile(filePath)
    if err != nil {
        return "", err
    }

    // Generate file ID as object name
    fileID := uuid.New().String()

    // Upload to Object Store using PutBytes
    _, err = store.PutBytes(context.Background(), fileID, data)
    if err != nil {
        return "", err
    }

    return fileID, nil
}

// DownloadFile downloads file from Object Store
func (c *Client) DownloadFile(fileID, destPath string) error {
    js, err := jetstream.New(c.nc)
    if err != nil {
        return err
    }

    store, err := js.ObjectStore(context.Background(), "light_link_files")
    if err != nil {
        return err
    }

    // Get file
    result, err := store.Get(context.Background(), fileID)
    if err != nil {
        return err
    }
    defer result.Close()

    // Read all data
    data, err := io.ReadAll(result)
    if err != nil {
        return err
    }

    return os.WriteFile(destPath, data, 0644)
}

// SendFile sends file to service (upload + notification)
func (c *Client) SendFile(filePath, fileName, targetService string) error {
    fileID, err := c.UploadFile(filePath, fileName)
    if err != nil {
        return err
    }

    // Send file metadata notification
    metadata := map[string]interface{}{
        "file_id":   fileID,
        "file_name": fileName,
        "to":        targetService,
    }

    return c.Publish("file.transfer", metadata)
}
