// Package provider defines interfaces and types for abstracting Git hosting providers.
//
// This package provides a common abstraction layer for working with multiple Git hosting
// providers (GitHub, GitLab, etc.) without coupling the application logic to any specific
// provider's API.
//
// The main interfaces defined in this package are:
//
//   - Provider: Basic provider identification and URL detection
//   - Client: Complete client interface for interacting with provider APIs
//
// Provider-agnostic types include:
//
//   - PullRequest: Represents a pull request (or merge request)
//   - Repository: Represents a repository
//   - User: Represents a user account
//   - Team: Represents a team or group
//   - PullRequestState: Enum for pull request states (open, closed, merged)
//
// The package also provides conversion functions for translating between provider-specific
// types (e.g., githubv4.PullRequestState) and the abstracted types defined here.
package provider
