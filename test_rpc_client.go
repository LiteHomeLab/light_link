package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/LiteHomeLab/light_link/sdk/go/client"
	"github.com/LiteHomeLab/light_link/sdk/go/types"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run test_rpc_client.go <service-name> <method>")
		fmt.Println("Example: go run test_rpc_client.go math-service add")
		os.Exit(1)
	}

	serviceName := os.Args[1]
	methodName := os.Args[2]

	fmt.Printf("=== RPC Test Client ===\n")
	fmt.Printf("Service: %s\n", serviceName)
	fmt.Printf("Method: %s\n", methodName)
	fmt.Printf("Subject: $SRV.%s.%s\n\n", serviceName, methodName)

	// Change to light_link_platform directory for certificate discovery
	os.Chdir("light_link_platform")

	// Discover client certificates
	fmt.Println("[1/3] Discovering TLS certificates...")
	certResult, err := types.DiscoverClientCerts()
	if err != nil {
		log.Fatalf("Failed to discover certificates: %v", err)
	}
	fmt.Printf("CA:   %s\n", certResult.CaFile)
	fmt.Printf("Cert: %s\n", certResult.CertFile)

	// Create TLS config
	tlsConfig := &client.TLSConfig{
		CaFile:     certResult.CaFile,
		CertFile:   certResult.CertFile,
		KeyFile:    certResult.KeyFile,
		ServerName: certResult.ServerName,
	}

	// Connect to NATS
	fmt.Println("\n[2/3] Connecting to NATS...")
	natsURL := "nats://172.18.200.47:4222"
	cli, err := client.NewClient(natsURL, client.WithTLS(tlsConfig))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer cli.Close()
	nc := cli.GetNATSConn()
	fmt.Println("Connected successfully!")

	// Build RPC request
	fmt.Println("\n[3/3] Sending RPC request...")
	request := types.RPCRequest{
		ID:     fmt.Sprintf("%d", time.Now().UnixNano()),
		Method: methodName,
		Args:   map[string]interface{}{"a": 10.0, "b": 20.0},
	}

	requestData, err := json.Marshal(request)
	if err != nil {
		log.Fatalf("Failed to marshal request: %v", err)
	}

	subject := fmt.Sprintf("$SRV.%s.%s", serviceName, methodName)
	fmt.Printf("Subject: %s\n", subject)
	fmt.Printf("Request: %s\n\n", string(requestData))

	// Send request with timeout
	respMsg, err := nc.Request(subject, requestData, 5*time.Second)
	if err != nil {
		fmt.Printf("❌ ERROR: %v\n", err)
		fmt.Printf("\nThis error means: No service is listening on subject '%s'\n", subject)
		fmt.Printf("\nPossible causes:\n")
		fmt.Printf("  1. Service is not running\n")
		fmt.Printf("  2. Service name mismatch (expected: %s)\n", serviceName)
		fmt.Printf("  3. Method '%s' not registered\n", methodName)
		os.Exit(1)
	}

	// Parse response
	var response types.RPCResponse
	if err := json.Unmarshal(respMsg.Data, &response); err != nil {
		log.Fatalf("Failed to parse response: %v", err)
	}

	// Display result
	if response.Success {
		fmt.Printf("✅ SUCCESS!\n")
		fmt.Printf("Result: %+v\n", response.Result)
	} else {
		fmt.Printf("❌ FAILED!\n")
		fmt.Printf("Error: %s\n", response.Error)
		os.Exit(1)
	}
}
