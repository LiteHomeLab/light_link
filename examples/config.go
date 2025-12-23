package examples

import "os"

// Config holds example configuration
type Config struct {
    NATSURL string
}

// GetConfig returns the configuration from environment or default values
func GetConfig() *Config {
    url := os.Getenv("NATS_URL")
    if url == "" {
        url = "nats://172.18.200.47:4222"
    }
    return &Config{
        NATSURL: url,
    }
}
