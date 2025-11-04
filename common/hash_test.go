package common

import (
	"strings"
	"testing"
)

func TestIsValidSHA1(t *testing.T) {
	t.Parallel()

	valid := strings.Repeat("a", 40)
	invalidLength := strings.Repeat("a", 39)
	invalidCharacter := strings.Repeat("a", 39) + "g"

	tests := []struct {
		name string
		hash string
		want bool
	}{
		{name: "valid", hash: valid, want: true},
		{name: "uppercase allowed", hash: strings.ToUpper(valid), want: true},
		{name: "invalid length", hash: invalidLength, want: false},
		{name: "invalid character", hash: invalidCharacter, want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsValidSHA1(tt.hash); got != tt.want {
				t.Fatalf("IsValidSHA1(%q) = %v, want %v", tt.hash, got, tt.want)
			}
		})
	}
}

func TestIsValidSHA256(t *testing.T) {
	t.Parallel()

	valid := strings.Repeat("b", 64)
	invalidLength := strings.Repeat("b", 65)
	invalidCharacter := strings.Repeat("b", 63) + "z"

	tests := []struct {
		name string
		hash string
		want bool
	}{
		{name: "valid", hash: valid, want: true},
		{name: "uppercase allowed", hash: strings.ToUpper(valid), want: true},
		{name: "invalid length", hash: invalidLength, want: false},
		{name: "invalid character", hash: invalidCharacter, want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsValidSHA256(tt.hash); got != tt.want {
				t.Fatalf("IsValidSHA256(%q) = %v, want %v", tt.hash, got, tt.want)
			}
		})
	}
}

func TestIsValidSHA512(t *testing.T) {
	t.Parallel()

	valid := strings.Repeat("c", 128)
	invalidLength := strings.Repeat("c", 127)
	invalidCharacter := strings.Repeat("c", 127) + "x"

	tests := []struct {
		name string
		hash string
		want bool
	}{
		{name: "valid", hash: valid, want: true},
		{name: "uppercase allowed", hash: strings.ToUpper(valid), want: true},
		{name: "invalid length", hash: invalidLength, want: false},
		{name: "invalid character", hash: invalidCharacter, want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsValidSHA512(tt.hash); got != tt.want {
				t.Fatalf("IsValidSHA512(%q) = %v, want %v", tt.hash, got, tt.want)
			}
		})
	}
}
