package service

import (
    "encoding/json"
    "fmt"
    "sync"
    "time"

    "github.com/nats-io/nats.go"
    "github.com/LiteHomeLab/light_link/sdk/go/types"
)

// RPCHandler RPC handler function type
type RPCHandler func(args map[string]interface{}) (map[string]interface{}, error)

// Service represents a service
type Service struct {
    name     string
    nc       *nats.Conn
    rpcMap   map[string]RPCHandler
    rpcMutex sync.RWMutex
    running  bool
}

// NewService creates a new service
func NewService(name, natsURL string, tlsConfig interface{}) (*Service, error) {
    nc, err := nats.Connect(natsURL,
        nats.Name("LightLink Service: "+name),
        nats.ReconnectWait(2*time.Second),
        nats.MaxReconnects(10),
    )
    if err != nil {
        return nil, err
    }

    return &Service{
        name:   name,
        nc:     nc,
        rpcMap: make(map[string]RPCHandler),
    }, nil
}

// Name returns the service name
func (s *Service) Name() string {
    return s.name
}

// RegisterRPC registers an RPC method
func (s *Service) RegisterRPC(method string, handler RPCHandler) error {
    s.rpcMutex.Lock()
    defer s.rpcMutex.Unlock()

    s.rpcMap[method] = handler
    return nil
}

// HasRPC checks if an RPC method is registered
func (s *Service) HasRPC(method string) bool {
    s.rpcMutex.RLock()
    defer s.rpcMutex.RUnlock()

    _, exists := s.rpcMap[method]
    return exists
}

// Start starts the service
func (s *Service) Start() error {
    if s.running {
        return fmt.Errorf("service already running")
    }

    // Subscribe to all RPC methods
    subject := fmt.Sprintf("$SRV.%s.>", s.name)
    _, err := s.nc.Subscribe(subject, s.handleRPC)
    if err != nil {
        return fmt.Errorf("subscribe failed: %w", err)
    }

    s.running = true
    return nil
}

// handleRPC handles RPC requests
func (s *Service) handleRPC(msg *nats.Msg) {
    // Parse request
    var request types.RPCRequest
    if err := json.Unmarshal(msg.Data, &request); err != nil {
        s.sendError(msg, "", "invalid request: "+err.Error())
        return
    }

    // Find handler
    s.rpcMutex.RLock()
    handler, exists := s.rpcMap[request.Method]
    s.rpcMutex.RUnlock()

    if !exists {
        s.sendError(msg, request.ID, "method not found: "+request.Method)
        return
    }

    // Call handler
    result, err := handler(request.Args)
    if err != nil {
        s.sendError(msg, request.ID, err.Error())
        return
    }

    // Send response
    response := types.RPCResponse{
        ID:      request.ID,
        Success: true,
        Result:  result,
    }

    respData, _ := json.Marshal(response)
    msg.Respond(respData)
}

// sendError sends an error response
func (s *Service) sendError(msg *nats.Msg, requestID, errMsg string) {
    response := types.RPCResponse{
        ID:      requestID,
        Success: false,
        Error:   errMsg,
    }
    respData, _ := json.Marshal(response)
    msg.Respond(respData)
}

// Stop stops the service
func (s *Service) Stop() error {
    if !s.running {
        return nil
    }

    s.nc.Close()
    s.running = false
    return nil
}
