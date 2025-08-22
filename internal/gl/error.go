package gl

import "strings"

// IsHTTPUnauthorized returns true if the given error is an HTTP 401 Unauthorized error.
func IsHTTPUnauthorized(err error) bool {
	// Check for GitLab-specific 401 unauthorized error patterns
	return strings.Contains(err.Error(), "status code: 401") ||
		strings.Contains(err.Error(), "401 Unauthorized") ||
		strings.Contains(err.Error(), "invalid token")
}

// IsHTTPForbidden returns true if the given error is an HTTP 403 Forbidden error.
func IsHTTPForbidden(err error) bool {
	// Check for GitLab-specific 403 forbidden error patterns
	return strings.Contains(err.Error(), "status code: 403") ||
		strings.Contains(err.Error(), "403 Forbidden") ||
		strings.Contains(err.Error(), "insufficient permissions")
}

// IsHTTPNotFound returns true if the given error is an HTTP 404 Not Found error.
func IsHTTPNotFound(err error) bool {
	// Check for GitLab-specific 404 not found error patterns
	return strings.Contains(err.Error(), "status code: 404") ||
		strings.Contains(err.Error(), "404 Not Found") ||
		strings.Contains(err.Error(), "project not found") ||
		strings.Contains(err.Error(), "merge request not found")
}

// IsRateLimited returns true if the given error indicates rate limiting.
func IsRateLimited(err error) bool {
	// Check for GitLab-specific rate limiting error patterns
	return strings.Contains(err.Error(), "status code: 429") ||
		strings.Contains(err.Error(), "429 Too Many Requests") ||
		strings.Contains(err.Error(), "rate limit")
}