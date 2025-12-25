package service

import (
    "encoding/json"
    "fmt"
    "sync"
    "time"

    "github.com/nats-io/nats.go"
    "github.com/LiteHomeLab/light_link/sdk/go/client"
    "github.com/LiteHomeLab/light_link/sdk/go/types"
)

// RPCHandler RPC handler function type
type RPCHandler func(args map[string]interface{}) (map[string]interface{}, error)

// ServiceOption is a function that configures a Service
type ServiceOption func(*Service) error

// Service represents a service
type Service struct {
	name          string
	nc            *nats.Conn
	tlsConfig     *client.TLSConfig
	rpcMap        map[string]RPCHandler
	rpcMutex      sync.RWMutex
	metadata      *types.ServiceMetadata
	metaMutex     sync.RWMutex
	methodsMeta   map[string]*types.MethodMetadata
	running       bool
	heartbeatStop chan struct{}
}

// WithServiceAutoTLS automatically discovers and uses server TLS certificates
// Searches in ./nats-server directory
func WithServiceAutoTLS() ServiceOption {
	return func(s *Service) error {
		result, err := types.DiscoverServerCerts()
		if err != nil {
			return fmt.Errorf("auto-discover server TLS failed: %w", err)
		}
		s.tlsConfig = &client.TLSConfig{
			CaFile:     result.CaFile,
			CertFile:   result.CertFile,
			KeyFile:    result.KeyFile,
			ServerName: result.ServerName,
		}
		return nil
	}
}

// WithServiceTLS uses the specified TLS configuration
func WithServiceTLS(tlsConfig *client.TLSConfig) ServiceOption {
	return func(s *Service) error {
		s.tlsConfig = tlsConfig
		return nil
	}
}

// NewService creates a new service with options
func NewService(name, natsURL string, opts ...ServiceOption) (*Service, error) {
	service := &Service{
		name:          name,
		rpcMap:        make(map[string]RPCHandler),
		methodsMeta:   make(map[string]*types.MethodMetadata),
		heartbeatStop: make(chan struct{}),
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(service); err != nil {
			return nil, err
		}
	}

	natsOpts := []nats.Option{
		nats.Name("LightLink Service: "+name),
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(10),
	}

	// Configure TLS
	if service.tlsConfig != nil {
		tlsOpt, err := client.CreateTLSOption(service.tlsConfig)
		if err != nil {
			return nil, err
		}
		natsOpts = append(natsOpts, tlsOpt)
	}

	nc, err := nats.Connect(natsURL, natsOpts...)
	if err != nil {
		return nil, err
	}

	service.nc = nc
	return service, nil
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

    // Start heartbeat
    if err := s.startHeartbeat(); err != nil {
        return fmt.Errorf("start heartbeat: %w", err)
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

    // Get method metadata for validation
    s.metaMutex.RLock()
    methodMeta, hasMeta := s.methodsMeta[request.Method]
    s.metaMutex.RUnlock()

    // Validate parameters if metadata exists
    if hasMeta {
        validator := NewValidator(methodMeta)
        if err := validator.Validate(request.Args); err != nil {
            // Handle ValidationError specifically
            if validationErr, ok := err.(*types.ValidationError); ok {
                s.sendValidationError(msg, request.ID, validationErr)
                return
            }
            s.sendError(msg, request.ID, err.Error())
            return
        }
    }

    // Call handler with panic recovery for type assertion errors
    result, err := s.callHandlerSafely(handler, request.Args, request.Method, hasMeta, methodMeta)
    if err != nil {
        // Check if it's a validation error from panic recovery
        if validationErr, ok := err.(*types.ValidationError); ok {
            s.sendValidationError(msg, request.ID, validationErr)
            return
        }
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

// sendValidationError sends a validation error response with structured error info
func (s *Service) sendValidationError(msg *nats.Msg, requestID string, err *types.ValidationError) {
    // Build structured error response
    errorDetail := map[string]interface{}{
        "type":           types.ValidationErrorType,
        "parameter_name": err.ParameterName,
        "expected_type":  err.ExpectedType,
        "actual_type":    err.ActualType,
        "message":        err.Message,
    }
    if err.ActualValue != nil {
        errorDetail["actual_value"] = err.ActualValue
    }

    response := types.RPCResponse{
        ID:      requestID,
        Success: false,
        Error:   err.Message,
        Result:  errorDetail,
    }

    respData, _ := json.Marshal(response)
    msg.Respond(respData)
}

// callHandlerSafely calls the handler with panic recovery
func (s *Service) callHandlerSafely(
    handler RPCHandler,
    args map[string]interface{},
    methodName string,
    hasMeta bool,
    methodMeta *types.MethodMetadata,
) (result map[string]interface{}, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = s.convertPanicToValidationError(r, args, methodName, hasMeta, methodMeta)
        }
    }()

    result, err = handler(args)
    return
}

// convertPanicToValidationError converts a panic to ValidationError
func (s *Service) convertPanicToValidationError(
    panicValue interface{},
    args map[string]interface{},
    methodName string,
    hasMeta bool,
    methodMeta *types.MethodMetadata,
) *types.ValidationError {
    // Try to parse the panic to find which parameter failed
    if hasMeta && methodMeta != nil {
        for _, param := range methodMeta.Params {
            if value, exists := args[param.Name]; exists {
                actualType := inferTypeString(value)
                if !isTypeCompatible(param.Type, actualType) {
                    return &types.ValidationError{
                        ParameterName: param.Name,
                        ExpectedType:  param.Type,
                        ActualType:    actualType,
                        ActualValue:   value,
                        Message:       fmt.Sprintf("parameter '%s': expected type %s, got %s",
                            param.Name, param.Type, actualType),
                    }
                }
            }
        }
    }

    // Fallback: generic error with panic info
    return &types.ValidationError{
        Message: fmt.Sprintf("parameter type mismatch: %v", panicValue),
    }
}

// Stop stops the service
func (s *Service) Stop() error {
    if !s.running {
        return nil
    }

    // Stop heartbeat
    close(s.heartbeatStop)
    s.heartbeatStop = make(chan struct{})

    s.nc.Close()
    s.running = false
    return nil
}
