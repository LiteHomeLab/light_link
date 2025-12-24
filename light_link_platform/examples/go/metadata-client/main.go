package main

import (
	"fmt"
	"log"
	"time"

	"github.com/LiteHomeLab/light_link/sdk/go/client"
	"github.com/LiteHomeLab/light_link/examples"
)

func main() {
	config := examples.GetConfig()

	fmt.Println("=== Metadata Service Client Demo ===")
	fmt.Println("NATS URL:", config.NATSURL)

	// Create client
	fmt.Println("\n[1/2] Creating client...")
	cli, err := client.NewClient(config.NATSURL, nil)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer cli.Close()
	fmt.Println("Client created successfully!")

	// Test service methods
	fmt.Println("\n[2/2] Calling service methods...")

	// Test add
	fmt.Println("\n--- Test 1: Add (10 + 20) ---")
	result, err := cli.Call("math-service", "add", map[string]interface{}{
		"a": 10.0,
		"b": 20.0,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Result: %v\n", result)
	}

	// Test multiply
	fmt.Println("\n--- Test 2: Multiply (5 * 6) ---")
	result, err = cli.Call("math-service", "multiply", map[string]interface{}{
		"a": 5.0,
		"b": 6.0,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Result: %v\n", result)
	}

	// Test power
	fmt.Println("\n--- Test 3: Power (2^10) ---")
	result, err = cli.Call("math-service", "power", map[string]interface{}{
		"base": 2.0,
		"exp":   10.0,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Result: %v\n", result)
	}

	// Test divide
	fmt.Println("\n--- Test 4: Divide (100 / 4) ---")
	result, err = cli.Call("math-service", "divide", map[string]interface{}{
		"numerator":   100.0,
		"denominator": 4.0,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Result: %v\n", result)
	}

	// Test division by zero error
	fmt.Println("\n--- Test 5: Divide by zero (error case) ---")
	result, err = cli.Call("math-service", "divide", map[string]interface{}{
		"numerator":   100.0,
		"denominator": 0.0,
	})
	if err != nil {
		fmt.Printf("Expected Error: %v\n", err)
	} else {
		fmt.Printf("Result: %v\n", result)
	}

	// Test CallWithTimeout
	fmt.Println("\n--- Test 6: Call with timeout (Add with 5s timeout) ---")
	start := time.Now()
	result, err = cli.CallWithTimeout("math-service", "add", map[string]interface{}{
		"a": 100.0,
		"b": 200.0,
	}, 5*time.Second)
	elapsed := time.Since(start)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Result: %v (took %v)\n", result, elapsed)
	}

	// Test with JSON numbers
	fmt.Println("\n--- Test 7: Large numbers (1000 + 2000) ---")
	result, err = cli.Call("math-service", "add", map[string]interface{}{
		"a": 1000.0,
		"b": 2000.0,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Result: %v\n", result)
	}

	// Test negative numbers
	fmt.Println("\n--- Test 8: Negative numbers (-10 + 5) ---")
	result, err = cli.Call("math-service", "add", map[string]interface{}{
		"a": -10.0,
		"b": 5.0,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Result: %v\n", result)
	}

	// Test power with larger exponent
	fmt.Println("\n--- Test 9: Power (3^5) ---")
	result, err = cli.Call("math-service", "power", map[string]interface{}{
		"base": 3.0,
		"exp":   5.0,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Result: %v\n", result)
	}

	fmt.Println("\n=== Metadata Service Client Demo Complete ===")
}
