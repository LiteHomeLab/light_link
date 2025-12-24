package manager

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/LiteHomeLab/light_link/console/server/storage"
	"github.com/LiteHomeLab/light_link/sdk/go/types"
	"github.com/nats-io/nats.go"
)

// HeartbeatMonitor handles heartbeat messages and timeout detection
type HeartbeatMonitor struct {
	db         *storage.Database
	nc         *nats.Conn
	eventCh    chan *types.ServiceEvent
	sub        *nats.Subscription
	timeout    time.Duration
	lastSeen   map[string]time.Time
	mu         sync.RWMutex
	stopCh     chan struct{}
}

// NewHeartbeatMonitor creates a new heartbeat monitor
func NewHeartbeatMonitor(db *storage.Database, nc *nats.Conn, timeout time.Duration) *HeartbeatMonitor {
	return &HeartbeatMonitor{
		db:       db,
		nc:       nc,
		eventCh:  make(chan *types.ServiceEvent, 100),
		timeout:  timeout,
		lastSeen: make(map[string]time.Time),
		stopCh:   make(chan struct{}),
	}
}

// Subscribe subscribes to heartbeat messages
func (h *HeartbeatMonitor) Subscribe() error {
	sub, err := h.nc.Subscribe("$LL.heartbeat.>", h.handleHeartbeat)
	if err != nil {
		return err
	}
	h.sub = sub
	return nil
}

// Unsubscribe unsubscribes from heartbeat messages
func (h *HeartbeatMonitor) Unsubscribe() {
	if h.sub != nil {
		h.sub.Unsubscribe()
		h.sub = nil
	}
}

// StartChecker starts the timeout checker goroutine
func (h *HeartbeatMonitor) StartChecker() {
	go h.checkLoop()
}

// Stop stops the heartbeat monitor
func (h *HeartbeatMonitor) Stop() {
	close(h.stopCh)
}

// handleHeartbeat handles a heartbeat message
func (h *HeartbeatMonitor) handleHeartbeat(msg *nats.Msg) {
	log.Printf("[Heartbeat] Received heartbeat from subject: %s", msg.Subject)
	var heartbeat types.HeartbeatMessage
	if err := json.Unmarshal(msg.Data, &heartbeat); err != nil {
		log.Printf("[Heartbeat] Failed to unmarshal heartbeat: %v", err)
		return
	}
	log.Printf("[Heartbeat] Service: %s, Version: %s", heartbeat.Service, heartbeat.Version)

	// Update last seen time
	h.mu.Lock()
	_ = h.lastSeen[heartbeat.Service].IsZero() || time.Since(h.lastSeen[heartbeat.Service]) < h.timeout
	h.lastSeen[heartbeat.Service] = heartbeat.Timestamp
	h.mu.Unlock()

	// Update database status
	status, _ := h.db.GetServiceStatus(heartbeat.Service)
	h.db.UpdateServiceStatus(heartbeat.Service, true, heartbeat.Version)

	// If service was offline, send online event
	if status != nil && !status.Online {
		log.Printf("[Heartbeat] Service back online: %s", heartbeat.Service)
		h.eventCh <- &types.ServiceEvent{
			Type:      "online",
			Service:   heartbeat.Service,
			Timestamp: heartbeat.Timestamp,
		}
	}

	// Also save heartbeat event
	h.db.SaveEvent(&storage.ServiceEvent{
		Type:      "heartbeat",
		Service:   heartbeat.Service,
		Timestamp: heartbeat.Timestamp,
	})
}

// checkLoop runs the periodic timeout check
func (h *HeartbeatMonitor) checkLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.checkTimeouts()
		case <-h.stopCh:
			return
		}
	}
}

// checkTimeouts checks for services that have timed out
func (h *HeartbeatMonitor) checkTimeouts() {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()

	for service, lastSeen := range h.lastSeen {
		if now.Sub(lastSeen) > h.timeout {
			// Service has timed out
			log.Printf("[Heartbeat] Service timeout: %s (last seen %v ago)", service, now.Sub(lastSeen))

			// Update database status
			h.db.UpdateServiceStatus(service, false, "")

			// Send offline event
			h.eventCh <- &types.ServiceEvent{
				Type:      "offline",
				Service:   service,
				Timestamp: now,
			}

			// Save event to database
			h.db.SaveEvent(&storage.ServiceEvent{
				Type:      "offline",
				Service:   service,
				Timestamp: now,
			})

			// Remove from lastSeen map
			delete(h.lastSeen, service)
		}
	}
}

// Events returns the event channel
func (h *HeartbeatMonitor) Events() <-chan *types.ServiceEvent {
	return h.eventCh
}

// GetLastSeen returns the last seen time for a service
func (h *HeartbeatMonitor) GetLastSeen(service string) (time.Time, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	t, ok := h.lastSeen[service]
	return t, ok
}
