package main

import (
    "fmt"
    "time"

    "github.com/LiteHomeLab/light_link/sdk/go/client"
)

func main() {
    cli, _ := client.NewClient("nats://localhost:4222", nil)
    defer cli.Close()

    // Subscribe
    cli.Subscribe("events.user", func(data map[string]interface{}) {
        fmt.Printf("Received: %v\n", data)
    })

    // Publish
    for i := 0; i < 5; i++ {
        cli.Publish("events.user", map[string]interface{}{
            "event":   "user_login",
            "user_id": fmt.Sprintf("U%03d", i),
        })
        time.Sleep(500 * time.Millisecond)
    }

    time.Sleep(1 * time.Second)
}
