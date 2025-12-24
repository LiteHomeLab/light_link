package storage

import (
	"os"
	"testing"
	"time"

	"github.com/LiteHomeLab/light_link/sdk/go/types"
)

func setupTestDB(t *testing.T) *Database {
	f, err := os.CreateTemp("", "test_*.db")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	db, err := NewDatabase(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
		os.Remove(f.Name())
	})
	return db
}

func TestDatabaseInit(t *testing.T) {
	db := setupTestDB(t)

	// Verify tables exist
	var tableName string
	err := db.db.QueryRow(`
		SELECT name FROM sqlite_master
		WHERE type='table' AND name='services'
	`).Scan(&tableName)
	if err != nil {
		t.Errorf("Services table not found: %v", err)
	}

	// Check other tables
	tables := []string{"methods", "service_status", "events", "users", "call_history", "service_status_history"}
	for _, table := range tables {
		err := db.db.QueryRow(`
			SELECT name FROM sqlite_master
			WHERE type='table' AND name=?
		`, table).Scan(&tableName)
		if err != nil {
			t.Errorf("Table %s not found: %v", table, err)
		}
	}
}

func TestSaveAndGetService(t *testing.T) {
	db := setupTestDB(t)

	meta := &ServiceMetadata{
		Name:        "test-service",
		Version:     "v1.0.0",
		Description: "Test service",
		Author:      "Test Author",
		Tags:        []string{"test", "demo"},
	}

	// Save service
	if err := db.SaveService(meta); err != nil {
		t.Fatalf("SaveService failed: %v", err)
	}

	// Get service
	retrieved, err := db.GetService("test-service")
	if err != nil {
		t.Fatalf("GetService failed: %v", err)
	}

	if retrieved.Name != "test-service" {
		t.Errorf("Expected name 'test-service', got '%s'", retrieved.Name)
	}
	if retrieved.Version != "v1.0.0" {
		t.Errorf("Expected version 'v1.0.0', got '%s'", retrieved.Version)
	}
	if len(retrieved.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(retrieved.Tags))
	}
}

func TestListServices(t *testing.T) {
	db := setupTestDB(t)

	// Save multiple services
	services := []ServiceMetadata{
		{Name: "svc1", Version: "v1.0.0"},
		{Name: "svc2", Version: "v2.0.0"},
		{Name: "svc3", Version: "v3.0.0"},
	}

	for _, svc := range services {
		if err := db.SaveService(&svc); err != nil {
			t.Fatalf("SaveService failed: %v", err)
		}
	}

	// List services
	list, err := db.ListServices()
	if err != nil {
		t.Fatalf("ListServices failed: %v", err)
	}

	if len(list) != 3 {
		t.Errorf("Expected 3 services, got %d", len(list))
	}
}

func TestSaveMethod(t *testing.T) {
	db := setupTestDB(t)

	// Create a service first
	service := &ServiceMetadata{
		Name:    "test-method-service",
		Version: "v1.0.0",
	}
	if err := db.SaveService(service); err != nil {
		t.Fatalf("SaveService failed: %v", err)
	}

	serviceID, err := db.GetServiceID("test-method-service")
	if err != nil {
		t.Fatalf("GetServiceID failed: %v", err)
	}

	// Save method
	method := &MethodMetadata{
		ServiceID:   serviceID,
		Name:        "testMethod",
		Description: "Test method",
		Params: []types.ParameterMetadata{
			{Name: "a", Type: "number", Required: true},
		},
		Returns: []types.ReturnMetadata{
			{Name: "result", Type: "number"},
		},
	}

	if err := db.SaveMethod(serviceID, method); err != nil {
		t.Fatalf("SaveMethod failed: %v", err)
	}

	// Get methods
	methods, err := db.GetMethods("test-method-service")
	if err != nil {
		t.Fatalf("GetMethods failed: %v", err)
	}

	if len(methods) != 1 {
		t.Errorf("Expected 1 method, got %d", len(methods))
	}
}

