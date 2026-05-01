package utils

import (
	"net"
	"net/url"
)

// IsURLSafe checks if a URL is safe to connect to (no SSRF).
func IsURLSafe(rawURL string) (bool, string) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false, "invalid URL"
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false, "unsupported scheme"
	}

	host := u.Hostname()
	blocked := []string{"localhost", "127.0.0.1", "::1", "0.0.0.0", "metadata.google.internal"}
	for _, b := range blocked {
		if host == b {
			return false, "blocked host: " + host
		}
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return false, "DNS lookup failed"
	}
	for _, ip := range ips {
		if ip.IsPrivate() || ip.IsLoopback() {
			return false, "private IP blocked: " + ip.String()
		}
	}
	return true, ""
}
