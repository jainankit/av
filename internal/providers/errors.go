package providers

import (
	"fmt"
)

// Common error types that all providers should use for consistent error handling

// ErrNotFound indicates that a requested resource was not found
type ErrNotFound struct {
	Resource string
	ID       string
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("%s %q not found", e.Resource, e.ID)
}

func (e ErrNotFound) Is(target error) bool {
	_, ok := target.(*ErrNotFound)
	return ok
}

// ErrUnauthorized indicates that the operation requires authentication or higher privileges
type ErrUnauthorized struct {
	Operation string
	Reason    string
}

func (e ErrUnauthorized) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("unauthorized to perform %s: %s", e.Operation, e.Reason)
	}
	return fmt.Sprintf("unauthorized to perform %s", e.Operation)
}

func (e ErrUnauthorized) Is(target error) bool {
	_, ok := target.(*ErrUnauthorized)
	return ok
}

// ErrRateLimited indicates that the provider API rate limit has been exceeded
type ErrRateLimited struct {
	RetryAfter *int // seconds to wait before retrying, if known
}

func (e ErrRateLimited) Error() string {
	if e.RetryAfter != nil {
		return fmt.Sprintf("rate limited, retry after %d seconds", *e.RetryAfter)
	}
	return "rate limited"
}

func (e ErrRateLimited) Is(target error) bool {
	_, ok := target.(*ErrRateLimited)
	return ok
}

// ErrInvalidInput indicates that the provided input parameters are invalid
type ErrInvalidInput struct {
	Field   string
	Message string
}

func (e ErrInvalidInput) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("invalid input for field %q: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("invalid input: %s", e.Message)
}

func (e ErrInvalidInput) Is(target error) bool {
	_, ok := target.(*ErrInvalidInput)
	return ok
}

// ErrProviderUnsupported indicates that a specific feature is not supported by the provider
type ErrProviderUnsupported struct {
	Provider string
	Feature  string
}

func (e ErrProviderUnsupported) Error() string {
	return fmt.Sprintf("feature %q not supported by provider %q", e.Feature, e.Provider)
}

func (e ErrProviderUnsupported) Is(target error) bool {
	_, ok := target.(*ErrProviderUnsupported)
	return ok
}

// Helper functions for creating common errors

// NewUserNotFound creates a standard user not found error
func NewUserNotFound(login string) error {
	return &ErrNotFound{Resource: "user", ID: login}
}

// NewRepositoryNotFound creates a standard repository not found error
func NewRepositoryNotFound(slug string) error {
	return &ErrNotFound{Resource: "repository", ID: slug}
}

// NewPullRequestNotFound creates a standard pull request not found error
func NewPullRequestNotFound(id string) error {
	return &ErrNotFound{Resource: "pull request", ID: id}
}

// NewTeamNotFound creates a standard team not found error
func NewTeamNotFound(slug string) error {
	return &ErrNotFound{Resource: "team", ID: slug}
}

// NewOrganizationNotFound creates a standard organization not found error
func NewOrganizationNotFound(login string) error {
	return &ErrNotFound{Resource: "organization", ID: login}
}