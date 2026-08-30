package tools

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/stephenfire/go-tools/strategy"
)

// StrategyFunc is a type alias for strategy.Func to maintain backwards compatibility
// and avoid circular dependencies between tools and strategy packages.
type StrategyFunc = strategy.Func

// IPString is a custom type for using IP addresses as map keys reliably.
type IPString string

// LocalMachineIfsIP2MAC scans all local network interfaces and extracts a map of IP addresses to MAC addresses
// matching the given strategy function.
//
// Parameters:
//   - fn: Filter strategy callback. Accepts Interface and IP, returns bool. If fn is nil, no filtering is applied.
//
// Returns:
//   - map[IPString]net.HardwareAddr: Map where key is normalized IP string and value is deep-copied HardwareAddr.
//   - error: Returned when reading system network interfaces fails.
func LocalMachineIfsIP2MAC(fn strategy.Func) (map[IPString]net.HardwareAddr, error) {
	// 1. Fetch all local network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get system interfaces: %w", err)
	}

	var macs map[IPString]net.HardwareAddr
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil || len(addrs) == 0 {
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

			if ip == nil {
				continue
			}

			// 2. Evaluate with strategy callback if provided (nil means match all)
			if fn == nil || fn(iface, ip) {
				// Lazy initialization of the result map
				if macs == nil {
					macs = make(map[IPString]net.HardwareAddr)
				}

				// 3. Use ip.String() to normalize map keys across 4-byte and 16-byte IPv4 representations
				ipKey := IPString(ip.String())

				// 4. Perform a deep copy of HardwareAddr for thread safety and memory isolation
				var macCopy net.HardwareAddr
				if len(iface.HardwareAddr) > 0 {
					macCopy = make(net.HardwareAddr, len(iface.HardwareAddr))
					copy(macCopy, iface.HardwareAddr)
				}

				macs[ipKey] = macCopy
			}
		}
	}
	return macs, nil
}

// LocalFirstIP scans local network interfaces sequentially and returns the first IP matching the strategy.
//
// Parameters:
//   - fn: Filter strategy callback. Accepts Interface and IP, returns bool. If fn is nil, no filtering is applied.
//
// Returns:
//   - net.IP: The first matching IP address.
//   - error: Returned when reading interfaces fails or no matching IP is found.
func LocalFirstIP(fn strategy.Func) (net.IP, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get system interfaces: %w", err)
	}

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil || len(addrs) == 0 {
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

			if ip == nil {
				continue
			}

			// Return immediately if fn is nil (no strategy) or if the strategy matches
			if fn == nil || fn(iface, ip) {
				return ip, nil
			}
		}
	}

	return nil, errors.New("no matching IP found")
}

// HostIP parses an input string (Host/IP/CIDR/Host:Port) into a valid net.IP object.
//
// Supported Formats:
//   - Plain IP: "192.168.1.1", "fe80::1"
//   - Address with Port: "192.168.1.1:8080", "example.com:443"
//   - CIDR Notation: "192.168.1.100/24" (Extracts host IP "192.168.1.100", not network address ".0")
//   - Domain Names: "localhost", "example.com" (Triggers DNS lookup, prioritizes IPv4)
//
// Parameters:
//   - hostOrIp: The string to be parsed.
//
// Returns:
//   - net.IP: Resolved standard IP object.
//   - error: Returned if format parsing or DNS lookup fails.
func HostIP(hostOrIp string) (net.IP, error) {
	// 1. Strip port suffix if present (e.g., "192.168.1.100:8080" -> "192.168.1.100")
	host, _, err := net.SplitHostPort(hostOrIp)
	if err != nil {
		host = hostOrIp
	}

	// 2. Direct parsing as standard IP address
	if ip := net.ParseIP(host); ip != nil {
		return ip, nil
	}

	// 3. Parse CIDR format (e.g., "192.168.1.100/24")
	if idx := strings.IndexByte(host, '/'); idx != -1 {
		ipStr := host[:idx]
		if ip := net.ParseIP(ipStr); ip != nil {
			return ip, nil
		}
	}

	// 4. Fallback to DNS resolution for hostnames
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, fmt.Errorf("looking up host (%s) failed: %w", host, err)
	}
	if len(ips) == 0 {
		return nil, errors.New("unknown host: no IP addresses found")
	}

	// 5. Prefer IPv4 address over IPv6
	for _, ip := range ips {
		if ip.To4() != nil {
			return ip, nil
		}
	}

	// Fallback to first resolved IP (e.g., IPv6)
	return ips[0], nil
}
