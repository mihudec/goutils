package common

import (
	"strings"
)

func IsValidSHA1(sha1 string) bool {
	if len(sha1) != 40 {
		return false
	}
	sha1 = strings.ToLower(sha1)
	for _, c := range sha1 {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

func IsValidSHA256(sha256 string) bool {
	if len(sha256) != 64 {
		return false
	}
	sha256 = strings.ToLower(sha256)
	for _, c := range sha256 {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}
func IsValidSHA512(sha512 string) bool {
	if len(sha512) != 128 {
		return false
	}
	sha512 = strings.ToLower(sha512)
	for _, c := range sha512 {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}
