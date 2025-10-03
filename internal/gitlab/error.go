package gitlab

import (
	"errors"
	"net/http"
	"strings"
)

// Common GitLab API error types
var (
	ErrUnauthorized = errors.New(
		"gitlab: unauthorized - check your access token",
	)
	ErrForbidden = errors.New(
		"gitlab: forbidden - insufficient permissions",
	)
	ErrNotFound        = errors.New("gitlab: resource not found")
	ErrRateLimited     = errors.New("gitlab: rate limit exceeded")
	ErrInternalServer  = errors.New("gitlab: internal server error")
	ErrBadGateway      = errors.New("gitlab: bad gateway")
	ErrServiceUnavail  = errors.New("gitlab: service unavailable")
	ErrGatewayTimeout  = errors.New("gitlab: gateway timeout")
)

// IsHTTPUnauthorized returns true if the given error is an HTTP 401
// Unauthorized error.
func IsHTTPUnauthorized(err error) bool {
	if err == nil {
		return false
	}
	
	// Check for our wrapped error
	if errors.Is(err, ErrUnauthorized) {
		return true
	}
	
	// Check error message for HTTP status codes
	errStr := err.Error()
	return strings.Contains(errStr, "status code: 401") ||
		strings.Contains(errStr, "401 Unauthorized") ||
		strings.Contains(errStr, "HTTP 401")
}

// IsHTTPForbidden returns true if the given error is an HTTP 403 Forbidden
// error.
func IsHTTPForbidden(err error) bool {
	if err == nil {
		return false
	}
	
	// Check for our wrapped error
	if errors.Is(err, ErrForbidden) {
		return true
	}
	
	// Check error message for HTTP status codes
	errStr := err.Error()
	return strings.Contains(errStr, "status code: 403") ||
		strings.Contains(errStr, "403 Forbidden") ||
		strings.Contains(errStr, "HTTP 403")
}

// IsHTTPNotFound returns true if the given error is an HTTP 404 Not Found
// error.
func IsHTTPNotFound(err error) bool {
	if err == nil {
		return false
	}
	
	// Check for our wrapped error
	if errors.Is(err, ErrNotFound) {
		return true
	}
	
	// Check error message for HTTP status codes
	errStr := err.Error()
	return strings.Contains(errStr, "status code: 404") ||
		strings.Contains(errStr, "404 Not Found") ||
		strings.Contains(errStr, "HTTP 404")
}

// IsHTTPRateLimited returns true if the given error is an HTTP 429 Too Many
// Requests error.
func IsHTTPRateLimited(err error) bool {
	if err == nil {
		return false
	}
	
	// Check for our wrapped error
	if errors.Is(err, ErrRateLimited) {
		return true
	}
	
	// Check error message for HTTP status codes
	errStr := err.Error()
	return strings.Contains(errStr, "status code: 429") ||
		strings.Contains(errStr, "429 Too Many Requests") ||
		strings.Contains(errStr, "HTTP 429") ||
		strings.Contains(errStr, "rate limit")
}

// IsHTTPServerError returns true if the given error is an HTTP 5xx server
// error.
func IsHTTPServerError(err error) bool {
	if err == nil {
		return false
	}
	
	// Check for our wrapped errors
	if errors.Is(err, ErrInternalServer) ||
		errors.Is(err, ErrBadGateway) ||
		errors.Is(err, ErrServiceUnavail) ||
		errors.Is(err, ErrGatewayTimeout) {
		return true
	}
	
	// Check error message for HTTP status codes
	errStr := err.Error()
	return strings.Contains(errStr, "status code: 50") ||
		strings.Contains(errStr, "500 Internal Server Error") ||
		strings.Contains(errStr, "502 Bad Gateway") ||
		strings.Contains(errStr, "503 Service Unavailable") ||
		strings.Contains(errStr, "504 Gateway Timeout") ||
		strings.Contains(errStr, "HTTP 50")
}

// IsRetryableError returns true if the error is likely to be temporary and
// the request should be retried.
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}
	
	// Rate limit errors should be retried with backoff
	if IsHTTPRateLimited(err) {
		return true
	}
	
	// Server errors are generally retryable
	if IsHTTPServerError(err) {
		return true
	}
	
	// Network-related errors that might be temporary
	errStr := err.Error()
	return strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "temporary failure") ||
		strings.Contains(errStr, "network is unreachable")
}

// WrapHTTPError wraps an HTTP error with a more user-friendly GitLab-specific
// error.
func WrapHTTPError(statusCode int, originalErr error) error {
	switch statusCode {
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusTooManyRequests:
		return ErrRateLimited
	case http.StatusInternalServerError:
		return ErrInternalServer
	case http.StatusBadGateway:
		return ErrBadGateway
	case http.StatusServiceUnavailable:
		return ErrServiceUnavail
	case http.StatusGatewayTimeout:
		return ErrGatewayTimeout
	default:
		return originalErr
	}
}

// IsGraphQLError checks if the error is a GraphQL-specific error.
// GitLab GraphQL API returns errors in a specific format.
func IsGraphQLError(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := err.Error()
	return strings.Contains(errStr, "graphql:") ||
		strings.Contains(errStr, "GraphQL") ||
		strings.Contains(errStr, "mutation") ||
		strings.Contains(errStr, "query")
}

// ExtractGraphQLErrors attempts to extract meaningful error messages from
// GitLab GraphQL API responses.
func ExtractGraphQLErrors(err error) []string {
	if err == nil {
		return nil
	}
	
	errStr := err.Error()
	var errors []string
	
	// Look for common GitLab GraphQL error patterns
	if strings.Contains(errStr, "does not exist") {
		errors = append(errors, "The requested resource does not exist")
	}
	if strings.Contains(errStr, "access denied") ||
		strings.Contains(errStr, "insufficient permissions") {
		errors = append(errors, "Access denied - check your permissions")
	}
	if strings.Contains(errStr, "validation failed") {
		errors = append(errors, "Validation failed - check your input data")
	}
	
	// If no specific patterns found, return the original error
	if len(errors) == 0 {
		errors = append(errors, errStr)
	}
	
	return errors
}
