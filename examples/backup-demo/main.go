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
	fmt.Println("\n[1/5] Starting Backup Agent...")
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
	fmt.Println("\n[2/5] Creating backup client...")
	cli, err := client.NewClient(config.NATSURL, nil)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer cli.Close()
	fmt.Println("   Client created")

	// Create backup versions
	fmt.Println("\n[3/5] Creating backup versions...")
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

	// Version 3
	data3 := []byte(`{"users": [{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}, {"id": 3, "name": "Charlie"}, {"id": 4, "name": "David"}]}`)
	version3, err := cli.CreateBackup(serviceName, backupName, data3)
	if err != nil {
		log.Fatalf("Failed to create backup v3: %v", err)
	}
	fmt.Printf("   Created backup v%d (%d bytes)\n", version3, len(data3))

	// List all versions
	fmt.Println("\n[4/5] Listing backup versions...")
	versions, err := cli.ListBackups(serviceName, backupName)
	if err != nil {
		log.Fatalf("Failed to list backups: %v", err)
	}
	fmt.Printf("   Total versions: %d\n", len(versions))
	for _, v := range versions {
		fmt.Printf("   - Version %d: %d bytes, checksum=%s\n", v.Version, v.FileSize, v.Checksum[:16]+"...")
	}

	// Restore specific version
	fmt.Println("\n[5/5] Restoring backup v2...")
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

	fmt.Println("\n========================================")
	fmt.Println("Demo completed successfully!")
	fmt.Println("========================================")
	fmt.Println("\nBackup storage location: ./backups/")
	fmt.Printf("Backup directory: %s.%s\n", serviceName, backupName)
	fmt.Println("\nYou can inspect the stored files:")
	fmt.Println("  - metadata.json  : Backup metadata")
	fmt.Println("  - v1.bin         : Version 1 data")
	fmt.Println("  - v2.bin         : Version 2 data")
	fmt.Println("  - v3.bin         : Version 3 data")
}
