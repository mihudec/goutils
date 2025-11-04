package common

import (
	"net/netip"
	"reflect"
	"testing"
)

func TestIsValidIPv4(t *testing.T) {
	tests := []struct {
		ip     string
		expect bool
	}{
		{"192.168.1.1", true},
		{"255.255.255.255", true},
		{"0.0.0.0", true},
		{"01.2.3.4", false},
		{"256.1.1.1", false},
		{"1.1.1", false},
		{"1.1.1.1.1", false},
		{"a.b.c.d", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := IsValidIPv4(tt.ip); got != tt.expect {
			t.Errorf("IsValidIPv4(%q) = %v, want %v", tt.ip, got, tt.expect)
		}
	}
}

func TestIsValidIPv4Net(t *testing.T) {
	tests := []struct {
		cidr   string
		expect bool
	}{
		{"192.168.1.0/24", true},
		{"10.0.0.0/32", true},
		{"1.1.1.1/33", false},
		{"1.1.1.1/-1", false},
		{"abc/24", false},
		{"1.1.1", false},
	}

	for _, tt := range tests {
		if got := IsValidIPv4Net(tt.cidr); got != tt.expect {
			t.Errorf("IsValidIPv4Net(%q) = %v, want %v", tt.cidr, got, tt.expect)
		}
	}
}

func TestStringToIP(t *testing.T) {
	tests := []struct {
		input   string
		want    netip.Addr
		wantErr bool
	}{
		{"192.168.0.1", netip.MustParseAddr("192.168.0.1"), false},
		{"::1", netip.MustParseAddr("::1"), false},
		{"192.168.1.1/32", netip.MustParseAddr("192.168.1.1"), false},
		{"2001:db8::/128", netip.MustParseAddr("2001:db8::"), false},
		{"192.168.1.1/24", netip.Addr{}, true},
		{"something", netip.Addr{}, true},
	}

	for _, tt := range tests {
		got, err := StringToIP(tt.input)
		if (err != nil) != tt.wantErr {
			t.Fatalf("StringToIP(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
		}
		if err == nil && got != tt.want {
			t.Fatalf("StringToIP(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestStringToIPPrefix(t *testing.T) {
	tests := []struct {
		input   string
		want    netip.Prefix
		wantErr bool
	}{
		{"10.0.0.0/24", netip.MustParsePrefix("10.0.0.0/24"), false},
		{"2001:db8::/32", netip.MustParsePrefix("2001:db8::/32"), false},
		{"192.168.1.1", netip.MustParsePrefix("192.168.1.1/32"), false},
		{"::1", netip.MustParsePrefix("::1/128"), false},
		{"broken", netip.Prefix{}, true},
	}

	for _, tt := range tests {
		got, err := StringToIPPrefix(tt.input)
		if (err != nil) != tt.wantErr {
			t.Fatalf("StringToIPPrefix(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
		}
		if err == nil && got.String() != tt.want.String() {
			t.Fatalf("StringToIPPrefix(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestStringsToIPs(t *testing.T) {
	input := []string{
		"192.168.1.1",
		"10.0.0.1",
		"192.168.1.1",
		"::1",
		"invalid",
	}
	want := []netip.Addr{
		netip.MustParseAddr("10.0.0.1"),
		netip.MustParseAddr("192.168.1.1"),
		netip.MustParseAddr("::1"),
	}
	got := StringsToIPs(input)

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("StringsToIPs() = %v want %v", got, want)
	}
}

func TestStringsToPrefixes(t *testing.T) {
	input := []string{
		"10.0.0.0/24",
		"10.0.0.0/24",
		"::1/128",
		"invalid",
	}
	want := []netip.Prefix{
		netip.MustParsePrefix("10.0.0.0/24"),
		netip.MustParsePrefix("::1/128"),
	}

	got := StringsToPrefixes(input)

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("StringsToPrefixes() = %v want %v", got, want)
	}
}

func TestCollapsePrefixes(t *testing.T) {
	input := []netip.Prefix{
		netip.MustParsePrefix("10.0.0.0/25"),
		netip.MustParsePrefix("10.0.0.128/25"),
	}
	want := []netip.Prefix{
		netip.MustParsePrefix("10.0.0.0/24"),
	}

	got := CollapsePrefixes(input)

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("CollapsePrefixes() = %v want %v", got, want)
	}
}

func TestIPPrefixToString(t *testing.T) {
	tests := []struct {
		input netip.Prefix
		want  string
	}{
		{netip.MustParsePrefix("1.1.1.1/32"), "1.1.1.1"},
		{netip.MustParsePrefix("1.1.0.0/16"), "1.1.0.0/16"},
		{netip.MustParsePrefix("::1/128"), "::1"},
		{netip.MustParsePrefix("2001:db8::/64"), "2001:db8::/64"},
	}

	for _, tt := range tests {
		if got := IPPrefixToString(tt.input); got != tt.want {
			t.Errorf("IPPrefixToString(%v) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
