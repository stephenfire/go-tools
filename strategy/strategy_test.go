package strategy

import (
	"net"
	"testing"
)

// Helper function to create mock network interface and IP.
func makeMockIfaceAndIP(name string, flags net.Flags, ipStr string) (net.Interface, net.IP) {
	iface := net.Interface{
		Index: 1,
		Name:  name,
		Flags: flags,
	}
	ip := net.ParseIP(ipStr)
	return iface, ip
}

func TestAtomicStrategies(t *testing.T) {
	upIface, ipv4 := makeMockIfaceAndIP("eth0", net.FlagUp, "192.168.1.100")
	downIface, _ := makeMockIfaceAndIP("eth1", 0, "10.0.0.1")
	loopIface, loopIP := makeMockIfaceAndIP("lo", net.FlagUp|net.FlagLoopback, "127.0.0.1")
	_, ipv6 := makeMockIfaceAndIP("eth0", net.FlagUp, "fe80::1")

	// Test IsUp
	if !IsUp(upIface, ipv4) {
		t.Errorf("IsUp expected true, got false")
	}
	if IsUp(downIface, ipv4) {
		t.Errorf("IsUp expected false, got true")
	}

	// Test IsIPv4 & IsIPv6
	if !IsIPv4(upIface, ipv4) || IsIPv6(upIface, ipv4) {
		t.Errorf("IsIPv4 evaluation failed for IPv4 address")
	}
	if !IsIPv6(upIface, ipv6) || IsIPv4(upIface, ipv6) {
		t.Errorf("IsIPv6 evaluation failed for IPv6 address")
	}

	// Test IsLoopback
	if !IsLoopback(loopIface, loopIP) {
		t.Errorf("IsLoopback expected true for loopback interface")
	}
	if IsLoopback(upIface, ipv4) {
		t.Errorf("IsLoopback expected false for physical interface")
	}
}

func TestCompositeAndFactories(t *testing.T) {
	physIface, privateIPv4 := makeMockIfaceAndIP("ens33", net.FlagUp, "192.168.1.50")
	_, publicIPv4 := makeMockIfaceAndIP("ens33", net.FlagUp, "8.8.8.8")
	dockerIface, _ := makeMockIfaceAndIP("docker0", net.FlagUp, "172.17.0.1")

	// Test PrivateIPv4 and PublicGlobalIPv4
	if !PrivateIPv4(physIface, privateIPv4) {
		t.Errorf("PrivateIPv4 failed to identify 192.168.1.50")
	}
	if PublicGlobalIPv4(physIface, privateIPv4) {
		t.Errorf("PublicGlobalIPv4 should not match private IP")
	}
	if !PublicGlobalIPv4(physIface, publicIPv4) {
		t.Errorf("PublicGlobalIPv4 failed to identify 8.8.8.8")
	}

	// Test FilterByInterfacePrefix
	ethFilter := FilterByInterfacePrefix("eth", "ens")
	if !ethFilter(physIface, privateIPv4) {
		t.Errorf("FilterByInterfacePrefix expected match for ens33")
	}
	if ethFilter(dockerIface, privateIPv4) {
		t.Errorf("FilterByInterfacePrefix should not match docker0")
	}

	// Test ExcludeVirtualIfaces
	virtFilter := ExcludeVirtualIfaces()
	if !virtFilter(physIface, privateIPv4) {
		t.Errorf("ExcludeVirtualIfaces should not exclude ens33")
	}
	if virtFilter(dockerIface, privateIPv4) {
		t.Errorf("ExcludeVirtualIfaces should exclude docker0")
	}
}

func TestCombinators(t *testing.T) {
	upIface, ipv4 := makeMockIfaceAndIP("eth0", net.FlagUp, "192.168.1.100")
	loopIface, loopIP := makeMockIfaceAndIP("lo", net.FlagUp|net.FlagLoopback, "127.0.0.1")

	// Test And & Not combinators
	customAnd := And(
		IsUp,
		Not(IsLoopback),
		IsIPv4,
	)

	if !customAnd(upIface, ipv4) {
		t.Errorf("And combinator failed for valid physical interface")
	}
	if customAnd(loopIface, loopIP) {
		t.Errorf("And combinator failed to exclude loopback interface")
	}

	// Test Or combinator
	customOr := Or(
		IsLoopback,
		FilterByInterfacePrefix("eth"),
	)
	if !customOr(upIface, ipv4) {
		t.Errorf("Or combinator expected match for eth prefix")
	}
	if !customOr(loopIface, loopIP) {
		t.Errorf("Or combinator expected match for loopback")
	}
}
