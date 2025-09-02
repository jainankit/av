package gh

import "github.com/aviator-co/av/internal/provider/github"

// PageInfo contains information about the current/previous/next page of results
// when using paginated APIs.
type PageInfo = github.PageInfo

// Ptr returns a pointer to the argument.
//
// It's a convenience function to make working with the API easier: since Go
// disallows pointers-to-literals, and optional input fields are expressed as
// pointers, this function can be used to easily set optional fields to non-nil
// primitives.
//
// For example, `githubv4.CreatePullRequestInput{Draft: Ptr(true)}`.
func Ptr[T any](v T) *T {
	return github.Ptr(v)
}