package manager

import (
	"encoding/json"
	"log"
	"time"

	"github.com/LiteHomeLab/light_link/light_link_platform/manager_base/server/storage"
	"github.com/LiteHomeLab/light_link/sdk/go/types"
	"github.com/nats-io/nats.go"
)

// Registry handles service registration messages
type Registry struct {
	db      *storage.Database
	nc      *nats.Conn
	eventCh chan *types.ServiceEvent
	sub    *nats.Subscription
}

// NewRegistry creates a new registry
func NewRegistry(db *storage.Database, nc *nats.Conn) *Registry {
	return &Registry{
		db:      db,
		nc:      nc,
		eventCh: make(chan *types.ServiceEvent, 100),
	}
}

// Subscribe subscribes to registration messages
func (r *Registry) Subscribe() error {
	sub, err := r.nc.Subscribe("$LL.register.>", r.handleRegister)
	if err != nil {
		return err
	}
	r.sub = sub
	return nil
}

// Unsubscribe unsubscribes from registration messages
func (r *Registry) Unsubscribe() {
	if r.sub != nil {
		r.sub.Unsubscribe()
		r.sub = nil
	}
}

// handleRegister handles a registration message
func (r *Registry) handleRegister(msg *nats.Msg) {
	var register types.RegisterMessage
	if err := json.Unmarshal(msg.Data, &register); err != nil {
		log.Printf("[Registry] Failed to unmarshal register message: %v", err)
		return
	}

	log.Printf("[Registry] Service registration: %s v%s", register.Service, register.Version)

	// Check if this is an update
	isUpdate := r.db.ServiceExists(register.Service)

	// Save service metadata
	meta := &storage.ServiceMetadata{
		Name:        register.Metadata.Name,
		Version:     register.Metadata.Version,
		Description: register.Metadata.Description,
		Author:      register.Metadata.Author,
		Tags:        register.Metadata.Tags,
	}

	if err := r.db.SaveService(meta); err != nil {
		log.Printf("[Registry] Failed to save service: %v", err)
		return
	}

	// Get service ID
	serviceID, err := r.db.GetServiceID(register.Metadata.Name)
	if err != nil {
		log.Printf("[Registry] Failed to get service ID: %v", err)
		return
	}

	// Save methods
	for _, m := range register.Metadata.Methods {
		methodMeta := &storage.MethodMetadata{
			ServiceID:   serviceID,
			Name:        m.Name,
			Description: m.Description,
			Params:      convertParams(m.Params),
			Returns:     convertReturns(m.Returns),
			Example:     convertExample(m.Example),
			Tags:        m.Tags,
			Deprecated:  m.Deprecated,
		}
		if err := r.db.SaveMethod(serviceID, methodMeta); err != nil {
			log.Printf("[Registry] Failed to save method %s: %v", m.Name, err)
		}
	}

	// Save or update instance record
	instanceKey := buildInstanceKey(register.InstanceInfo.HostIP, register.InstanceInfo.HostMAC, register.Metadata.Name)
	instance := &storage.Instance{
		ServiceName:   register.Metadata.Name,
		InstanceKey:   instanceKey,
		Language:      register.InstanceInfo.Language,
		HostIP:        register.InstanceInfo.HostIP,
		HostMAC:       register.InstanceInfo.HostMAC,
		WorkingDir:    register.InstanceInfo.WorkingDir,
		Version:       register.Version,
		FirstSeen:     time.Now(),
		LastHeartbeat: time.Now(),
		Online:        true,
	}
	if err := r.db.SaveInstance(instance); err != nil {
		log.Printf("[Registry] Failed to save instance: %v", err)
	}

	// Update status to online
	if err := r.db.UpdateServiceStatus(register.Metadata.Name, true, register.Version); err != nil {
		log.Printf("[Registry] Failed to update status: %v", err)
	}

	// Send event
	eventType := "registered"
	if isUpdate {
		eventType = "updated"
	}

	r.eventCh <- &types.ServiceEvent{
		Type:      eventType,
		Service:   register.Metadata.Name,
		Timestamp: time.Unix(register.Timestamp, 0),
	}
}

// Events returns the event channel
func (r *Registry) Events() <-chan *types.ServiceEvent {
	return r.eventCh
}

// Helper functions to convert between SDK and storage types

func convertParams(params []types.ParameterMetadata) []types.ParameterMetadata {
	// Same type, just copy
	result := make([]types.ParameterMetadata, len(params))
	copy(result, params)
	return result
}

func convertReturns(returns []types.ReturnMetadata) []types.ReturnMetadata {
	result := make([]types.ReturnMetadata, len(returns))
	copy(result, returns)
	return result
}

func convertExample(example *types.ExampleMetadata) *types.ExampleMetadata {
	if example == nil {
		return nil
	}
	return &types.ExampleMetadata{
		Input:       example.Input,
		Output:      example.Output,
		Description: example.Description,
	}
}

// normalizeMAC removes colons and dashes from MAC address for use as lock key
func normalizeMAC(mac string) string {
	result := make([]byte, 0, len(mac))
	for _, c := range mac {
		if c != ':' && c != '-' {
			result = append(result, byte(c))
		}
	}
	return string(result)
}

// buildInstanceKey builds the unique instance key from IP, MAC, and service name
func buildInstanceKey(ip, mac, serviceName string) string {
	return ip + ":" + normalizeMAC(mac) + ":" + serviceName
}
