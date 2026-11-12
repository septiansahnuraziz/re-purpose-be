package utils

import (
	"net"
	"net/url"
	"regexp"
	"strings"
)

var whitespacePattern = regexp.MustCompile(`\s+`)

// NormalizeWhitespace collapses consecutive whitespace into a single space.
func NormalizeWhitespace(value string) string {
	return strings.TrimSpace(whitespacePattern.ReplaceAllString(value, " "))
}

// WordCount counts whitespace-separated words in a text.
func WordCount(value string) int {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0
	}
	return len(strings.Fields(trimmed))
}

// IsHTTPURL reports whether the value is a well-formed absolute http(s) URL.
func IsHTTPURL(value string) bool {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(value))
	if err != nil {
		return false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}
	return parsed.Host != ""
}

// IsPrivateHost is a minimal SSRF guard: it rejects localhost and private IP literals.
// TODO: resolve DNS and validate every resolved IP before making the outbound request.
func IsPrivateHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" || host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return true
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsUnspecified()
}
