package main

import (
    "fmt"
    "log"
    "time"

    "github.com/LiteHomeLab/light_link/sdk/go/client"
    "github.com/LiteHomeLab/light_link/sdk/go/service"
    "github.com/LiteHomeLab/light_link/examples"
)

func main() {
    config := examples.GetConfig()

    fmt.Println("=== RPC Demo ===")
    fmt.Println("NATS URL:", config.NATSURL)

    // Start service
    fmt.Println("\n[1/2] Starting service...")
    svc, err := service.NewService("demo-service", config.NATSURL, nil)
    if err != nil {
        log.Fatalf("Failed to create service: %v", err)
    }

    err = svc.RegisterRPC("add", func(args map[string]interface{}) (map[string]interface{}, error) {
        a := int(args["a"].(float64))
        b := int(args["b"].(float64))
        return map[string]interface{}{"sum": a + b}, nil
    })
    if err != nil {
        log.Fatalf("Failed to register RPC: %v", err)
    }

    err = svc.RegisterRPC("multiply", func(args map[string]interface{}) (map[string]interface{}, error) {
        a := int(args["a"].(float64))
        b := int(args["b"].(float64))
        return map[string]interface{}{"product": a * b}, nil
    })
    if err != nil {
        log.Fatalf("Failed to register RPC: %v", err)
    }

    err = svc.Start()
    if err != nil {
        log.Fatalf("Failed to start service: %v", err)
    }
    defer svc.Stop()
    fmt.Println("Service started successfully!")

    // Wait for service to be ready
    time.Sleep(500 * time.Millisecond)

    // Client call
    fmt.Println("\n[2/2] Testing RPC calls...")
    cli, err := client.NewClient(config.NATSURL, nil)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer cli.Close()

    // Test add
    fmt.Println("\nTest 1: Add (10 + 20)")
    result, err := cli.Call("demo-service", "add", map[string]interface{}{
        "a": 10,
        "b": 20,
    })
    if err != nil {
        fmt.Println("Error:", err)
    } else {
        fmt.Println("Result:", result)
    }

    // Test multiply
    fmt.Println("\nTest 2: Multiply (5 * 6)")
    result, err = cli.Call("demo-service", "multiply", map[string]interface{}{
        "a": 5,
        "b": 6,
    })
    if err != nil {
        fmt.Println("Error:", err)
    } else {
        fmt.Println("Result:", result)
    }

    fmt.Println("\n=== RPC Demo Complete ===")
}
