package main

import (
    "fmt"
    "time"

    "github.com/LiteHomeLab/light_link/sdk/go/client"
    "github.com/LiteHomeLab/light_link/sdk/go/service"
)

func main() {
    // Start service
    svc, _ := service.NewService("demo-service", "nats://localhost:4222", nil)

    svc.RegisterRPC("add", func(args map[string]interface{}) (map[string]interface{}, error) {
        a := int(args["a"].(float64))
        b := int(args["b"].(float64))
        return map[string]interface{}{"sum": a + b}, nil
    })

    svc.Start()
    defer svc.Stop()

    // Wait for service to be ready
    time.Sleep(100 * time.Millisecond)

    // Client call
    cli, _ := client.NewClient("nats://localhost:4222", nil)
    defer cli.Close()

    result, err := cli.Call("demo-service", "add", map[string]interface{}{
        "a": 10,
        "b": 20,
    })

    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    fmt.Println("Result:", result)
}
