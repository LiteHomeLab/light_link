package main

import (
    "fmt"
    "time"

    "github.com/LiteHomeLab/light_link/sdk/go/client"
)

func main() {
    cli, _ := client.NewClient("nats://localhost:4222", nil)
    defer cli.Close()

    // Watch state changes
    stop, _ := cli.WatchState("device.sensor01", func(state map[string]interface{}) {
        fmt.Printf("State updated: %v\n", state)
    })
    defer stop()

    // Update state
    for i := 0; i < 3; i++ {
        cli.SetState("device.sensor01", map[string]interface{}{
            "temperature": 20.0 + float64(i),
            "humidity":    50 + i,
        })
        time.Sleep(500 * time.Millisecond)
    }

    // Get latest state
    state, _ := cli.GetState("device.sensor01")
    fmt.Printf("Latest state: %v\n", state)
}
