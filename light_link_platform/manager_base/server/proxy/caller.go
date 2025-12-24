package proxy

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/LiteHomeLab/light_link/sdk/go/types"
	"github.com/nats-io/nats.go"
)

// Caller handles RPC calls to services
type Caller struct {
	nc      *nats.Conn
	timeout time.Duration
}

// CallResult represents the result of an RPC call
type CallResult struct {
	Success    bool                   `json:"success"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Duration   int64                  `json:"duration"`
}

// NewCaller creates a new RPC caller
func NewCaller(nc *nats.Conn, timeout time.Duration) *Caller {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &Caller{
		nc:      nc,
		timeout: timeout,
	}
}

// Call calls an RPC method on a service
func (c *Caller) Call(serviceName, methodName string, params map[string]interface{}) (*CallResult, error) {
	start := time.Now()

	// Build RPC request
	request := types.RPCRequest{
		ID:     generateID(),
		Method: methodName,
		Args:   params,
	}

	requestData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	// Send request to service
	subject := fmt.Sprintf("$SRV.%s.%s", serviceName, methodName)
	respMsg, err := c.nc.Request(subject, requestData, c.timeout)
	if err != nil {
		return &CallResult{
			Success:  false,
			Error:    err.Error(),
			Duration: time.Since(start).Milliseconds(),
		}, nil
	}

	// Parse response
	var response types.RPCResponse
	if err := json.Unmarshal(respMsg.Data, &response); err != nil {
		return &CallResult{
			Success:  false,
			Error:    fmt.Sprintf("parse response: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}, nil
	}

	// Handle validation error with structured details
	errorMsg := response.Error
	if !response.Success && response.Result != nil {
		// Check if this is a structured validation error
		if errType, ok := response.Result["type"].(string); ok && errType == types.ValidationErrorType {
			// Build more detailed error message
			if paramName, ok := response.Result["parameter_name"].(string); ok {
				expectedType, _ := response.Result["expected_type"].(string)
				actualType, _ := response.Result["actual_type"].(string)
				errorMsg = fmt.Sprintf("参数 '%s' 类型错误: 期望 %s，实际 %s",
					paramName, expectedType, actualType)
			}
		}
	}

	return &CallResult{
		Success:  response.Success,
		Data:     response.Result,
		Error:    errorMsg,
		Duration: time.Since(start).Milliseconds(),
	}, nil
}

// CallWithTimeout calls an RPC method with a custom timeout
func (c *Caller) CallWithTimeout(serviceName, methodName string, params map[string]interface{}, timeout time.Duration) (*CallResult, error) {
	start := time.Now()

	request := types.RPCRequest{
		ID:     generateID(),
		Method: methodName,
		Args:   params,
	}

	requestData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	subject := fmt.Sprintf("$SRV.%s.%s", serviceName, methodName)
	respMsg, err := c.nc.Request(subject, requestData, timeout)
	if err != nil {
		return &CallResult{
			Success:  false,
			Error:    err.Error(),
			Duration: time.Since(start).Milliseconds(),
		}, nil
	}

	var response types.RPCResponse
	if err := json.Unmarshal(respMsg.Data, &response); err != nil {
		return &CallResult{
			Success:  false,
			Error:    fmt.Sprintf("parse response: %v", err),
			Duration: time.Since(start).Milliseconds(),
		}, nil
	}

	return &CallResult{
		Success:  response.Success,
		Data:     response.Result,
		Error:    response.Error,
		Duration: time.Since(start).Milliseconds(),
	}, nil
}

// CallAsync calls an RPC method asynchronously
func (c *Caller) CallAsync(serviceName, methodName string, params map[string]interface{}, callback func(*CallResult)) {
	go func() {
		result, err := c.Call(serviceName, methodName, params)
		if err != nil {
			log.Printf("[Caller] Async call failed: %v", err)
			callback(&CallResult{Success: false, Error: err.Error()})
			return
		}
		callback(result)
	}()
}

// generateID generates a unique request ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
