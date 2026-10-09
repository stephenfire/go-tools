package tools

import (
	"net"
	"testing"

	"github.com/stephenfire/go-tools/strategy"
)

func TestLocalFirstIP_ExplicitStrategy(t *testing.T) {
	// Must pass an explicit strategy since no default exists
	ip, err := LocalFirstIP(strategy.StandardPhysicalIPv4)
	if err != nil {
		t.Logf("LocalFirstIP returned error on current system: %v", err)
		return
	}

	if ip != nil {
		if ip.To4() == nil {
			t.Errorf("LocalFirstIP returned an invalid IPv4 address: %s", ip.String())
		}
		if ip.IsLoopback() {
			t.Errorf("LocalFirstIP unexpectedly returned loopback IP: %s", ip.String())
		}
	}
}

func TestLocalFirstIP_NilStrategy(t *testing.T) {
	// Passing nil strategy means no filtering; should match the very first available IP (e.g., 127.0.0.1 or IPv6)
	ip, err := LocalFirstIP(nil)
	if err != nil {
		t.Fatalf("LocalFirstIP with nil strategy failed: %v", err)
	}

	if ip == nil {
		t.Errorf("LocalFirstIP with nil strategy expected to return at least one system IP, got nil")
	}
}

func TestHostIP_ParseFormats(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		hasErr   bool
	}{
		{"IPv4 Plain", "192.168.1.1", "192.168.1.1", false},
		{"IPv4 with Port", "10.0.0.1:8080", "10.0.0.1", false},
		{"IPv4 CIDR", "172.16.0.5/24", "172.16.0.5", false},
		{"IPv6 Plain", "fe80::1", "fe80::1", false},
		{"Localhost Domain", "localhost", "127.0.0.1", false},
		{"Invalid String", "invalid_host_123456789", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip, err := HostIP(tt.input)
			if (err != nil) != tt.hasErr {
				t.Fatalf("HostIP(%q) error status = %v, wantErr %v", tt.input, err, tt.hasErr)
			}
			if !tt.hasErr && ip.String() != tt.expected && tt.name != "Localhost Domain" {
				t.Errorf("HostIP(%q) = %v, want %v", tt.input, ip.String(), tt.expected)
			}
		})
	}
}

func TestLocalMachineIfsIP2MAC_Integration(t *testing.T) {
	mapping, err := LocalMachineIfsIP2MAC(strategy.And(
		strategy.StandardPhysicalIPv4,
		strategy.ExcludeVirtualIfaces(),
	))
	if err != nil {
		t.Fatalf("LocalMachineIfsIP2MAC execution failed: %v", err)
	}

	for ip, mac := range mapping {
		if net.ParseIP(string(ip)) == nil {
			t.Errorf("Invalid IP key in returned map: %s", ip)
		}
		if string(mac) == "" {
			t.Errorf("Empty MAC address value for IP: %s", ip)
		}
	}
}

func TestCustomNeverMatchStrategy(t *testing.T) {
	neverMatch := func(iface net.Interface, ip net.IP) bool {
		return false
	}

	mapping, err := LocalMachineIfsIP2MAC(neverMatch)
	if err != nil {
		t.Fatalf("Unexpected error with custom strategy: %v", err)
	}

	if len(mapping) != 0 {
		t.Errorf("Expected empty map for neverMatch strategy, got size: %d", len(mapping))
	}
}

func TestLocalFirstIP(t *testing.T) {
	ip, err := LocalFirstIP(strategy.StandardPhysicalIPv4)
	if err != nil {
		t.Fatalf("LocalFirstIP returned error on current system: %v", err)
	}
	t.Logf("LocalFirstIP: %q", ip)
}
