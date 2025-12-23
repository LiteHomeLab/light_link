package main

import (
    "fmt"
    "time"

    "github.com/LiteHomeLab/light_link/sdk/go/client"
    "github.com/LiteHomeLab/light_link/sdk/go/service"
)

func main() {
    // Remote NATS server URL
    natsURL := "nats://172.18.200.47:4222"

    fmt.Println("Testing connection to remote NATS server:", natsURL)

    // Start service
    fmt.Println("\n[1/3] Starting service...")
    svc, err := service.NewService("test-service", natsURL, nil)
    if err != nil {
        fmt.Println("Failed to create service:", err)
        return
    }
    defer svc.Stop()

    err = svc.RegisterRPC("ping", func(args map[string]interface{}) (map[string]interface{}, error) {
        return map[string]interface{}{"message": "pong"}, nil
    })
    if err != nil {
        fmt.Println("Failed to register RPC:", err)
        return
    }

    err = svc.RegisterRPC("add", func(args map[string]interface{}) (map[string]interface{}, error) {
        a := int(args["a"].(float64))
        b := int(args["b"].(float64))
        return map[string]interface{}{"sum": a + b}, nil
    })
    if err != nil {
        fmt.Println("Failed to register RPC:", err)
        return
    }

    err = svc.Start()
    if err != nil {
        fmt.Println("Failed to start service:", err)
        return
    }
    fmt.Println("Service started successfully!")

    // Wait for service to be ready
    time.Sleep(500 * time.Millisecond)

    // Client call
    fmt.Println("\n[2/3] Testing RPC calls...")
    cli, err := client.NewClient(natsURL, nil)
    if err != nil {
        fmt.Println("Failed to create client:", err)
        return
    }
    defer cli.Close()
    fmt.Println("Client connected successfully!")

    // Test ping
    fmt.Println("\nTest 1: Ping")
    result, err := cli.Call("test-service", "ping", map[string]interface{}{})
    if err != nil {
        fmt.Println("Error:", err)
    } else {
        fmt.Println("Result:", result)
    }

    // Test add
    fmt.Println("\nTest 2: Add (10 + 20)")
    result, err = cli.Call("test-service", "add", map[string]interface{}{
        "a": 10,
        "b": 20,
    })
    if err != nil {
        fmt.Println("Error:", err)
    } else {
        fmt.Println("Result:", result)
    }

    // Test pub/sub
    fmt.Println("\n[3/3] Testing pub/sub...")
    subscription, err := cli.Subscribe("test.events", func(msg map[string]interface{}) {
        fmt.Println("Received message:", msg)
    })
    if err != nil {
        fmt.Println("Failed to subscribe:", err)
        return
    }
    defer subscription.Unsubscribe()
    fmt.Println("Subscribed to test.events")

    // Publish messages
    for i := 1; i <= 3; i++ {
        err := cli.Publish("test.events", map[string]interface{}{
            "id":      i,
            "message": fmt.Sprintf("Hello #%d", i),
        })
        if err != nil {
            fmt.Println("Failed to publish:", err)
        }
    }

    // Wait for messages to be received
    time.Sleep(500 * time.Millisecond)

    fmt.Println("\n=== All tests completed ===")
}
