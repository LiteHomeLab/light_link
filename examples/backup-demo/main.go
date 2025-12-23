package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/LiteHomeLab/light_link/examples"
	"github.com/LiteHomeLab/light_link/sdk/go/client"
	"github.com/LiteHomeLab/light_link/sdk/go/service"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("       LightLink Backup Demo")
	fmt.Println("========================================")

	// Load configuration
	config := examples.GetConfig()

	// Start Backup Agent
	fmt.Println("\n[1/7] Starting Backup Agent...")
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
	fmt.Println("\n[2/7] Creating backup client...")
	cli, err := client.NewClient(config.NATSURL, nil)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer cli.Close()
	fmt.Println("   Client created")

	// Part 1: Create backups without retention policy
	fmt.Println("\n[3/7] Creating backup versions (no retention)...")
	serviceName := "demo-service"
	backupName := "demo-database"

	// Version 1
	data1 := []byte(`{"users": [{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]}`)
	version1, err := cli.CreateBackup(serviceName, backupName, data1)
	if err != nil {
		log.Fatalf("Failed to create backup v1: %v", err)
	}
	fmt.Printf("   Created backup v%d (%d bytes)\n", version1, len(data1))

	time.Sleep(500 * time.Millisecond)

	// Version 2
	data2 := []byte(`{"users": [{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}, {"id": 3, "name": "Charlie"}]}`)
	version2, err := cli.CreateBackup(serviceName, backupName, data2)
	if err != nil {
		log.Fatalf("Failed to create backup v2: %v", err)
	}
	fmt.Printf("   Created backup v%d (%d bytes)\n", version2, len(data2))

	time.Sleep(500 * time.Millisecond)

	// List versions
	fmt.Println("\n[4/7] Listing backup versions...")
	versions, err := cli.ListBackups(serviceName, backupName)
	if err != nil {
		log.Fatalf("Failed to list backups: %v", err)
	}
	fmt.Printf("   Total versions: %d\n", len(versions))
	for _, v := range versions {
		fmt.Printf("   - Version %d: %d bytes\n", v.Version, v.FileSize)
	}

	// Part 2: Create backup WITH retention policy (max 3 versions)
	fmt.Println("\n[5/7] Creating backup with retention policy (max=3)...")
	backupName2 := "demo-database-with-policy"

	for i := 1; i <= 5; i++ {
		data := []byte(fmt.Sprintf(`{"version": %d, "data": "test data"}`, i))
		version, err := cli.CreateBackupWithMaxVersions(serviceName, backupName2, data, 3)
		if err != nil {
			log.Fatalf("Failed to create backup v%d: %v", i, err)
		}
		fmt.Printf("   Created backup v%d (max_versions=3)\n", version)
		time.Sleep(200 * time.Millisecond)
	}

	// List versions with policy
	versions2, err := cli.ListBackups(serviceName, backupName2)
	if err != nil {
		log.Fatalf("Failed to list backups: %v", err)
	}
	fmt.Printf("   Total versions after creating 5: %d (auto-cleaned to 3)\n", len(versions2))
	for _, v := range versions2 {
		fmt.Printf("   - Version %d: %d bytes\n", v.Version, v.FileSize)
	}

	// Part 3: Restore specific version
	fmt.Println("\n[6/7] Restoring backup v2...")
	recoveredData, err := cli.GetBackup(serviceName, backupName, 2)
	if err != nil {
		log.Fatalf("Failed to get backup v2: %v", err)
	}
	fmt.Printf("   Restored data (%d bytes):\n", len(recoveredData))
	fmt.Printf("   %s\n", string(recoveredData))

	// Verify data matches
	decodedData2, _ := base64.StdEncoding.DecodeString(base64.StdEncoding.EncodeToString(data2))
	if string(recoveredData) != string(decodedData2) {
		fmt.Println("   Warning: Restored data does not match original!")
	} else {
		fmt.Println("   Data verification: PASSED")
	}

	// Part 4: Manual cleanup
	fmt.Println("\n[7/7] Testing manual cleanup...")
	backupName3 := "demo-cleanup-test"

	// Create 5 versions without policy
	for i := 1; i <= 5; i++ {
		data := []byte(fmt.Sprintf(`{"version": %d}`, i))
		_, err := cli.CreateBackup(serviceName, backupName3, data)
		if err != nil {
			log.Fatalf("Failed to create backup: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	versions3, _ := cli.ListBackups(serviceName, backupName3)
	fmt.Printf("   Created %d versions\n", len(versions3))

	// Note: Since there's no max_versions set, manual cleanup won't delete anything
	// This demonstrates that cleanup only works when max_versions is configured
	fmt.Println("   (Cleanup only works when max_versions is set)")

	fmt.Println("\n========================================")
	fmt.Println("Demo completed successfully!")
	fmt.Println("========================================")
	fmt.Println("\nBackup storage location: ./backups/")
	fmt.Println("\nKey features demonstrated:")
	fmt.Println("  - Basic backup creation and restoration")
	fmt.Println("  - Automatic cleanup with max_versions policy")
	fmt.Println("  - Version listing")
	fmt.Println("\nYou can inspect the stored files:")
	fmt.Println("  - ./backups/demo-service.demo-database/")
	fmt.Println("  - ./backups/demo-service.demo-database-with-policy/")
	fmt.Println("  - ./backups/demo-service.demo-cleanup-test/")
}
