package provider

import "emperror.dev/errors"

// Common error types that abstract provider-specific errors

// ErrMergeRequestNotFound indicates that a merge request was not found
var ErrMergeRequestNotFound = errors.New("merge request not found")

// ErrRepositoryNotFound indicates that a repository was not found
var ErrRepositoryNotFound = errors.New("repository not found")

// ErrUnauthorized indicates that the user is not authorized
var ErrUnauthorized = errors.New("unauthorized")

// ErrForbidden indicates that the operation is forbidden
var ErrForbidden = errors.New("forbidden")

// ErrConflict indicates a conflict (e.g., merge request title already exists)
var ErrConflict = errors.New("conflict")

// ErrNotImplemented indicates that a feature is not implemented for this provider
var ErrNotImplemented = errors.New("not implemented")

// IsNotFound returns true if the error indicates a resource was not found
func IsNotFound(err error) bool {
	return errors.Is(err, ErrMergeRequestNotFound) || errors.Is(err, ErrRepositoryNotFound)
}

// IsUnauthorized returns true if the error indicates unauthorized access
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsForbidden returns true if the error indicates forbidden access
func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}

// IsConflict returns true if the error indicates a conflict
func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict)
}

// IsNotImplemented returns true if the error indicates a feature is not implemented
func IsNotImplemented(err error) bool {
	return errors.Is(err, ErrNotImplemented)
}