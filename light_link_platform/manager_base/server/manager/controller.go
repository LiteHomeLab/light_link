package manager

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/LiteHomeLab/light_link/light_link_platform/manager_base/server/storage"
	"github.com/LiteHomeLab/light_link/sdk/go/types"
	"github.com/nats-io/nats.go"
)

// Controller handles remote control of service instances
type Controller struct {
	db *storage.Database
	nc *nats.Conn
}

// NewController creates a new controller
func NewController(manager *Manager) *Controller {
	return &Controller{
		db: manager.db,
		nc: manager.nc,
	}
}

// StopInstance stops a specific service instance
func (c *Controller) StopInstance(instanceKey string) error {
	// Verify instance exists and is online
	instance, err := c.db.GetInstance(instanceKey)
	if err != nil {
		return fmt.Errorf("instance not found: %w", err)
	}

	if !instance.Online {
		return fmt.Errorf("instance is not online")
	}

	// Send stop command
	controlMsg := types.ControlMessage{
		Service:     instance.ServiceName,
		InstanceKey: instance.InstanceKey,
		Command:     "stop",
		Timestamp:   time.Now().Unix(),
	}

	return c.sendControlMessage(&controlMsg)
}

// RestartInstance restarts a specific service instance
func (c *Controller) RestartInstance(instanceKey string) error {
	// Verify instance exists and is online
	instance, err := c.db.GetInstance(instanceKey)
	if err != nil {
		return fmt.Errorf("instance not found: %w", err)
	}

	if !instance.Online {
		return fmt.Errorf("instance is not online")
	}

	// Send restart command
	controlMsg := types.ControlMessage{
		Service:     instance.ServiceName,
		InstanceKey: instance.InstanceKey,
		Command:     "restart",
		Timestamp:   time.Now().Unix(),
	}

	return c.sendControlMessage(&controlMsg)
}

// StopServiceInstances stops all instances of a service
func (c *Controller) StopServiceInstances(serviceName string) (int, error) {
	instances, err := c.db.GetInstancesByService(serviceName)
	if err != nil {
		return 0, fmt.Errorf("failed to get instances: %w", err)
	}

	stopped := 0
	for _, instance := range instances {
		if instance.Online {
			if err := c.StopInstance(instance.InstanceKey); err == nil {
				stopped++
			} else {
				log.Printf("[Controller] Failed to stop instance %s: %v", instance.InstanceKey, err)
			}
		}
	}

	return stopped, nil
}

// RestartServiceInstances restarts all instances of a service
func (c *Controller) RestartServiceInstances(serviceName string) (int, error) {
	instances, err := c.db.GetInstancesByService(serviceName)
	if err != nil {
		return 0, fmt.Errorf("failed to get instances: %w", err)
	}

	restarted := 0
	for _, instance := range instances {
		if instance.Online {
			if err := c.RestartInstance(instance.InstanceKey); err == nil {
				restarted++
			} else {
				log.Printf("[Controller] Failed to restart instance %s: %v", instance.InstanceKey, err)
			}
		}
	}

	return restarted, nil
}

// ListInstances returns all instances for a service
func (c *Controller) ListInstances(serviceName string) ([]*storage.Instance, error) {
	return c.db.GetInstancesByService(serviceName)
}

// ListAllInstances returns all instances
func (c *Controller) ListAllInstances() ([]*storage.Instance, error) {
	return c.db.ListAllInstances()
}

// GetInstance returns a specific instance
func (c *Controller) GetInstance(instanceKey string) (*storage.Instance, error) {
	return c.db.GetInstance(instanceKey)
}

// DeleteOfflineInstance deletes an offline instance from the database
func (c *Controller) DeleteOfflineInstance(instanceKey string) error {
	instance, err := c.db.GetInstance(instanceKey)
	if err != nil {
		return fmt.Errorf("instance not found: %w", err)
	}

	if instance.Online {
		return fmt.Errorf("cannot delete online instance")
	}

	return c.db.DeleteInstance(instanceKey)
}

// sendControlMessage sends a control message via NATS
func (c *Controller) sendControlMessage(msg *types.ControlMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal control message: %w", err)
	}

	subject := fmt.Sprintf("$LL.control.%s", msg.Service)
	log.Printf("[Controller] Sending %s command to %s (instance: %s)",
		msg.Command, subject, msg.InstanceKey)

	if err := c.nc.Publish(subject, data); err != nil {
		return fmt.Errorf("publish control message: %w", err)
	}

	return nil
}