func TestUpdateServiceStatus(t *testing.T) {
	db := setupTestDB(t)

	// Create a service first
	service := &ServiceMetadata{
		Name:    "test-status-service",
		Version: "v1.0.0",
	}
	if err := db.SaveService(service); err != nil {
		t.Fatalf("SaveService failed: %v", err)
	}

	// Update status to online
	if err := db.UpdateServiceStatus("test-status-service", true, "v1.0.0"); err != nil {
		t.Fatalf("UpdateServiceStatus failed: %v", err)
	}

	// Get status
	status, err := db.GetServiceStatus("test-status-service")
	if err != nil {
		t.Fatalf("GetServiceStatus failed: %v", err)
	}

	if !status.Online {
		t.Error("Expected service to be online")
	}

	// Check history
	history, err := db.GetServiceStatusHistory("test-status-service", 10)
	if err != nil {
		t.Fatalf("GetServiceStatusHistory failed: %v", err)
	}

	if len(history) != 1 {
		t.Errorf("Expected 1 history entry, got %d", len(history))
	}
}

func TestSaveEvent(t *testing.T) {
	db := setupTestDB(t)

	event := &ServiceEvent{
		Type:      "registered",
		Service:   "test-service",
		Timestamp: time.Now(),
	}

	if err := db.SaveEvent(event); err != nil {
		t.Fatalf("SaveEvent failed: %v", err)
	}

	// List events
	events, err := db.ListEvents(10, 0)
	if err != nil {
		t.Fatalf("ListEvents failed: %v", err)
	}

	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
}

func TestCreateUser(t *testing.T) {
	db := setupTestDB(t)

	if err := db.CreateUser("testuser", "password123", "admin"); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Validate user
	user, err := db.ValidateUser("testuser", "password123")
	if err != nil {
		t.Fatalf("ValidateUser failed: %v", err)
	}

	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}

	if user.Role != "admin" {
		t.Errorf("Expected role 'admin', got '%s'", user.Role)
	}

	// Test wrong password
	_, err = db.ValidateUser("testuser", "wrongpassword")
	if err == nil {
		t.Error("Expected error for wrong password")
	}
}

func TestInitAdminUser(t *testing.T) {
	db := setupTestDB(t)

	// First call should create user
	if err := db.InitAdminUser("admin", "admin123"); err != nil {
		t.Fatalf("InitAdminUser failed: %v", err)
	}

	// Verify user exists
	user, err := db.GetUser("admin")
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	if user.Role != "admin" {
		t.Errorf("Expected role 'admin', got '%s'", user.Role)
	}

	// Second call should not fail (user already exists)
	if err := db.InitAdminUser("admin", "different"); err != nil {
		t.Fatalf("Second InitAdminUser should not fail: %v", err)
	}

	// Password should not have changed
	_, err = db.ValidateUser("admin", "different")
	if err == nil {
		t.Error("Password should not have changed")
	}
}

func TestGetOnlineServices(t *testing.T) {
	db := setupTestDB(t)

	// Create services
	for _, name := range []string{"svc1", "svc2", "svc3"} {
		service := &ServiceMetadata{Name: name, Version: "v1.0.0"}
		if err := db.SaveService(service); err != nil {
			t.Fatalf("SaveService failed: %v", err)
		}
	}

	// Set svc1 and svc2 online
	db.UpdateServiceStatus("svc1", true, "v1.0.0")
	db.UpdateServiceStatus("svc2", true, "v1.0.0")
	db.UpdateServiceStatus("svc3", false, "v1.0.0")

	// Get online services
	online, err := db.GetOnlineServices()
	if err != nil {
		t.Fatalf("GetOnlineServices failed: %v", err)
	}

	if len(online) != 2 {
		t.Errorf("Expected 2 online services, got %d", len(online))
	}

	// Get offline services
	offline, err := db.GetOfflineServices()
	if err != nil {
		t.Fatalf("GetOfflineServices failed: %v", err)
	}

	if len(offline) != 1 {
		t.Errorf("Expected 1 offline service, got %d", len(offline))
	}
}

func TestListServiceStatus(t *testing.T) {
	db := setupTestDB(t)

	// Create services
	for _, name := range []string{"svc1", "svc2"} {
		service := &ServiceMetadata{Name: name, Version: "v1.0.0"}
		if err := db.SaveService(service); err != nil {
			t.Fatalf("SaveService failed: %v", err)
		}
	}

	db.UpdateServiceStatus("svc1", true, "v1.0.0")
	db.UpdateServiceStatus("svc2", false, "v2.0.0")

	// List all statuses
	statuses, err := db.ListServiceStatus()
	if err != nil {
		t.Fatalf("ListServiceStatus failed: %v", err)
	}

	if len(statuses) != 2 {
		t.Errorf("Expected 2 statuses, got %d", len(statuses))
	}
}
