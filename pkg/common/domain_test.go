package common

import (
	"reflect"
	"strings"
	"testing"
)

func TestIsValidDomain(t *testing.T) {
	t.Parallel()

	label63 := strings.Repeat("a", 63)
	validDomain := strings.Join([]string{
		label63,
		strings.Repeat("b", 63),
		strings.Repeat("c", 63),
		strings.Repeat("d", 61),
	}, ".")
	tooLongLabel := strings.Repeat("a", 64)
	tooLongDomain := strings.Join([]string{
		label63,
		strings.Repeat("b", 63),
		strings.Repeat("c", 63),
		strings.Repeat("d", 62),
	}, ".")

	tests := []struct {
		name   string
		domain string
		want   bool
	}{
		{name: "simple domain", domain: "example.com", want: true},
		{name: "multiple labels", domain: "sub.domain.example", want: true},
		{name: "maximum length", domain: validDomain, want: true},
		{name: "empty", domain: "", want: false},
		{name: "single label", domain: "localhost", want: false},
		{name: "label too long", domain: tooLongLabel + ".com", want: false},
		{name: "domain too long", domain: tooLongDomain, want: false},
		{name: "invalid characters", domain: "exa$mple.com", want: false},
		{name: "starts with hyphen", domain: "-example.com", want: false},
		{name: "uppercase rejected", domain: "Example.com", want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsValidDomain(tt.domain); got != tt.want {
				t.Fatalf("IsValidDomain(%q) = %v, want %v", tt.domain, got, tt.want)
			}
		})
	}
}

func TestValidateDomainNames(t *testing.T) {
	t.Parallel()

	input := []string{
		" Example.com ",
		"foo_bar.com",
		"duplicate.com",
		"duplicate.com",
		"-invalid.com",
		"",
		"TOO-LONG-LABEL-" + strings.Repeat("a", 63) + ".com",
	}

	valid, invalid := ValidateDomainNames(input)

	wantValid := []string{"example.com", "foo_bar.com", "duplicate.com"}
	wantInvalid := []string{"-invalid.com", "too-long-label-" + strings.Repeat("a", 63) + ".com"}

	if !reflect.DeepEqual(valid, wantValid) {
		t.Fatalf("ValidateDomainNames valid = %v, want %v", valid, wantValid)
	}

	if !reflect.DeepEqual(invalid, wantInvalid) {
		t.Fatalf("ValidateDomainNames invalid = %v, want %v", invalid, wantInvalid)
	}
}

func TestValidateDomainNamesOrIp(t *testing.T) {
	t.Parallel()

	input := []string{
		"example.com",
		"192.168.1.1",
		"::1",
		"invalid_domain",
		"EXAMPLE.COM",
		"192.168.1.1",
	}

	valid, invalid := ValidateDomainNamesOrIp(input)

	wantValid := []string{"example.com", "192.168.1.1", "::1"}
	wantInvalid := []string{"invalid_domain"}

	if !reflect.DeepEqual(valid, wantValid) {
		t.Fatalf("ValidateDomainNamesOrIp valid = %v, want %v", valid, wantValid)
	}

	if !reflect.DeepEqual(invalid, wantInvalid) {
		t.Fatalf("ValidateDomainNamesOrIp invalid = %v, want %v", invalid, wantInvalid)
	}
}

func TestGenerateDomainPairs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		domain string
		want   [][2]string
	}{
		{
			name:   "three labels",
			domain: "a.b.example",
			want: [][2]string{
				{"a.b.example", "b.example"},
				{"b.example", "example"},
			},
		},
		{
			name:   "two labels",
			domain: "example.com",
			want: [][2]string{
				{"example.com", "com"},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := GenerateDomainPairs(tt.domain); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("GenerateDomainPairs(%q) = %v, want %v", tt.domain, got, tt.want)
			}
		})
	}
}

func TestGenerateSubdomainPairsPSL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		domain string
		want   [][2]string
	}{
		{
			name:   "stops at eTLD+1",
			domain: "a.b.example.co.uk",
			want: [][2]string{
				{"a.b.example.co.uk", "b.example.co.uk"},
				{"b.example.co.uk", "example.co.uk"},
			},
		},
		{
			name:   "invalid domain",
			domain: "localhost",
			want:   nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := GenerateSubdomainPairsPSL(tt.domain); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("GenerateSubdomainPairsPSL(%q) = %v, want %v", tt.domain, got, tt.want)
			}
		})
	}
}
