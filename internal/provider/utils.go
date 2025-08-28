package provider

import (
	"strings"
)

// Ptr returns a pointer to the argument.
// This is a convenience function for working with optional fields in API calls.
func Ptr[T any](v T) *T {
	return &v
}

// Nullable returns a pointer to the argument if it's not the zero value,
// otherwise it returns nil. This is useful for translating between Go's
// "unset is zero" and APIs that distinguish between unset (null) and zero values.
func Nullable[T comparable](v T) *T {
	var zero T
	if v == zero {
		return nil
	}
	return &v
}

// NormalizeBranchName removes common prefixes from branch names
func NormalizeBranchName(branchName string) string {
	// Remove refs/heads/ prefix if present
	const prefix = "refs/heads/"
	return strings.TrimPrefix(branchName, prefix)
}

// ParseRepositorySlug parses owner/repo from a repository slug
func ParseRepositorySlug(slug string) (owner, repo string, ok bool) {
	parts := strings.Split(slug, "/")
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}