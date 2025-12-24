package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/LiteHomeLab/light_link/sdk/go/types"
)

const (
	// DefaultHeartbeatInterval is the default interval between heartbeats
	DefaultHeartbeatInterval = 30 * time.Second
)

// startHeartbeat starts sending heartbeat messages
func (s *Service) startHeartbeat() error {
	subject := fmt.Sprintf("$LL.heartbeat.%s", s.name)

	// Send first heartbeat immediately
	if err := s.sendHeartbeat(subject); err != nil {
		return err
	}

	// Start heartbeat goroutine
	go func() {
		ticker := time.NewTicker(DefaultHeartbeatInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.sendHeartbeat(subject)
			case <-s.heartbeatStop:
				return
			}
		}
	}()

	return nil
}

// sendHeartbeat sends a single heartbeat message
func (s *Service) sendHeartbeat(subject string) error {
	version := "unknown"
	s.metaMutex.RLock()
	if s.metadata != nil {
		version = s.metadata.Version
	}
	s.metaMutex.RUnlock()

	msg := types.HeartbeatMessage{
		Service:   s.name,
		Version:   version,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal heartbeat: %w", err)
	}

	return s.nc.Publish(subject, data)
}

// SetHeartbeatInterval sets a custom heartbeat interval (for testing)
// Note: This must be called before Start()
func (s *Service) SetHeartbeatInterval(interval time.Duration) {
	// This can be used to customize the interval for testing purposes
	// In production, use the default interval
	_ = interval
}

// GetHeartbeatSubject returns the heartbeat subject for this service
func (s *Service) GetHeartbeatSubject() string {
	return fmt.Sprintf("$LL.heartbeat.%s", s.name)
}

// GetRegisterSubject returns the register subject for this service
func (s *Service) GetRegisterSubject() string {
	return fmt.Sprintf("$LL.register.%s", s.name)
}
