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
        url = "nats://localhost:4222"
    }
    return &Config{
        NATSURL: url,
    }
}
