package strategy

import (
	"net"
	"strings"
)

// Func defines the signature for a network interface/IP filtering strategy.
type Func func(iface net.Interface, ip net.IP) bool

// Compile-time type binding checks for IDE indexing and usage tracking.
var (
	_ Func = IsUp
	_ Func = IsIPv4
	_ Func = IsIPv6
	_ Func = IsLoopback
	_ Func = IsGlobalUnicast
	_ Func = StandardPhysicalIPv4
	_ Func = StandardPhysicalIPv6
	_ Func = PublicGlobalIPv4
	_ Func = PrivateIPv4
)

// -----------------------------------------------------------------------------
// Atomic Strategies
// -----------------------------------------------------------------------------

// IsUp checks whether the network interface is up.
func IsUp(iface net.Interface, ip net.IP) bool {
	return iface.Flags&net.FlagUp != 0
}

// IsIPv4 checks whether the IP address is a valid IPv4 address.
func IsIPv4(iface net.Interface, ip net.IP) bool {
	return ip.To4() != nil
}

// IsIPv6 checks whether the IP address is a valid IPv6 address.
func IsIPv6(iface net.Interface, ip net.IP) bool {
	return ip.To4() == nil && ip.To16() != nil
}

// IsLoopback checks whether the interface or IP address is a loopback address.
func IsLoopback(iface net.Interface, ip net.IP) bool {
	return iface.Flags&net.FlagLoopback != 0 || ip.IsLoopback()
}

// IsGlobalUnicast checks whether the IP address is a global unicast address.
func IsGlobalUnicast(iface net.Interface, ip net.IP) bool {
	return ip.IsGlobalUnicast()
}

// -----------------------------------------------------------------------------
// Built-in Composite Strategies
// -----------------------------------------------------------------------------

// StandardPhysicalIPv4 checks for an UP, non-loopback IPv4 address.
func StandardPhysicalIPv4(iface net.Interface, ip net.IP) bool {
	return IsUp(iface, ip) && !IsLoopback(iface, ip) && IsIPv4(iface, ip)
}

// StandardPhysicalIPv6 checks for an UP, non-loopback IPv6 address.
func StandardPhysicalIPv6(iface net.Interface, ip net.IP) bool {
	return IsUp(iface, ip) && !IsLoopback(iface, ip) && IsIPv6(iface, ip)
}

// PublicGlobalIPv4 checks for a public global unicast IPv4 address (excluding private IPs).
func PublicGlobalIPv4(iface net.Interface, ip net.IP) bool {
	return StandardPhysicalIPv4(iface, ip) && ip.IsGlobalUnicast() && !ip.IsPrivate()
}

// PrivateIPv4 checks for a private IPv4 address (e.g., 10.0.0.0/8, 192.168.0.0/16).
func PrivateIPv4(iface net.Interface, ip net.IP) bool {
	return StandardPhysicalIPv4(iface, ip) && ip.IsPrivate()
}

// -----------------------------------------------------------------------------
// Dynamic Strategy Factories
// -----------------------------------------------------------------------------

// FilterByInterfacePrefix returns a strategy that matches interface names with specific prefixes.
func FilterByInterfacePrefix(prefixes ...string) Func {
	return func(iface net.Interface, ip net.IP) bool {
		for _, prefix := range prefixes {
			if strings.HasPrefix(iface.Name, prefix) {
				return true
			}
		}
		return false
	}
}

// ExcludeVirtualIfaces returns a strategy that filters out common virtual network interfaces.
func ExcludeVirtualIfaces() Func {
	virtualPrefixes := []string{"docker", "veth", "flannel", "cni", "virbr", "vboxnet", "tun", "tap"}
	return func(iface net.Interface, ip net.IP) bool {
		name := strings.ToLower(iface.Name)
		for _, vPrefix := range virtualPrefixes {
			if strings.HasPrefix(name, vPrefix) {
				return false
			}
		}
		return true
	}
}

// -----------------------------------------------------------------------------
// Strategy Combinators
// -----------------------------------------------------------------------------

// And combines multiple strategies using logical AND with short-circuit evaluation.
func And(strategies ...Func) Func {
	return func(iface net.Interface, ip net.IP) bool {
		for _, s := range strategies {
			if s != nil && !s(iface, ip) {
				return false
			}
		}
		return true
	}
}

// Or combines multiple strategies using logical OR with short-circuit evaluation.
func Or(strategies ...Func) Func {
	return func(iface net.Interface, ip net.IP) bool {
		for _, s := range strategies {
			if s != nil && s(iface, ip) {
				return true
			}
		}
		return false
	}
}

// Not negates a given strategy using logical NOT.
func Not(s Func) Func {
	return func(iface net.Interface, ip net.IP) bool {
		if s == nil {
			return true
		}
		return !s(iface, ip)
	}
}
