package client

import (
    "encoding/json"
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/LiteHomeLab/light_link/sdk/go/types"
)

// Call makes a synchronous RPC call
func (c *Client) Call(service, method string, args map[string]interface{}) (map[string]interface{}, error) {
    return c.CallWithTimeout(service, method, args, 5*time.Second)
}

// CallWithTimeout makes an RPC call with timeout
func (c *Client) CallWithTimeout(service, method string, args map[string]interface{}, timeout time.Duration) (map[string]interface{}, error) {
    // Generate request ID
    requestID := uuid.New().String()

    // Build request
    request := types.RPCRequest{
        ID:     requestID,
        Method: method,
        Args:   args,
    }

    reqData, err := json.Marshal(request)
    if err != nil {
        return nil, fmt.Errorf("marshal request: %w", err)
    }

    // Build subject
    subject := fmt.Sprintf("$SRV.%s.%s", service, method)

    // Send request and wait for response
    msg, err := c.nc.Request(subject, reqData, timeout)
    if err != nil {
        return nil, fmt.Errorf("RPC request failed: %w", err)
    }

    // Parse response
    var response types.RPCResponse
    if err := json.Unmarshal(msg.Data, &response); err != nil {
        return nil, fmt.Errorf("unmarshal response: %w", err)
    }

    if !response.Success {
        return nil, fmt.Errorf("RPC error: %s", response.Error)
    }

    return response.Result, nil
}
