package client

import (
	"strings"
	"testing"
)

// TestGetHostInfo tests retrieving host information
func TestGetHostInfo(t *testing.T) {
	info, err := GetHostInfo()
	if err != nil {
		t.Fatalf("Failed to get host info: %v", err)
	}

	// Check that we got valid information
	if info.IP == "" {
		t.Error("IP address is empty")
	}
	if info.MAC == "" {
		t.Error("MAC address is empty")
	}
	if info.WorkingDir == "" {
		t.Error("Working directory is empty")
	}
	if info.Language != "go" {
		t.Errorf("Expected language 'go', got '%s'", info.Language)
	}

	t.Logf("Host Info: IP=%s, MAC=%s, WorkingDir=%s, Language=%s",
		info.IP, info.MAC, info.WorkingDir, info.Language)
}

// TestFormatMAC tests MAC address formatting
func TestFormatMAC(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasColon bool
	}{
		{
			name:     "MAC with colons",
			input:    "aa:bb:cc:dd:ee:ff",
			hasColon: true,
		},
		{
			name:     "MAC without colons",
			input:    "aabbccddeeff",
			hasColon: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatMAC(tt.input)
			if tt.hasColon && !strings.Contains(result, ":") {
				t.Errorf("formatMAC(%q) expected colons, got %q", tt.input, result)
			}
		})
	}
}

// TestGetLanguage tests language detection
func TestGetLanguage(t *testing.T) {
	lang := getLanguage()
	if lang != "go" {
		t.Errorf("Expected 'go', got '%s'", lang)
	}
}

// TestGetLanguageVersion tests that we can get the Go version
func TestGetLanguageVersion(t *testing.T) {
	version := GetLanguageVersion()
	if version == "" {
		t.Error("Version is empty")
	}
	if !strings.HasPrefix(version, "go") {
		t.Errorf("Version should start with 'go', got '%s'", version)
	}
	t.Logf("Go version: %s", version)
}
