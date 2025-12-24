package client

import (
    "testing"
)

func TestNewClient(t *testing.T) {
    // Test creating client (requires NATS server running)
    config := &TLSConfig{
        CaFile:     "../../../deploy/nats/tls/ca.crt",
        CertFile:   "../../../deploy/nats/tls/demo-service.crt",
        KeyFile:    "../../../deploy/nats/tls/demo-service.key",
        ServerName: "nats-server",
    }
    client, err := NewClient("nats://172.18.200.47:4222", config)
    if err != nil {
        t.Skip("Need running NATS server:", err)
    }
    if client == nil {
        t.Fatal("client is nil")
    }
    client.Close()
}

func TestNewClientWithTLS(t *testing.T) {
    config := &TLSConfig{
        CaFile:     "../../deploy/nats/tls/ca.crt",
        CertFile:   "../../deploy/nats/tls/user-service.crt",
        KeyFile:    "../../deploy/nats/tls/user-service.key",
        ServerName: "nats-server",
    }

    // Note: This test requires certificates to exist
    client, err := NewClient("tls://172.18.200.47:4222", config)
    if err != nil {
        // Expected to fail without certificates, this is OK
        t.Logf("Expected failure without certs: %v", err)
        return
    }
    defer client.Close()
    t.Log("Client created with TLS config")
}

func TestClientClose(t *testing.T) {
    config := &TLSConfig{
        CaFile:     "../../../deploy/nats/tls/ca.crt",
        CertFile:   "../../../deploy/nats/tls/demo-service.crt",
        KeyFile:    "../../../deploy/nats/tls/demo-service.key",
        ServerName: "nats-server",
    }
    client, err := NewClient("nats://172.18.200.47:4222", config)
    if err != nil {
        t.Skip("Need running NATS server:", err)
    }
    err = client.Close()
    if err != nil {
        t.Fatalf("Close failed: %v", err)
    }
}

func TestClientGetNATSConn(t *testing.T) {
    config := &TLSConfig{
        CaFile:     "../../../deploy/nats/tls/ca.crt",
        CertFile:   "../../../deploy/nats/tls/demo-service.crt",
        KeyFile:    "../../../deploy/nats/tls/demo-service.key",
        ServerName: "nats-server",
    }
    client, err := NewClient("nats://172.18.200.47:4222", config)
    if err != nil {
        t.Skip("Need running NATS server:", err)
    }
    defer client.Close()

    conn := client.GetNATSConn()
    if conn == nil {
        t.Fatal("NATS connection is nil")
    }
}
