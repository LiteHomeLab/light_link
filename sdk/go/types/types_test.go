package types

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverClientCerts(t *testing.T) {
	tempDir := t.TempDir()
	clientDir := filepath.Join(tempDir, "client")
	os.MkdirAll(clientDir, 0755)

	os.WriteFile(filepath.Join(clientDir, "ca.crt"), []byte("ca content"), 0644)
	os.WriteFile(filepath.Join(clientDir, "client.crt"), []byte("cert content"), 0644)
	os.WriteFile(filepath.Join(clientDir, "client.key"), []byte("key content"), 0644)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	result, err := DiscoverClientCerts()
	if err != nil {
		t.Fatalf("DiscoverClientCerts failed: %v", err)
	}

	if !result.Found {
		t.Fatal("Expected certificates to be found")
	}

	if result.ServerName != DefaultServerName {
		t.Errorf("Expected ServerName=%s, got=%s", DefaultServerName, result.ServerName)
	}
}

func TestDiscoverClientCertsNotFound(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	_, err := DiscoverClientCerts()
	if err == nil {
		t.Fatal("Expected error when certificates not found")
	}
}

func TestDiscoverServerCerts(t *testing.T) {
	tempDir := t.TempDir()
	serverDir := filepath.Join(tempDir, "nats-server")

	os.MkdirAll(serverDir, 0755)
	os.WriteFile(filepath.Join(serverDir, "ca.crt"), []byte("ca content"), 0644)
	os.WriteFile(filepath.Join(serverDir, "server.crt"), []byte("cert content"), 0644)
	os.WriteFile(filepath.Join(serverDir, "server.key"), []byte("key content"), 0644)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	result, err := DiscoverServerCerts()
	if err != nil {
		t.Fatalf("DiscoverServerCerts failed: %v", err)
	}

	if !result.Found {
		t.Fatal("Expected certificates to be found")
	}
}

func TestDiscoverServerCertsNotFound(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	_, err := DiscoverServerCerts()
	if err == nil {
		t.Fatal("Expected error when certificates not found")
	}
}

func TestCertDiscoveryResultToTLSConfig(t *testing.T) {
	result := &CertDiscoveryResult{
		CaFile:     "ca.crt",
		CertFile:   "cert.crt",
		KeyFile:    "key.key",
		ServerName: "test-server",
		Found:      true,
	}

	config := CertDiscoveryResultToTLSConfig(result)

	if config.CaFile != result.CaFile {
		t.Errorf("CaFile mismatch")
	}
	if config.ServerName != result.ServerName {
		t.Errorf("ServerName mismatch")
	}
}

func TestDiscoverClientCertsParentDirectory(t *testing.T) {
	tempDir := t.TempDir()

	clientDir := filepath.Join(tempDir, "client")
	subDir := filepath.Join(tempDir, "subdir")
	os.MkdirAll(clientDir, 0755)
	os.MkdirAll(subDir, 0755)

	os.WriteFile(filepath.Join(clientDir, "ca.crt"), []byte("ca content"), 0644)
	os.WriteFile(filepath.Join(clientDir, "client.crt"), []byte("cert content"), 0644)
	os.WriteFile(filepath.Join(clientDir, "client.key"), []byte("key content"), 0644)

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(subDir)

	result, err := DiscoverClientCerts()
	if err != nil {
		t.Logf("tempDir: %s", tempDir)
		t.Logf("clientDir: %s", clientDir)
		t.Logf("subDir: %s", subDir)
		t.Logf("cwd: %s", mustGetwd())
		t.Fatalf("DiscoverClientCerts failed: %v", err)
	}

	if !result.Found {
		t.Fatal("Expected certificates to be found in parent directory")
	}
}

func mustGetwd() string {
	wd, _ := os.Getwd()
	return wd
}

func TestFileExists(t *testing.T) {
	tempDir := t.TempDir()

	existingFile := filepath.Join(tempDir, "existing.txt")
	os.WriteFile(existingFile, []byte("content"), 0644)

	nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")

	if !fileExists(existingFile) {
		t.Error("Expected file to exist")
	}

	if fileExists(nonExistentFile) {
		t.Error("Expected file to not exist")
	}

	dir := filepath.Join(tempDir, "directory")
	os.Mkdir(dir, 0755)
	if fileExists(dir) {
		t.Error("Expected directory to return false")
	}
}
