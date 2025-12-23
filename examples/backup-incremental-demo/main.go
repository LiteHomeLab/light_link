package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/LiteHomeLab/light_link/examples"
	"github.com/LiteHomeLab/light_link/sdk/go/client"
	"github.com/LiteHomeLab/light_link/sdk/go/service"
)

// DatabaseData simulates a database snapshot
type DatabaseData struct {
	Users []User `json:"users"`
}

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

func main() {
	fmt.Println("========================================")
	fmt.Println("   LightLink Incremental Backup Demo")
	fmt.Println("========================================")

	// Load configuration
	config := examples.GetConfig()

	// Start Backup Agent
	fmt.Println("\n[1/8] Starting Backup Agent...")
	backupSvc, err := service.NewBackupService("backup-agent", config.NATSURL, nil, "./backups")
	if err != nil {
		log.Fatalf("Failed to create backup service: %v", err)
	}
	defer backupSvc.Stop()

	if err := backupSvc.Start(); err != nil {
		log.Fatalf("Failed to start backup service: %v", err)
	}
	fmt.Println("   Backup Agent started")

	// Create backup client
	fmt.Println("\n[2/8] Creating backup client...")
	cli, err := client.NewClient(config.NATSURL, nil)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer cli.Close()
	fmt.Println("   Client created")

	serviceName := "demo-service"
	backupName := "incremental-test-db"

	// Initial database state
	fmt.Println("\n[3/8] Creating initial full backup...")
	db1 := DatabaseData{
		Users: []User{
			{ID: 1, Name: "Alice", Role: "admin"},
			{ID: 2, Name: "Bob", Role: "user"},
			{ID: 3, Name: "Charlie", Role: "user"},
		},
	}
	data1, _ := json.Marshal(db1)

	_, err = cli.CreateBackup(serviceName, backupName, data1)
	if err != nil {
		log.Fatalf("Failed to create backup v1: %v", err)
	}
	fmt.Printf("   Created full backup v1 (%d bytes)\n", len(data1))
	fmt.Printf("   Data: %s\n", string(data1))

	time.Sleep(500 * time.Millisecond)

	// Add one user - incremental
	fmt.Println("\n[4/8] Adding one user (incremental backup)...")
	db2 := DatabaseData{
		Users: []User{
			{ID: 1, Name: "Alice", Role: "admin"},
			{ID: 2, Name: "Bob", Role: "user"},
			{ID: 3, Name: "Charlie", Role: "user"},
			{ID: 4, Name: "David", Role: "user"},
		},
	}
	data2, _ := json.Marshal(db2)

	_, err = cli.CreateIncrementalBackup(serviceName, backupName, data2)
	if err != nil {
		log.Fatalf("Failed to create incremental backup v2: %v", err)
	}
	fmt.Printf("   Created incremental backup v2\n")
	fmt.Printf("   Data: %s\n", string(data2))

	time.Sleep(500 * time.Millisecond)

	// Modify one user - incremental
	fmt.Println("\n[5/8] Modifying one user (incremental backup)...")
	db3 := DatabaseData{
		Users: []User{
			{ID: 1, Name: "Alice", Role: "admin"},
			{ID: 2, Name: "Bob", Role: "moderator"},
			{ID: 3, Name: "Charlie", Role: "user"},
			{ID: 4, Name: "David", Role: "user"},
		},
	}
	data3, _ := json.Marshal(db3)

	_, err = cli.CreateIncrementalBackup(serviceName, backupName, data3)
	if err != nil {
		log.Fatalf("Failed to create incremental backup v3: %v", err)
	}
	fmt.Printf("   Created incremental backup v3\n")
	fmt.Printf("   Data: %s\n", string(data3))

	time.Sleep(500 * time.Millisecond)

	// Add more users - incremental
	fmt.Println("\n[6/8] Adding more users (incremental backup)...")
	db4 := DatabaseData{
		Users: []User{
			{ID: 1, Name: "Alice", Role: "admin"},
			{ID: 2, Name: "Bob", Role: "moderator"},
			{ID: 3, Name: "Charlie", Role: "user"},
			{ID: 4, Name: "David", Role: "user"},
			{ID: 5, Name: "Eve", Role: "user"},
			{ID: 6, Name: "Frank", Role: "user"},
		},
	}
	data4, _ := json.Marshal(db4)

	_, err = cli.CreateIncrementalBackup(serviceName, backupName, data4)
	if err != nil {
		log.Fatalf("Failed to create incremental backup v4: %v", err)
	}
	fmt.Printf("   Created incremental backup v4\n")
	fmt.Printf("   Data: %s\n", string(data4))

	// List all versions
	fmt.Println("\n[7/8] Listing all backup versions...")
	versions, err := cli.ListBackups(serviceName, backupName)
	if err != nil {
		log.Fatalf("Failed to list backups: %v", err)
	}
	fmt.Printf("   Total versions: %d\n", len(versions))
	for _, v := range versions {
		fmt.Printf("   - Version %d: type=%s, size=%d bytes, base=%d\n",
			v.Version, v.Type, v.FileSize, v.BaseVersion)
	}

	// Calculate space savings
	fmt.Println("\n[8/8] Space analysis...")
	fullSize := len(data1)
	totalIncrementalSize := 0
	for _, v := range versions {
		if v.Type == "incremental" {
			totalIncrementalSize += int(v.FileSize)
		}
	}
	fmt.Printf("   Full backup size: %d bytes\n", fullSize)
	fmt.Printf("   Total incremental size: %d bytes\n", totalIncrementalSize)
	if totalIncrementalSize < len(data4) {
		savings := 100.0 - (float64(totalIncrementalSize) / float64(len(data4)) * 100)
		fmt.Printf("   Space saved: %.1f%%\n", savings)
	}

	fmt.Println("\n========================================")
	fmt.Println("Incremental Backup Demo Completed!")
	fmt.Println("========================================")
	fmt.Println("\nKey findings:")
	fmt.Println("  - Incremental backups only store changes")
	fmt.Println("  - Significant space savings for small changes")
	fmt.Println("  - Each version can be restored independently")
	fmt.Println("\nNote: For this demo, restoring from incremental backups")
	fmt.Println("      would require applying the diff chain (not shown here).")
}
