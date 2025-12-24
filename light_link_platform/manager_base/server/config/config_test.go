package config

import (
	"os"
	"testing"
	"time"
)

func TestGetDefaultConfig(t *testing.T) {
	cfg := GetDefaultConfig()

	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Expected host '0.0.0.0', got '%s'", cfg.Server.Host)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", cfg.Server.Port)
	}

	if cfg.Database.Path != "data/light_link.db" {
		t.Errorf("Expected path 'data/light_link.db', got '%s'", cfg.Database.Path)
	}

	if cfg.JWT.Expiry != 24*time.Hour {
		t.Errorf("Expected expiry 24h, got %v", cfg.JWT.Expiry)
	}

	if cfg.Admin.Username != "admin" {
		t.Errorf("Expected username 'admin', got '%s'", cfg.Admin.Username)
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	f, err := os.CreateTemp("", "test_config_*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	cfg := &Config{
		Server: ServerConfig{
			Host: "127.0.0.1",
			Port: 9090,
		},
		Database: DatabaseConfig{
			Path: "test.db",
		},
		JWT: JWTConfig{
			Secret: "my-secret",
			Expiry: 1 * time.Hour,
		},
	}

	if err := cfg.Save(f.Name()); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(f.Name())
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Server.Host != "127.0.0.1" {
		t.Errorf("Expected host '127.0.0.1', got '%s'", loaded.Server.Host)
	}

	if loaded.Server.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", loaded.Server.Port)
	}

	if loaded.Database.Path != "test.db" {
		t.Errorf("Expected path 'test.db', got '%s'", loaded.Database.Path)
	}

	if loaded.JWT.Secret != "my-secret" {
		t.Errorf("Expected secret 'my-secret', got '%s'", loaded.JWT.Secret)
	}
}

func TestServerAddr(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: 3000,
		},
	}

	expected := "localhost:3000"
	if addr := cfg.ServerAddr(); addr != expected {
		t.Errorf("Expected '%s', got '%s'", expected, addr)
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	_, err := Load("non-existent.yaml")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestConfigDefaults(t *testing.T) {
	f, err := os.CreateTemp("", "test_config_defaults_*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	// Write minimal config
	f.WriteString("server:\n  port: 9000\n")
	f.Close()

	cfg, err := Load(f.Name())
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Check defaults
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Expected default host '0.0.0.0', got '%s'", cfg.Server.Host)
	}

	if cfg.Database.Path != "data/light_link.db" {
		t.Errorf("Expected default path, got '%s'", cfg.Database.Path)
	}

	if cfg.JWT.Expiry != 24*time.Hour {
		t.Errorf("Expected default expiry, got %v", cfg.JWT.Expiry)
	}
}
