package main

import (
    "fmt"
    "log"
    "time"

    "github.com/LiteHomeLab/light_link/sdk/go/client"
    "github.com/LiteHomeLab/light_link/examples"
)

func main() {
    config := examples.GetConfig()

    fmt.Println("=== Pub/Sub Demo ===")
    fmt.Println("NATS URL:", config.NATSURL)

    cli, err := client.NewClient(config.NATSURL, nil)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer cli.Close()

    fmt.Println("\n[1/2] Subscribing to events.user...")
    sub, err := cli.Subscribe("events.user", func(data map[string]interface{}) {
        fmt.Printf("Received: %v\n", data)
    })
    if err != nil {
        log.Fatalf("Failed to subscribe: %v", err)
    }
    defer sub.Unsubscribe()

    fmt.Println("\n[2/2] Publishing 5 messages...")
    for i := 0; i < 5; i++ {
        err := cli.Publish("events.user", map[string]interface{}{
            "event":     "user_login",
            "user_id":   fmt.Sprintf("U%03d", i),
            "timestamp": time.Now().Unix(),
        })
        if err != nil {
            fmt.Printf("Failed to publish: %v\n", err)
        }
        time.Sleep(500 * time.Millisecond)
    }

    fmt.Println("\nWaiting for messages to complete...")
    time.Sleep(1 * time.Second)

    fmt.Println("=== Pub/Sub Demo Complete ===")
}
