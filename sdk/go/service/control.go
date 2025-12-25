package service

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/LiteHomeLab/light_link/sdk/go/types"
	"github.com/nats-io/nats.go"
)

// ControlHandler handles control messages from the management platform
type ControlHandler struct {
	nc          *nats.Conn
	serviceName string
	instanceKey string
	sub         *nats.Subscription
	mu          sync.RWMutex
	running     bool
}

// NewControlHandler creates a new control handler
func NewControlHandler(nc *nats.Conn, serviceName, instanceKey string) *ControlHandler {
	return &ControlHandler{
		nc:          nc,
		serviceName: serviceName,
		instanceKey: instanceKey,
	}
}

// Subscribe subscribes to control messages for this service instance
func (c *ControlHandler) Subscribe() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return fmt.Errorf("control handler already running")
	}

	// Subscribe to service-specific control messages
	subject := fmt.Sprintf("$LL.control.%s.>", c.serviceName)
	sub, err := c.nc.Subscribe(subject, c.handleControl)
	if err != nil {
		return fmt.Errorf("failed to subscribe to control messages: %w", err)
	}

	c.sub = sub
	c.running = true

	log.Printf("[Control] Subscribed to %s", subject)
	return nil
}

// Unsubscribe unsubscribes from control messages
func (c *ControlHandler) Unsubscribe() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sub != nil {
		c.sub.Unsubscribe()
		c.sub = nil
	}
	c.running = false
}

// handleControl handles a control message
func (c *ControlHandler) handleControl(msg *nats.Msg) {
	log.Printf("[Control] Received message on subject: %s", msg.Subject)
	log.Printf("[Control] Message data: %s", string(msg.Data))

	var control types.ControlMessage
	if err := json.Unmarshal(msg.Data, &control); err != nil {
		log.Printf("[Control] Failed to unmarshal control message: %v", err)
		return
	}

	log.Printf("[Control] Received command: %s for instance: %s (my key: %s)", control.Command, control.InstanceKey, c.instanceKey)

	// Check if this message is for this instance
	if control.InstanceKey != c.instanceKey {
		log.Printf("[Control] Message not for this instance (expected %s), ignoring", c.instanceKey)
		return
	}

	// Handle the command
	switch control.Command {
	case "stop":
		log.Printf("[Control] Stopping service...")
		c.Unsubscribe()
		os.Exit(0)

	case "restart":
		log.Printf("[Control] Restarting service...")
		c.Unsubscribe()
		// Exit code 99 indicates restart
		os.Exit(99)

	default:
		log.Printf("[Control] Unknown command: %s", control.Command)
	}
}

// IsRunning returns whether the control handler is running
func (c *ControlHandler) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running
}
