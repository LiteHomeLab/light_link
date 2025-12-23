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

    fmt.Println("=== State Management Demo ===")
    fmt.Println("NATS URL:", config.NATSURL)

    cli, err := client.NewClient(config.NATSURL, nil)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer cli.Close()

    fmt.Println("\n[1/3] Watching state changes for device.sensor01...")
    stop, err := cli.WatchState("device.sensor01", func(state map[string]interface{}) {
        fmt.Printf("State updated: %v\n", state)
    })
    if err != nil {
        log.Fatalf("Failed to watch state: %v", err)
    }
    defer stop()

    fmt.Println("\n[2/3] Updating state 3 times...")
    for i := 0; i < 3; i++ {
        err := cli.SetState("device.sensor01", map[string]interface{}{
            "temperature": 20.0 + float64(i),
            "humidity":    50 + i,
            "timestamp":   time.Now().Unix(),
        })
        if err != nil {
            fmt.Printf("Failed to set state: %v\n", err)
        }
        time.Sleep(500 * time.Millisecond)
    }

    fmt.Println("\n[3/3] Getting latest state...")
    state, err := cli.GetState("device.sensor01")
    if err != nil {
        fmt.Printf("Failed to get state: %v\n", err)
    } else {
        fmt.Printf("Latest state: %v\n", state)
    }

    fmt.Println("\n=== State Management Demo Complete ===")
}
