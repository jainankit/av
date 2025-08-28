package gl

import "strings"

// IsHTTPUnauthorized returns true if the given error is an HTTP 401 Unauthorized error.
func IsHTTPUnauthorized(err error) bool {
	// This is a bit fragile because it relies on the error message from the
	// GraphQL package. It doesn't export proper error types so we have to check
	// the string.
	return strings.Contains(err.Error(), "status code: 401")
}

// IsHTTPForbidden returns true if the given error is an HTTP 403 Forbidden error.
// GitLab may return 403 for insufficient permissions on projects/groups.
func IsHTTPForbidden(err error) bool {
	return strings.Contains(err.Error(), "status code: 403")
}

// IsHTTPNotFound returns true if the given error is an HTTP 404 Not Found error.
// GitLab may return 404 for projects that don't exist or are not accessible.
func IsHTTPNotFound(err error) bool {
	return strings.Contains(err.Error(), "status code: 404")
}