package common

import (
	"fmt"
	"net"
	"path/filepath"
	"strconv"
	"strings"
)

// ValidateIP validates an IPv4 address
func ValidateIP(ip string) error {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return fmt.Errorf("invalid IP address: %s", ip)
	}

	// Ensure it's IPv4
	if parsed.To4() == nil {
		return fmt.Errorf("not a valid IPv4 address: %s", ip)
	}

	return nil
}

// ValidatePort validates a port number (1-65535)
func ValidatePort(port string) error {
	p, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("invalid port number: %s", port)
	}

	if p < 1 || p > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got: %d", p)
	}

	return nil
}

// ValidateCIDR validates an IPv4 CIDR block such as 10.0.0.1/24
func ValidateCIDR(cidr string) error {
	if cidr == "" {
		return fmt.Errorf("CIDR cannot be empty")
	}
	ip, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR: %s", cidr)
	}
	if ip.To4() == nil {
		return fmt.Errorf("CIDR must be IPv4: %s", cidr)
	}
	if network == nil {
		return fmt.Errorf("invalid CIDR network: %s", cidr)
	}
	return nil
}

// ValidatePath validates that a path is absolute
func ValidatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}
	if !filepath.IsAbs(path) {
		return fmt.Errorf("path must be absolute: %s", path)
	}
	return nil
}

// ValidateSafePath validates a path is absolute and contains no shell metacharacters
// This provides defense-in-depth against command injection when paths are used in system commands
func ValidateSafePath(path string) error {
	// First validate it's a valid absolute path
	if err := ValidatePath(path); err != nil {
		return err
	}

	// Check for shell metacharacters that could be exploited
	// Even though we use exec.Command which doesn't use a shell,
	// this provides defense-in-depth protection
	forbiddenChars := []string{
		";",  // Command separator
		"&",  // Background/AND operator
		"|",  // Pipe operator
		"$",  // Variable expansion
		"`",  // Command substitution
		"(",  // Subshell
		")",  // Subshell
		"<",  // Redirection
		">",  // Redirection
		"\n", // Newline
		"\r", // Carriage return
		"*",  // Glob wildcard
		"?",  // Glob wildcard
		"[",  // Glob wildcard
		"]",  // Glob wildcard
		"{",  // Brace expansion
		"}",  // Brace expansion
	}

	for _, char := range forbiddenChars {
		if strings.Contains(path, char) {
			return fmt.Errorf("path contains forbidden shell metacharacter '%s': %s", char, path)
		}
	}

	// Check for null bytes
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("path contains null byte")
	}

	return nil
}

// ValidateUsername validates a Unix username
func ValidateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	// Basic username validation (alphanumeric, underscore, hyphen, must start with letter or underscore)
	if len(username) > 32 {
		return fmt.Errorf("username too long (max 32 characters): %s", username)
	}

	firstChar := username[0]
	if !((firstChar >= 'a' && firstChar <= 'z') || (firstChar >= 'A' && firstChar <= 'Z') || firstChar == '_') {
		return fmt.Errorf("username must start with a letter or underscore: %s", username)
	}

	for _, c := range username {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-') {
			return fmt.Errorf("username contains invalid character: %s", username)
		}
	}

	return nil
}

// ValidateNotEmpty validates that a string is not empty
func ValidateNotEmpty(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("value cannot be empty")
	}
	return nil
}

// ValidateDomain validates a domain name (basic validation)
func ValidateDomain(domain string) error {
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}

	// Basic domain validation - allow alphanumeric, dots, and hyphens
	if len(domain) > 253 {
		return fmt.Errorf("domain name too long: %s", domain)
	}

	parts := strings.Split(domain, ".")
	for _, part := range parts {
		if part == "" {
			return fmt.Errorf("invalid domain (empty label): %s", domain)
		}
		if len(part) > 63 {
			return fmt.Errorf("domain label too long: %s", part)
		}

		for i, c := range part {
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-') {
				return fmt.Errorf("invalid character in domain: %s", domain)
			}
			// Hyphen cannot be at start or end
			if c == '-' && (i == 0 || i == len(part)-1) {
				return fmt.Errorf("domain label cannot start or end with hyphen: %s", part)
			}
		}
	}

	return nil
}

// ValidateWireGuardKey validates a WireGuard public/private key format
// WireGuard keys are base64-encoded, exactly 44 characters, ending with '='
func ValidateWireGuardKey(key string) error {
	if key == "" {
		return fmt.Errorf("WireGuard key cannot be empty")
	}

	// WireGuard keys are always 44 characters (base64-encoded 32 bytes + padding)
	if len(key) != 44 {
		return fmt.Errorf("WireGuard key must be exactly 44 characters, got %d", len(key))
	}

	// Must end with '=' (base64 padding)
	if !strings.HasSuffix(key, "=") {
		return fmt.Errorf("WireGuard key must end with '=' (base64 padding)")
	}

	// Check for valid base64 characters [A-Za-z0-9+/=]
	for i, c := range key {
		isValid := (c >= 'A' && c <= 'Z') ||
			(c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') ||
			c == '+' || c == '/' ||
			(c == '=' && i == len(key)-1) // '=' only valid at the end

		if !isValid {
			return fmt.Errorf("WireGuard key contains invalid character at position %d: '%c'", i, c)
		}
	}

	return nil
}

// ValidateTimezone validates a timezone string (basic check)
func ValidateTimezone(tz string) error {
	if tz == "" {
		return fmt.Errorf("timezone cannot be empty")
	}

	// Basic validation - should contain a slash and reasonable length
	if !strings.Contains(tz, "/") {
		return fmt.Errorf("invalid timezone format (should be Region/City): %s", tz)
	}

	if len(tz) > 64 {
		return fmt.Errorf("timezone string too long: %s", tz)
	}

	return nil
}
