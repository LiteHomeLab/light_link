# LightLink Quick Start

## Install NATS Server

```bash
# Windows
nats-server -config deploy/nats/nats-server.conf
```

## Generate TLS Certificates

See `deploy/nats/tls/README.md`

## Client Usage

```go
import "github.com/LiteHomeLab/light_link/sdk/go/client"

cli, _ := client.NewClient("nats://localhost:4222", nil)
defer cli.Close()

// RPC call
result, _ := cli.Call("user-service", "getUser", map[string]interface{}{
    "user_id": "U001",
})

// Publish/Subscribe
cli.Publish("events.user", map[string]interface{}{"msg": "hello"})
cli.Subscribe("events.user", func(data map[string]interface{}) {
    fmt.Println(data)
})

// State management
cli.SetState("device.temp", map[string]interface{}{"value": 25.5})
state, _ := cli.GetState("device.temp")

// File transfer
fileID, _ := cli.UploadFile("./data.csv", "data.csv")
cli.DownloadFile(fileID, "./downloaded.csv")
```

## Service Usage

```go
import "github.com/LiteHomeLab/light_link/sdk/go/service"

svc, _ := service.NewService("my-service", "nats://localhost:4222", nil)

svc.RegisterRPC("add", func(args map[string]interface{}) (map[string]interface{}, error) {
    a := int(args["a"].(float64))
    b := int(args["b"].(float64))
    return map[string]interface{}{"sum": a + b}, nil
})

svc.Start()
```
