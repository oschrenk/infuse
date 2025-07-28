package core

import (
	"net/url"
	"regexp"
	"strings"
)

// NormalizeRemoteURL parses a git remote URL and returns normalized components.
// Returns [host, path] for valid URLs, or [rawURL] if parsing fails.
func NormalizeRemoteURL(rawURL string) []string {
	var host, path string
	
	// Handle SSH URLs (git@host:repo format)
	if strings.HasPrefix(rawURL, "git@") {
		sshRegex := regexp.MustCompile(`git@([^:]+):(.+)`)
		matches := sshRegex.FindStringSubmatch(rawURL)
		if len(matches) == 3 {
			host = matches[1]
			path = strings.TrimSuffix(matches[2], ".git")
		}
	} else if strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://") {
		// Handle HTTP/HTTPS URLs
		if u, err := url.Parse(rawURL); err == nil {
			host = u.Host
			path = strings.TrimPrefix(u.Path, "/")
			path = strings.TrimSuffix(path, ".git")
		}
	}
	
	if host != "" && path != "" {
		return []string{host, path}
	}
	
	return []string{rawURL}
}