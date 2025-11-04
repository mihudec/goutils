package common

import (
	"fmt"
	"go4.org/netipx"
	"net/netip"
	"sort"
	"strconv"
	"strings"
)

func IsValidIPv4(input string) bool {
	ip, err := netip.ParseAddr(input)
	if err != nil {
		return false
	}
	return ip.Is4()
}

func IsValidIPv4Net(ipnet string) bool {
	parts := strings.Split(ipnet, "/")
	if len(parts) != 2 {
		return false
	}
	ip := parts[0]
	prefixLenStr := parts[1]

	if !IsValidIPv4(ip) {
		return false
	}

	prefixLen, err := strconv.Atoi(prefixLenStr)
	if err != nil || prefixLen < 0 || prefixLen > 32 {
		return false
	}

	return true
}

func IsValidIPv6(input string) bool {
	ip, err := netip.ParseAddr(input)
	if err != nil {
		return false
	}
	return ip.Is6()
}

func StringToIP(input string) (netip.Addr, error) {
	// Try parsing as plain IP
	if addr, err := netip.ParseAddr(input); err == nil {
		return addr, nil
	}

	// Try parsing as prefix and accept /32 or /128 only
	if pfx, err := netip.ParsePrefix(input); err == nil {
		bits := pfx.Bits()
		if (pfx.Addr().Is4() && bits == 32) || (!pfx.Addr().Is4() && bits == 128) {
			return pfx.Addr(), nil
		}
		return netip.Addr{}, fmt.Errorf("prefix is not a single IP: %s", input)
	}

	return netip.Addr{}, fmt.Errorf("invalid IP or host prefix: %s", input)
}

func StringToIPPrefix(input string) (netip.Prefix, error) {
	// Try parsing as CIDR first
	if pfx, err := netip.ParsePrefix(input); err == nil {
		return pfx, nil
	}

	// Try parsing as IP and convert to /32 or /128 prefix
	if addr, err := netip.ParseAddr(input); err == nil {
		var bits int
		if addr.Is4() {
			bits = 32
		} else {
			bits = 128
		}
		return netip.PrefixFrom(addr, bits), nil
	}

	return netip.Prefix{}, fmt.Errorf("invalid prefix or IP: %s", input)
}

func StringsToIPs(input []string) []netip.Addr {
	output := make([]netip.Addr, 0)
	seen := make(map[string]struct{})

	for _, v := range input {
		ip, err := StringToIP(v)
		if err != nil {
			continue
		}
		key := ip.String()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		output = append(output, ip)
	}

	sort.Slice(output, func(i, j int) bool {
		a, b := output[i], output[j]

		if a.Is4() && !b.Is4() {
			return true
		}
		if !a.Is4() && b.Is4() {
			return false
		}
		return a.Less(b)
	})

	return output
}

func StringsToPrefixes(input []string) []netip.Prefix {
	output := make([]netip.Prefix, 0)
	seen := make(map[string]struct{})

	for _, v := range input {
		pfx, err := StringToIPPrefix(v)
		if err != nil {
			continue
		}
		key := pfx.String()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		output = append(output, pfx)
	}

	sort.Slice(output, func(i, j int) bool {
		a, b := output[i], output[j]

		if a.Addr().Is4() && !b.Addr().Is4() {
			return true
		}
		if !a.Addr().Is4() && b.Addr().Is4() {
			return false
		}
		if a.Addr().Compare(b.Addr()) != 0 {
			return a.Addr().Less(b.Addr())
		}
		return a.Bits() < b.Bits()

	})

	return output
}

func CollapsePrefixes(input []netip.Prefix) []netip.Prefix {
	if len(input) == 0 {
		return nil
	}
	var b netipx.IPSetBuilder
	for _, p := range input {
		b.AddPrefix(p)
	}

	ipset, err := b.IPSet()
	if err != nil {
		// You may want to handle/log the error
		return nil
	}

	return ipset.Prefixes()
}

func IPPrefixToString(p netip.Prefix) string {
	if (p.Addr().Is4() && p.Bits() == 32) || (p.Addr().Is6() && p.Bits() == 128) {
		return p.Addr().String()
	}
	return p.String() // CIDR notation
}
