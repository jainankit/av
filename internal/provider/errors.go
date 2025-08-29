package provider

import "emperror.dev/errors"

// Common error types that abstract provider-specific errors.
var (
	// ErrMergeRequestNotFound is returned when a merge request cannot be found.
	ErrMergeRequestNotFound = errors.New("merge request not found")

	// ErrRepositoryNotFound is returned when a repository cannot be found.
	ErrRepositoryNotFound = errors.New("repository not found")

	// ErrUnauthorized is returned when the user lacks permission to perform an operation.
	ErrUnauthorized = errors.New("unauthorized: insufficient permissions")

	// ErrForbidden is returned when the operation is forbidden.
	ErrForbidden = errors.New("forbidden: operation not allowed")

	// ErrRateLimited is returned when the API rate limit is exceeded.
	ErrRateLimited = errors.New("rate limit exceeded")

	// ErrInvalidInput is returned when the provided input is invalid.
	ErrInvalidInput = errors.New("invalid input")

	// ErrConflict is returned when there's a conflict with the current state.
	ErrConflict = errors.New("conflict with current state")

	// ErrNetworkError is returned when there's a network-related error.
	ErrNetworkError = errors.New("network error")

	// ErrProviderUnavailable is returned when the provider service is unavailable.
	ErrProviderUnavailable = errors.New("provider service unavailable")

	// ErrInvalidCredentials is returned when the provided credentials are invalid.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrUnsupportedOperation is returned when an operation is not supported by the provider.
	ErrUnsupportedOperation = errors.New("operation not supported by provider")
)

// ProviderError represents a provider-specific error with additional context.
type ProviderError struct {
	// Provider is the name of the provider that generated the error.
	Provider ProviderType
	// Operation is the operation that was being performed when the error occurred.
	Operation string
	// StatusCode is the HTTP status code (if applicable).
	StatusCode int
	// Err is the underlying error.
	Err error
}

// Error implements the error interface.
func (e *ProviderError) Error() string {
	if e.Operation != "" {
		return errors.Errorf("provider %s operation %s failed: %v", e.Provider, e.Operation, e.Err).Error()
	}
	return errors.Errorf("provider %s error: %v", e.Provider, e.Err).Error()
}

// Unwrap returns the underlying error.
func (e *ProviderError) Unwrap() error {
	return e.Err
}

// Is implements error matching for common error types.
func (e *ProviderError) Is(target error) bool {
	return errors.Is(e.Err, target)
}

// NewProviderError creates a new ProviderError.
func NewProviderError(provider ProviderType, operation string, statusCode int, err error) *ProviderError {
	return &ProviderError{
		Provider:   provider,
		Operation:  operation,
		StatusCode: statusCode,
		Err:        err,
	}
}

// IsNotFound checks if an error indicates a "not found" condition.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrMergeRequestNotFound) || errors.Is(err, ErrRepositoryNotFound)
}

// IsUnauthorized checks if an error indicates an authorization failure.
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized) || errors.Is(err, ErrInvalidCredentials)
}

// IsForbidden checks if an error indicates a forbidden operation.
func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}

// IsRateLimited checks if an error indicates rate limiting.
func IsRateLimited(err error) bool {
	return errors.Is(err, ErrRateLimited)
}

// IsConflict checks if an error indicates a conflict.
func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict)
}

// IsNetworkError checks if an error is network-related.
func IsNetworkError(err error) bool {
	return errors.Is(err, ErrNetworkError) || errors.Is(err, ErrProviderUnavailable)
}