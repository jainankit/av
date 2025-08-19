package gl

import (
	"fmt"
	"strings"

	"emperror.dev/errors"
)

// IsHTTPUnauthorized returns true if the given error is an HTTP 401 Unauthorized error.
func IsHTTPUnauthorized(err error) bool {
	// This is a bit fragile because it relies on the error message from the
	// GraphQL package. It doesn't export proper error types so we have to check
	// the string.
	return strings.Contains(err.Error(), "status code: 401")
}

// errMergeRequestClosed represents an error when a merge request is closed.
// This will be implemented once MergeRequest struct is available.
type errMergeRequestClosed struct {
	// TODO: Replace with *MergeRequest when merge request types are implemented
	IID   int
	State string
}

func (e errMergeRequestClosed) Error() string {
	return fmt.Sprintf("merge request !%d is %s", e.IID, e.State)
}

// GitLab-specific error constants
var (
	ErrNoGitLabToken        = errors.Sentinel("No GitLab token is set (do you need to configure one?).")
	ErrMergeRequestClosed   = errors.Sentinel("Merge request is closed")
)

// WrapWithContext wraps an error with additional context for GitLab operations.
// This follows the existing pattern used throughout the codebase for error wrapping.
func WrapWithContext(err error, context string) error {
	if err == nil {
		return nil
	}
	return errors.Wrap(err, context)
}

// WrapWithContextf wraps an error with formatted context for GitLab operations.
// This follows the existing pattern used throughout the codebase for error wrapping.
func WrapWithContextf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return errors.Wrapf(err, format, args...)
}