package manager

import (
	"log"
	"sync"
	"time"

	"github.com/LiteHomeLab/light_link/console/server/storage"
	"github.com/LiteHomeLab/light_link/sdk/go/types"
	"github.com/nats-io/nats.go"
)

// Manager manages service registration, heartbeat monitoring, and event handling
type Manager struct {
	db       *storage.Database
	nc       *nats.Conn
	registry *Registry
	monitor  *HeartbeatMonitor
	eventCh  chan *types.ServiceEvent
	mu       sync.RWMutex
	stopCh   chan struct{}
}

// NewManager creates a new service manager
func NewManager(db *storage.Database, nc *nats.Conn, heartbeatTimeout time.Duration) *Manager {
	return &Manager{
		db:      db,
		nc:      nc,
		eventCh: make(chan *types.ServiceEvent, 100),
		stopCh:  make(chan struct{}),
	}
}

// Start starts the service manager
func (m *Manager) Start() error {
	log.Println("[Manager] Starting service manager...")

	// Create registry
	m.registry = NewRegistry(m.db, m.nc)
	if err := m.registry.Subscribe(); err != nil {
		return err
	}
	log.Println("[Manager] Registry subscribed to $LL.register.>")

	// Create heartbeat monitor
	m.monitor = NewHeartbeatMonitor(m.db, m.nc, 90*time.Second)
	if err := m.monitor.Subscribe(); err != nil {
		m.registry.Unsubscribe()
		return err
	}
	m.monitor.StartChecker()
	log.Println("[Manager] Heartbeat monitor subscribed to $LL.heartbeat.>")

	// Start event forwarding
	go m.forwardEvents()

	log.Println("[Manager] Service manager started")
	return nil
}

// Stop stops the service manager
func (m *Manager) Stop() {
	log.Println("[Manager] Stopping service manager...")
	close(m.stopCh)

	if m.registry != nil {
		m.registry.Unsubscribe()
	}

	if m.monitor != nil {
		m.monitor.Stop()
	}

	log.Println("[Manager] Service manager stopped")
}

// forwardEvents forwards events from registry and monitor to the main event channel
func (m *Manager) forwardEvents() {
	for {
		select {
		case e := <-m.registry.Events():
			// Save to database
			m.db.SaveEvent(&storage.ServiceEvent{
				Type:      e.Type,
				Service:   e.Service,
				Timestamp: e.Timestamp,
			})
			// Forward to main channel
			select {
			case m.eventCh <- e:
			case <-m.stopCh:
				return
			}

		case e := <-m.monitor.Events():
			// Save to database
			m.db.SaveEvent(&storage.ServiceEvent{
				Type:      e.Type,
				Service:   e.Service,
				Timestamp: e.Timestamp,
			})
			// Forward to main channel
			select {
			case m.eventCh <- e:
			case <-m.stopCh:
				return
			}

		case <-m.stopCh:
			return
		}
	}
}

// Events returns the event channel for receiving service events
func (m *Manager) Events() <-chan *types.ServiceEvent {
	return m.eventCh
}

// GetDatabase returns the database instance
func (m *Manager) GetDatabase() *storage.Database {
	return m.db
}

// GetNATSConn returns the NATS connection
func (m *Manager) GetNATSConn() *nats.Conn {
	return m.nc
}

// GetRegistry returns the registry instance
func (m *Manager) GetRegistry() *Registry {
	return m.registry
}

// GetMonitor returns the heartbeat monitor instance
func (m *Manager) GetMonitor() *HeartbeatMonitor {
	return m.monitor
}

// ListServices returns all registered services
func (m *Manager) ListServices() ([]*storage.ServiceMetadata, error) {
	return m.db.ListServices()
}

// GetService returns a specific service
func (m *Manager) GetService(name string) (*storage.ServiceMetadata, error) {
	return m.db.GetService(name)
}

// ListServiceStatus returns all service statuses
func (m *Manager) ListServiceStatus() ([]*storage.ServiceStatus, error) {
	return m.db.ListServiceStatus()
}

// GetServiceStatus returns the status of a specific service
func (m *Manager) GetServiceStatus(name string) (*storage.ServiceStatus, error) {
	return m.db.GetServiceStatus(name)
}

// ListEvents returns events with pagination
func (m *Manager) ListEvents(limit, offset int) ([]*storage.ServiceEvent, error) {
	return m.db.ListEvents(limit, offset)
}

// GetOnlineServices returns a list of online services
func (m *Manager) GetOnlineServices() ([]string, error) {
	return m.db.GetOnlineServices()
}

// GetOfflineServices returns a list of offline services
func (m *Manager) GetOfflineServices() ([]string, error) {
	return m.db.GetOfflineServices()
}

// GetServiceMethods returns all methods for a service
func (m *Manager) GetServiceMethods(serviceName string) ([]*storage.MethodMetadata, error) {
	return m.db.GetMethods(serviceName)
}

// GetServiceMethod returns a specific method
func (m *Manager) GetServiceMethod(serviceName, methodName string) (*storage.MethodMetadata, error) {
	return m.db.GetMethod(serviceName, methodName)
}
