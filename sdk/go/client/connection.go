package client

import (
    "crypto/tls"
    "crypto/x509"
    "io/ioutil"
    "time"

    "github.com/nats-io/nats.go"
)

// TLSConfig TLS configuration
type TLSConfig struct {
    CaFile   string
    CertFile string
    KeyFile  string
}

// Client represents a NATS client
type Client struct {
    nc *nats.Conn
}

// NewClient creates a new client
func NewClient(url string, tlsConfig *TLSConfig) (*Client, error) {
    opts := []nats.Option{
        nats.Name("LightLink Client"),
        nats.ReconnectWait(2 * time.Second),
        nats.MaxReconnects(10),
        nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
            if err != nil {
                println("Disconnected:", err.Error())
            }
        }),
        nats.ReconnectHandler(func(nc *nats.Conn) {
            println("Reconnected to", nc.ConnectedUrl())
        }),
    }

    // Configure TLS
    if tlsConfig != nil {
        tlsOpt, err := createTLSOption(tlsConfig)
        if err != nil {
            return nil, err
        }
        opts = append(opts, tlsOpt)
    }

    nc, err := nats.Connect(url, opts...)
    if err != nil {
        return nil, err
    }

    return &Client{nc: nc}, nil
}

// createTLSOption creates a TLS option
func createTLSOption(config *TLSConfig) (nats.Option, error) {
    // Load client certificate
    cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
    if err != nil {
        return nil, err
    }

    // Create CA pool
    pool := x509.NewCertPool()
    caCert, err := ioutil.ReadFile(config.CaFile)
    if err != nil {
        return nil, err
    }
    pool.AppendCertsFromPEM(caCert)

    // Create TLS config
    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
        RootCAs:      pool,
        MinVersion:   tls.VersionTLS12,
    }

    return nats.Secure(tlsConfig), nil
}

// GetNATSConn returns the NATS connection
func (c *Client) GetNATSConn() *nats.Conn {
    return c.nc
}

// Close closes the client
func (c *Client) Close() error {
    if c.nc != nil {
        c.nc.Close()
    }
    return nil
}
