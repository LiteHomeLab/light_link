package client

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
)

// HostInfo contains information about the host where the service is running
type HostInfo struct {
	IP         string
	MAC        string
	WorkingDir string
	Language   string
}

// GetHostInfo retrieves host information for service instance identification
func GetHostInfo() (*HostInfo, error) {
	ip, mac, err := getPrimaryIPAndMAC()
	if err != nil {
		return nil, fmt.Errorf("get IP and MAC: %w", err)
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}

	return &HostInfo{
		IP:         ip,
		MAC:        mac,
		WorkingDir: workingDir,
		Language:   getLanguage(),
	}, nil
}

// getPrimaryIPAndMAC returns the primary IP address and MAC address
func getPrimaryIPAndMAC() (string, string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", "", fmt.Errorf("get interfaces: %w", err)
	}

	for _, iface := range interfaces {
		// Skip down interfaces and loopback
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// Skip interfaces without MAC address
		mac := iface.HardwareAddr.String()
		if mac == "" {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			// Prefer IPv4 addresses
			ip = ip.To4()
			if ip != nil {
				return ip.String(), formatMAC(mac), nil
			}
		}
	}

	return "", "", fmt.Errorf("no valid IP address found")
}

// formatMAC formats MAC address with colons (aa:bb:cc:dd:ee:ff)
func formatMAC(mac string) string {
	if len(mac) < 12 {
		return mac
	}
	// If already formatted with colons, return as is
	if strings.Contains(mac, ":") {
		return mac
	}
	// Format without colons to with colons
	var result string
	for i := 0; i < len(mac); i += 2 {
		if i+2 <= len(mac) {
			if result != "" {
				result += ":"
			}
			result += mac[i : i+2]
		}
	}
	return result
}

// getLanguage returns the programming language name
func getLanguage() string {
	return "go"
}

// GetLanguageVersion returns the Go runtime version
func GetLanguageVersion() string {
	return runtime.Version()
}
