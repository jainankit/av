package gl

import "strings"

// IsHTTPUnauthorized returns true if the given error is an HTTP 401 Unauthorized error.
func IsHTTPUnauthorized(err error) bool {
	// Check for GitLab API 401 errors in the error message
	return strings.Contains(err.Error(), "status 401") || 
		   strings.Contains(err.Error(), "Unauthorized")
}

// IsHTTPNotFound returns true if the given error is an HTTP 404 Not Found error.
func IsHTTPNotFound(err error) bool {
	// Check for GitLab API 404 errors in the error message
	return strings.Contains(err.Error(), "status 404") || 
		   strings.Contains(err.Error(), "Not Found")
}

// IsHTTPForbidden returns true if the given error is an HTTP 403 Forbidden error.
func IsHTTPForbidden(err error) bool {
	// Check for GitLab API 403 errors in the error message
	return strings.Contains(err.Error(), "status 403") || 
		   strings.Contains(err.Error(), "Forbidden")
}