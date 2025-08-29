package gh

import (
	"strings"

	"emperror.dev/errors"
	"github.com/aviator-co/av/internal/provider"
)

// IsHTTPUnauthorized returns true if the given error is an HTTP 401 Unauthorized error.
func IsHTTPUnauthorized(err error) bool {
	// This is a bit fragile because it relies on the error message from the
	// GraphQL package. It doesn't export proper error types so we have to check
	// the string.
	return strings.Contains(err.Error(), "status code: 401")
}

// MapGitHubErrorToProvider maps GitHub API errors to common provider errors
func MapGitHubErrorToProvider(err error) error {
	if err == nil {
		return nil
	}
	
	errStr := err.Error()
	
	// Check for HTTP status codes in the error message
	switch {
	case strings.Contains(errStr, "status code: 401"):
		return errors.Wrap(provider.ErrUnauthorized, err.Error())
	case strings.Contains(errStr, "status code: 403"):
		return errors.Wrap(provider.ErrForbidden, err.Error())
	case strings.Contains(errStr, "status code: 404"):
		// Determine if it's a repository or merge request not found
		if strings.Contains(errStr, "repository") || strings.Contains(errStr, "repo") {
			return errors.Wrap(provider.ErrRepositoryNotFound, err.Error())
		}
		if strings.Contains(errStr, "pull request") || strings.Contains(errStr, "pullrequest") {
			return errors.Wrap(provider.ErrMergeRequestNotFound, err.Error())
		}
		// Default to repository not found for 404s
		return errors.Wrap(provider.ErrRepositoryNotFound, err.Error())
	case strings.Contains(errStr, "status code: 409"):
		return errors.Wrap(provider.ErrConflict, err.Error())
	case strings.Contains(errStr, "status code: 422"):
		// GitHub returns 422 for validation errors, often conflicts
		return errors.Wrap(provider.ErrConflict, err.Error())
	}
	
	// Check for specific GraphQL error messages
	switch {
	case strings.Contains(errStr, "not found"):
		if strings.Contains(errStr, "repository") {
			return errors.Wrap(provider.ErrRepositoryNotFound, err.Error())
		}
		if strings.Contains(errStr, "pull request") {
			return errors.Wrap(provider.ErrMergeRequestNotFound, err.Error())
		}
		return errors.Wrap(provider.ErrRepositoryNotFound, err.Error())
	case strings.Contains(errStr, "already exists"):
		return errors.Wrap(provider.ErrConflict, err.Error())
	case strings.Contains(errStr, "unauthorized") || strings.Contains(errStr, "authentication"):
		return errors.Wrap(provider.ErrUnauthorized, err.Error())
	case strings.Contains(errStr, "forbidden") || strings.Contains(errStr, "permission"):
		return errors.Wrap(provider.ErrForbidden, err.Error())
	}
	
	// Return original error if no mapping is found
	return err
}
