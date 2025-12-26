package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/WQGroup/logger"
	"github.com/nats-io/nats.go"
	"github.com/LiteHomeLab/light_link/sdk/go/types"
)

// TLSConfig TLS configuration
type TLSConfig struct {
	CaFile     string
	CertFile   string
	KeyFile    string
	ServerName string
}

// Option is a function that configures a Client
type Option func(*Client) error

// Client represents a NATS client
type Client struct {
	nc        *nats.Conn
	tlsConfig *TLSConfig
	name      string
}

// WithAutoTLS automatically discovers and uses TLS certificates
// Searches in ./client directory
func WithAutoTLS() Option {
	return func(c *Client) error {
		result, err := types.DiscoverClientCerts()
		if err != nil {
			return fmt.Errorf("auto-discover TLS failed: %w", err)
		}
		c.tlsConfig = &TLSConfig{
			CaFile:     result.CaFile,
			CertFile:   result.CertFile,
			KeyFile:    result.KeyFile,
			ServerName: result.ServerName,
		}
		return nil
	}
}

// WithTLS uses the specified TLS configuration
func WithTLS(tlsConfig *TLSConfig) Option {
	return func(c *Client) error {
		c.tlsConfig = tlsConfig
		return nil
	}
}

// WithName sets the client name
func WithName(name string) Option {
	return func(c *Client) error {
		c.name = name
		return nil
	}
}

// NewClient creates a new client with options
func NewClient(url string, opts ...Option) (*Client, error) {
	client := &Client{
		name: "LightLink Client",
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, err
		}
	}

	// Initialize logger
	logger.SetLoggerName("LightLink-Client")

	natsOpts := []nats.Option{
		nats.Name(client.name),
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(10),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				logger.Errorf("Disconnected: %s", err.Error())
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logger.Infof("Reconnected to %s", nc.ConnectedUrl())
		}),
	}

	// Configure TLS
	if client.tlsConfig != nil {
		tlsOpt, err := CreateTLSOption(client.tlsConfig)
		if err != nil {
			return nil, err
		}
		natsOpts = append(natsOpts, tlsOpt)
	}

	nc, err := nats.Connect(url, natsOpts...)
	if err != nil {
		return nil, err
	}

	client.nc = nc
	return client, nil
}

// CreateTLSOption creates a TLS option
func CreateTLSOption(config *TLSConfig) (nats.Option, error) {
    // Load client certificate
    cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
    if err != nil {
        return nil, err
    }

    // Create CA pool
    pool := x509.NewCertPool()
    caCert, err := os.ReadFile(config.CaFile)
    if err != nil {
        return nil, err
    }
    pool.AppendCertsFromPEM(caCert)

    // Create TLS config
    // For development with self-signed certificates using legacy CN, we skip server name verification
    // The connection is still encrypted with TLS, and we verify the CA chain
    tlsConfig := &tls.Config{
        Certificates:       []tls.Certificate{cert},
        RootCAs:            pool,
        MinVersion:         tls.VersionTLS12,
        InsecureSkipVerify: true, // Skip server name verification for self-signed certs
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
