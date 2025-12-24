package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LiteHomeLab/light_link/console/server/api"
	"github.com/LiteHomeLab/light_link/console/server/auth"
	"github.com/LiteHomeLab/light_link/console/server/config"
	"github.com/LiteHomeLab/light_link/console/server/manager"
	"github.com/LiteHomeLab/light_link/console/server/proxy"
	"github.com/LiteHomeLab/light_link/console/server/storage"
	"github.com/LiteHomeLab/light_link/console/server/ws"
	"github.com/nats-io/nats.go"
)

func main() {
	// Load configuration
	configPath := "console.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Printf("Failed to load config from %s, using defaults: %v", configPath, err)
		cfg = config.GetDefaultConfig()
	}

	// Create data directory
	if err := os.MkdirAll("data", 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Connect to NATS
	nc, err := connectNATS(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()
	log.Println("Connected to NATS:", cfg.NATS.URL)

	// Initialize database
	db, err := storage.NewDatabase(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Println("Database:", cfg.Database.Path)

	// Initialize admin user
	if err := db.InitAdminUser(cfg.Admin.Username, cfg.Admin.Password); err != nil {
		log.Printf("Warning: Failed to init admin user: %v", err)
	} else {
		log.Printf("Admin user: %s (please change password after first login)", cfg.Admin.Username)
	}

	// Start service manager
	mgr := manager.NewManager(db, nc, cfg.Heartbeat.Timeout)
	if err := mgr.Start(); err != nil {
		log.Fatalf("Failed to start manager: %v", err)
	}
	log.Println("Service manager started")

	// Create auth middleware
	authMW := auth.NewAuthMiddleware(cfg.JWT.Secret, cfg.JWT.Expiry)

	// Create WebSocket hub
	hub := ws.NewHub()
	go hub.Run()
	log.Println("WebSocket hub started")

	// Connect manager events to hub
	go func() {
		for event := range mgr.Events() {
			hub.Events() <- event
		}
	}()

	// Create RPC caller
	rpcCaller := proxy.NewCaller(nc, 30*time.Second)

	// Create API handler
	apiHandler := api.NewHandler(db, mgr, authMW)
	go runServer(cfg, apiHandler, hub, mgr, rpcCaller)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")

	// Stop manager
	mgr.Stop()

	// Give server time to shutdown gracefully
	time.Sleep(1 * time.Second)
	log.Println("Server stopped")
}

// connectNATS connects to NATS with optional TLS
func connectNATS(cfg *config.Config) (*nats.Conn, error) {
	opts := []nats.Option{
		nats.Name("LightLink Console"),
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(-1), // Infinite reconnects
	}

	if cfg.NATS.TLS.Enabled {
		// Configure TLS with client certificate
		if cfg.NATS.TLS.ServerName != "" {
			opts = append(opts, nats.TLSClientAuth(cfg.NATS.TLS.CA, cfg.NATS.TLS.Cert, cfg.NATS.TLS.Key))
			// Create custom TLS config with ServerName
			tlsConfig := &tls.Config{
				ServerName: cfg.NATS.TLS.ServerName,
				MinVersion: tls.VersionTLS12,
			}
			opts = append(opts, nats.Secure(tlsConfig))
		} else {
			opts = append(opts,
				nats.RootCAs(cfg.NATS.TLS.CA),
				nats.ClientCert(cfg.NATS.TLS.Cert, cfg.NATS.TLS.Key),
			)
		}
	}

	return nats.Connect(cfg.NATS.URL, opts...)
}

// runServer starts the HTTP server
func runServer(cfg *config.Config, apiHandler *api.Handler, hub *ws.Hub, mgr *manager.Manager, caller *proxy.Caller) {
	// Wrap API handler with WebSocket support
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// WebSocket endpoint
		if r.URL.Path == "/api/ws" {
			// TODO: Add auth for WebSocket
			hub.HandleConnection(w, r)
			return
		}

		// Pass to API handler
		apiHandler.Routes().ServeHTTP(w, r)
	})

	server := &http.Server{
		Addr:    cfg.ServerAddr(),
		Handler: handler,
	}

	log.Printf("Server starting on %s", cfg.ServerAddr())
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("Server error: %v", err)
	}
}
