package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/LiteHomeLab/light_link/sdk/go/types"
)

// RegisterMetadata registers service metadata and sends it to NATS
func (s *Service) RegisterMetadata(metadata *types.ServiceMetadata) error {
	s.metaMutex.Lock()
	defer s.metaMutex.Unlock()

	// Store metadata
	s.metadata = metadata

	// Send registration message to NATS with instance info
	msg := types.RegisterMessage{
		Service:   s.name,
		Version:   metadata.Version,
		Metadata:  *metadata,
		Timestamp: time.Now().Unix(),
		InstanceInfo: types.InstanceInfo{
			Language:   "go",
			HostIP:     s.hostInfo.IP,
			HostMAC:    s.hostInfo.MAC,
			WorkingDir: s.hostInfo.WorkingDir,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	subject := fmt.Sprintf("$LL.register.%s", s.name)
	if err := s.nc.Publish(subject, data); err != nil {
		return fmt.Errorf("publish metadata: %w", err)
	}

	return nil
}

// GetMetadata returns the service metadata
func (s *Service) GetMetadata() *types.ServiceMetadata {
	s.metaMutex.RLock()
	defer s.metaMutex.RUnlock()
	return s.metadata
}

// RegisterMethodWithMetadata registers a method with its metadata
func (s *Service) RegisterMethodWithMetadata(
	name string,
	handler RPCHandler,
	metadata *types.MethodMetadata,
) error {
	// Store method metadata
	s.metaMutex.Lock()
	if s.methodsMeta == nil {
		s.methodsMeta = make(map[string]*types.MethodMetadata)
	}
	s.methodsMeta[name] = metadata
	s.metaMutex.Unlock()

	// Register the RPC handler
	return s.RegisterRPC(name, handler)
}

// GetMethodMetadata returns the metadata for a method
func (s *Service) GetMethodMetadata(name string) (*types.MethodMetadata, bool) {
	s.metaMutex.RLock()
	defer s.metaMutex.RUnlock()
	if s.methodsMeta == nil {
		return nil, false
	}
	meta, ok := s.methodsMeta[name]
	return meta, ok
}

// ListMethodMetadata returns all method metadata
func (s *Service) ListMethodMetadata() map[string]*types.MethodMetadata {
	s.metaMutex.RLock()
	defer s.metaMutex.RUnlock()
	if s.methodsMeta == nil {
		return make(map[string]*types.MethodMetadata)
	}
	result := make(map[string]*types.MethodMetadata, len(s.methodsMeta))
	for k, v := range s.methodsMeta {
		result[k] = v
	}
	return result
}

// UpdateMetadata updates and re-sends the service metadata
func (s *Service) UpdateMetadata(metadata *types.ServiceMetadata) error {
	s.metaMutex.Lock()
	// Keep existing methods if not provided
	if len(metadata.Methods) == 0 && s.metadata != nil {
		metadata.Methods = s.metadata.Methods
	}
	s.metaMutex.Unlock()

	return s.RegisterMetadata(metadata)
}

// BuildCurrentMetadata builds metadata from current registered methods
func (s *Service) BuildCurrentMetadata(name, version, description, author string, tags []string) *types.ServiceMetadata {
	s.metaMutex.RLock()
	defer s.metaMutex.RUnlock()

	methods := make([]types.MethodMetadata, 0, len(s.methodsMeta))
	for _, meta := range s.methodsMeta {
		methods = append(methods, *meta)
	}

	return &types.ServiceMetadata{
		Name:        name,
		Version:     version,
		Description: description,
		Author:      author,
		Tags:        tags,
		Methods:     methods,
		RegisteredAt: time.Now(),
		LastSeen:    time.Now(),
	}
}
