package main

import (
	"fmt"
	"log"

	"github.com/LiteHomeLab/light_link/sdk/go/service"
	"github.com/LiteHomeLab/light_link/sdk/go/types"
	"github.com/LiteHomeLab/light_link/examples"
)

func main() {
	config := examples.GetConfig()

	fmt.Println("=== Metadata Registration Demo ===")
	fmt.Println("NATS URL:", config.NATSURL)

	// Auto-discover client certificates for TLS
	fmt.Println("\n[1/4] Discovering TLS certificates...")
	certResult, err := types.DiscoverClientCerts()
	if err != nil {
		log.Fatalf("Failed to discover client certificates: %v", err)
	}
	fmt.Printf("Certificates found:\n")
	fmt.Printf("  CA:   %s\n", certResult.CaFile)
	fmt.Printf("  Cert: %s\n", certResult.CertFile)
	fmt.Printf("  Key:  %s\n", certResult.KeyFile)

	tlsConfig := &types.TLSConfig{
		CaFile:     certResult.CaFile,
		CertFile:   certResult.CertFile,
		KeyFile:    certResult.KeyFile,
		ServerName: certResult.ServerName,
	}

	// Create service with TLS
	fmt.Println("\n[2/5] Creating service...")
	svc, err := service.NewService("math-service", config.NATSURL, service.WithServiceTLS(tlsConfig))
	if err != nil {
		log.Fatalf("Failed to create service: %v", err)
	}
	defer svc.Stop()
	fmt.Println("Service created successfully!")

	// Define method metadata for 'add'
	addMeta := &types.MethodMetadata{
		Name:        "add",
		Description: "Add two numbers together",
		Params: []types.ParameterMetadata{
			{
				Name:        "a",
				Type:        "number",
				Required:    true,
				Description: "First number",
			},
			{
				Name:        "b",
				Type:        "number",
				Required:    true,
				Description: "Second number",
			},
		},
		Returns: []types.ReturnMetadata{
			{
				Name:        "sum",
				Type:        "number",
				Description: "The sum of a and b",
			},
		},
		Example: &types.ExampleMetadata{
			Input:       map[string]any{"a": 10, "b": 20},
			Output:      map[string]any{"sum": 30},
			Description: "10 + 20 = 30",
		},
		Tags: []string{"math", "basic", "arithmetic"},
	}

	// Define method metadata for 'multiply'
	multiplyMeta := &types.MethodMetadata{
		Name:        "multiply",
		Description: "Multiply two numbers",
		Params: []types.ParameterMetadata{
			{
				Name:        "a",
				Type:        "number",
				Required:    true,
				Description: "First factor",
			},
			{
				Name:        "b",
				Type:        "number",
				Required:    true,
				Description: "Second factor",
			},
		},
		Returns: []types.ReturnMetadata{
			{
				Name:        "product",
				Type:        "number",
				Description: "The product of a and b",
			},
		},
		Example: &types.ExampleMetadata{
			Input:       map[string]any{"a": 5, "b": 6},
			Output:      map[string]any{"product": 30},
			Description: "5 * 6 = 30",
		},
		Tags: []string{"math", "basic", "arithmetic"},
	}

	// Define method metadata for 'power'
	powerMeta := &types.MethodMetadata{
		Name:        "power",
		Description: "Calculate a to the power of b",
		Params: []types.ParameterMetadata{
			{
				Name:        "base",
				Type:        "number",
				Required:    true,
				Description: "The base number",
			},
			{
				Name:        "exp",
				Type:        "number",
				Required:    true,
				Description: "The exponent",
			},
		},
		Returns: []types.ReturnMetadata{
			{
				Name:        "result",
				Type:        "number",
				Description: "base raised to the power of exp",
			},
		},
		Example: &types.ExampleMetadata{
			Input:       map[string]any{"base": 2, "exp": 10},
			Output:      map[string]any{"result": 1024},
			Description: "2^10 = 1024",
		},
		Tags: []string{"math", "advanced"},
	}

	// Define method metadata for 'divide'
	divideMeta := &types.MethodMetadata{
		Name:        "divide",
		Description: "Divide two numbers",
		Params: []types.ParameterMetadata{
			{
				Name:        "numerator",
				Type:        "number",
				Required:    true,
				Description: "The number to be divided",
			},
			{
				Name:        "denominator",
				Type:        "number",
				Required:    true,
				Description: "The number to divide by",
			},
		},
		Returns: []types.ReturnMetadata{
			{
				Name:        "quotient",
				Type:        "number",
				Description: "The result of division",
			},
		},
		Example: &types.ExampleMetadata{
			Input:       map[string]any{"numerator": 100, "denominator": 4},
			Output:      map[string]any{"quotient": 25},
			Description: "100 / 4 = 25",
		},
		Tags: []string{"math", "basic", "arithmetic"},
	}

	// Register methods with metadata
	fmt.Println("\n[3/5] Registering methods with metadata...")

	if err := svc.RegisterMethodWithMetadata("add", addHandler, addMeta); err != nil {
		log.Fatalf("Failed to register add: %v", err)
	}
	fmt.Println("  - add: registered")

	if err := svc.RegisterMethodWithMetadata("multiply", multiplyHandler, multiplyMeta); err != nil {
		log.Fatalf("Failed to register multiply: %v", err)
	}
	fmt.Println("  - multiply: registered")

	if err := svc.RegisterMethodWithMetadata("power", powerHandler, powerMeta); err != nil {
		log.Fatalf("Failed to register power: %v", err)
	}
	fmt.Println("  - power: registered")

	if err := svc.RegisterMethodWithMetadata("divide", divideHandler, divideMeta); err != nil {
		log.Fatalf("Failed to register divide: %v", err)
	}
	fmt.Println("  - divide: registered")

	// Build and register service metadata
	fmt.Println("\n[4/5] Registering service metadata...")
	metadata := svc.BuildCurrentMetadata(
		"math-service",
		"v1.0.0",
		"A mathematical operations service providing basic and advanced math functions",
		"LiteHomeLab",
		[]string{"demo", "math", "calculator"},
	)

	if err := svc.RegisterMetadata(metadata); err != nil {
		log.Fatalf("Failed to register metadata: %v", err)
	}
	fmt.Println("Service metadata registered to NATS!")
	fmt.Printf("  Service: %s\n", metadata.Name)
	fmt.Printf("  Version: %s\n", metadata.Version)
	fmt.Printf("  Methods: %d\n", len(metadata.Methods))

	// Start service
	fmt.Println("\n[5/5] Starting service...")
	if err := svc.Start(); err != nil {
		log.Fatalf("Failed to start service: %v", err)
	}
	fmt.Println("Service started successfully!")

	fmt.Println("\n=== Service Information ===")
	fmt.Printf("Service Name: %s\n", svc.Name())
	fmt.Printf("Registered Methods:\n")
	for name, meta := range svc.ListMethodMetadata() {
		fmt.Printf("  - %s: %s\n", name, meta.Description)
	}

	fmt.Println("\n=== Metadata Registration Demo Complete ===")
	fmt.Println("\nThe service is now running and will send heartbeat every 30 seconds.")
	fmt.Println("You can:")
	fmt.Println("  1. Use the LightLink Console to see this service")
	fmt.Println("  2. Call the methods via RPC")
	fmt.Println("  3. View service metadata in the Console UI")
	fmt.Println("\nPress Ctrl+C to stop the service.")

	// Keep service running
	select {}
}

// addHandler adds two numbers
func addHandler(args map[string]interface{}) (map[string]interface{}, error) {
	a := args["a"].(float64)
	b := args["b"].(float64)
	return map[string]interface{}{"sum": a + b}, nil
}

// multiplyHandler multiplies two numbers
func multiplyHandler(args map[string]interface{}) (map[string]interface{}, error) {
	a := args["a"].(float64)
	b := args["b"].(float64)
	return map[string]interface{}{"product": a * b}, nil
}

// powerHandler calculates a to the power of b
func powerHandler(args map[string]interface{}) (map[string]interface{}, error) {
	base := args["base"].(float64)
	exp := args["exp"].(float64)

	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}

	return map[string]interface{}{"result": result}, nil
}

// divideHandler divides two numbers
func divideHandler(args map[string]interface{}) (map[string]interface{}, error) {
	numerator := args["numerator"].(float64)
	denominator := args["denominator"].(float64)

	if denominator == 0 {
		return nil, fmt.Errorf("division by zero")
	}

	return map[string]interface{}{"quotient": numerator / denominator}, nil
}
