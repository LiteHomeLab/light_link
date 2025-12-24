package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the console server configuration
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	NATS      NATSConfig      `yaml:"nats"`
	Database  DatabaseConfig  `yaml:"database"`
	JWT       JWTConfig       `yaml:"jwt"`
	Heartbeat HeartbeatConfig `yaml:"heartbeat"`
	Admin     AdminConfig     `yaml:"admin"`
}

// ServerConfig represents the server configuration
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// NATSConfig represents the NATS configuration
type NATSConfig struct {
	URL string    `yaml:"url"`
	TLS TLSConfig `yaml:"tls"`
}

// TLSConfig represents the TLS configuration
type TLSConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Cert       string `yaml:"cert"`
	Key        string `yaml:"key"`
	CA         string `yaml:"ca"`
	ServerName string `yaml:"server_name,omitempty"`
}

// DatabaseConfig represents the database configuration
type DatabaseConfig struct {
	Path string `yaml:"path"`
}

// JWTConfig represents the JWT configuration
type JWTConfig struct {
	Secret string        `yaml:"secret"`
	Expiry time.Duration `yaml:"expiry"`
}

// HeartbeatConfig represents the heartbeat configuration
type HeartbeatConfig struct {
	Interval time.Duration `yaml:"interval"`
	Timeout  time.Duration `yaml:"timeout"`
}

// AdminConfig represents the default admin configuration
type AdminConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// Load loads the configuration from a file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Set defaults
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Database.Path == "" {
		cfg.Database.Path = "data/light_link.db"
	}
	if cfg.JWT.Expiry == 0 {
		cfg.JWT.Expiry = 24 * time.Hour
	}
	if cfg.Heartbeat.Timeout == 0 {
		cfg.Heartbeat.Timeout = 90 * time.Second
	}
	if cfg.Admin.Username == "" {
		cfg.Admin.Username = "admin"
	}
	if cfg.Admin.Password == "" {
		cfg.Admin.Password = "admin123"
	}

	return &cfg, nil
}

// ServerAddr returns the server address
func (c *Config) ServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// GetDefaultConfig returns a default configuration
func GetDefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		NATS: NATSConfig{
			URL: "nats://172.18.200.47:4222",
			TLS: TLSConfig{
				Enabled:    false,
				CA:         "tls/ca.crt",
				Cert:       "tls/manager.crt",
				Key:        "tls/manager.key",
				ServerName: "nats-server",
			},
		},
		Database: DatabaseConfig{
			Path: "data/light_link.db",
		},
		JWT: JWTConfig{
			Secret: "change-this-secret-in-production",
			Expiry: 24 * time.Hour,
		},
		Heartbeat: HeartbeatConfig{
			Timeout: 90 * time.Second,
		},
		Admin: AdminConfig{
			Username: "admin",
			Password: "admin123",
		},
	}
}

// Save saves the configuration to a file
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}
