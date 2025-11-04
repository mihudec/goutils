package common

import (
	"regexp"
	"strings"

	"golang.org/x/net/publicsuffix"
)

func IsValidDomain(domain string) bool {
	var labelRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-_]{0,61}[a-z0-9])?$`)
	if len(domain) == 0 || len(domain) > 253 {
		return false
	}

	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return false
	}

	for _, label := range parts {
		if len(label) == 0 || len(label) > 63 {
			return false
		}
		if !labelRegex.MatchString(label) {
			return false
		}
	}

	return true
}

func ValidateDomainNames(domains []string) ([]string, []string) {
	var validDomains, invalidDomains []string
	seen := make(map[string]struct{})

	for _, raw := range domains {
		domain := strings.ToLower(strings.TrimSpace(raw))
		if domain == "" {
			continue
		}
		if _, exists := seen[domain]; exists {
			continue
		}
		seen[domain] = struct{}{}

		if IsValidDomain(domain) {
			validDomains = append(validDomains, domain)
		} else {
			invalidDomains = append(invalidDomains, domain)
		}
	}

	return validDomains, invalidDomains
}

func ValidateDomainNamesOrIp(domains []string) ([]string, []string) {
	var validDomains, invalidDomains []string
	seen := make(map[string]struct{})

	for _, raw := range domains {
		domain := strings.ToLower(strings.TrimSpace(raw))
		if domain == "" {
			continue
		}
		if _, exists := seen[domain]; exists {
			continue
		}
		seen[domain] = struct{}{}

		if IsValidDomain(domain) {
			validDomains = append(validDomains, domain)
		} else if IsValidIPv4(domain) || IsValidIPv6(domain) {
			// If it's a valid IP address, we consider it valid for this context
			validDomains = append(validDomains, domain)
		} else {
			invalidDomains = append(invalidDomains, domain)
		}
	}

	return validDomains, invalidDomains
}

// GenerateSubdomainPairs returns child → parent domain pairs
func GenerateDomainPairs(domain string) [][2]string {
	domain = strings.ToLower(strings.TrimSpace(domain))
	parts := strings.Split(domain, ".")

	// max pairs = len(parts)-1
	n := len(parts) - 1
	result := make([][2]string, 0, n)

	for i := range n { // Go 1.22: range over int
		child := strings.Join(parts[i:], ".")
		parent := strings.Join(parts[i+1:], ".")
		result = append(result, [2]string{child, parent})
	}

	return result
}

// GenerateSubdomainPairsPSL returns child→parent pairs but stops at the PSL boundary.
func GenerateSubdomainPairsPSL(domain string) [][2]string {
	domain = strings.ToLower(strings.TrimSpace(domain))
	parts := strings.Split(domain, ".")

	// Determine the "registrable domain" (eTLD+1)
	eTLDPlusOne, err := publicsuffix.EffectiveTLDPlusOne(domain)
	if err != nil {
		// Invalid domain
		return nil
	}

	publicSuffix, _ := publicsuffix.PublicSuffix(domain)

	// Number of "hops" we can do = len(parts)-1, but we stop earlier if needed.
	n := len(parts) - 1

	result := make([][2]string, 0, n)

	for i := range n {
		child := strings.Join(parts[i:], ".")
		parent := strings.Join(parts[i+1:], ".")

		result = append(result, [2]string{child, parent})

		// Stop once we reach registrable domain or public suffix
		if parent == eTLDPlusOne || parent == publicSuffix {
			break
		}
	}

	return result
}
