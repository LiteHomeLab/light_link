package main

import (
	"context"
	"log"

	"github.com/LiteHomeLab/light_link/examples"
	"github.com/LiteHomeLab/light_link/sdk/go/client"
	"github.com/WQGroup/logger"
)

func main() {
	logger.SetLoggerName("call-math-service-go")
	logger.Info("=== Call Math Service Demo ===")

	config := examples.GetConfig()
	logger.Infof("NATS URL: %s", config.NATSURL)

	// Create client
	logger.Info("Connecting to NATS...")
	cli, err := client.NewClient(config.NATSURL, client.WithAutoTLS())
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer cli.Close()
	logger.Info("Connected successfully")

	// Define dependencies
	deps := []client.Dependency{
		{
			ServiceName: "math-service-go",
			Methods:     []string{"add", "multiply", "power", "divide"},
		},
	}

	// Wait for dependencies
	checker := client.NewDependencyChecker(cli.GetNATSConn(), deps)
	if err := checker.WaitForDependencies(context.Background()); err != nil {
		log.Fatalf("Failed to wait for dependencies: %v", err)
	}
	defer checker.Close()

	// Perform calculations
	performCalculations(cli)

	logger.Info("=== Demo complete ===")
}

func performCalculations(cli *client.Client) {
	logger.Info("")
	logger.Info("=== Starting calculations ===")
	logger.Info("")

	// 1. add(10, 20)
	result, err := cli.Call("math-service-go", "add", map[string]interface{}{
		"a": float64(10),
		"b": float64(20),
	})
	if err != nil {
		logger.Errorf("add failed: %v", err)
	} else {
		logger.Infof("add(10, 20) = %v", result)
	}

	// 2. multiply(5, 6)
	result, err = cli.Call("math-service-go", "multiply", map[string]interface{}{
		"a": float64(5),
		"b": float64(6),
	})
	if err != nil {
		logger.Errorf("multiply failed: %v", err)
	} else {
		logger.Infof("multiply(5, 6) = %v", result)
	}

	// 3. power(2, 10)
	result, err = cli.Call("math-service-go", "power", map[string]interface{}{
		"base": float64(2),
		"exp":   float64(10),
	})
	if err != nil {
		logger.Errorf("power failed: %v", err)
	} else {
		logger.Infof("power(2, 10) = %v", result)
	}

	// 4. divide(100, 4)
	result, err = cli.Call("math-service-go", "divide", map[string]interface{}{
		"numerator":   float64(100),
		"denominator": float64(4),
	})
	if err != nil {
		logger.Errorf("divide failed: %v", err)
	} else {
		logger.Infof("divide(100, 4) = %v", result)
	}

	// 5. Complex calculation: add(multiply(3, 4), divide(20, 2))
	// First: multiply(3, 4)
	result, err = cli.Call("math-service-go", "multiply", map[string]interface{}{
		"a": float64(3),
		"b": float64(4),
	})
	if err != nil {
		logger.Errorf("Complex calculation multiply failed: %v", err)
		return
	}
	product := result["product"].(float64)

	// Second: divide(20, 2)
	result, err = cli.Call("math-service-go", "divide", map[string]interface{}{
		"numerator":   float64(20),
		"denominator": float64(2),
	})
	if err != nil {
		logger.Errorf("Complex calculation divide failed: %v", err)
		return
	}
	quotient := result["quotient"].(float64)

	// Third: add(product, quotient)
	result, err = cli.Call("math-service-go", "add", map[string]interface{}{
		"a": product,
		"b": quotient,
	})
	if err != nil {
		logger.Errorf("Complex calculation add failed: %v", err)
	} else {
		logger.Infof("Complex: add(multiply(3, 4), divide(20, 2)) = add(%.0f, %.0f) = %v",
			product, quotient, result)
	}
}
