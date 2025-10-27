package provider

import (
	"context"
	"net/url"
)

// ProviderType represents the type of Git hosting provider.
type ProviderType string

const (
	// ProviderTypeGitHub represents GitHub (github.com or GitHub Enterprise).
	ProviderTypeGitHub ProviderType = "github"
	// ProviderTypeGitLab represents GitLab (gitlab.com or self-hosted GitLab).
	ProviderTypeGitLab ProviderType = "gitlab"
)

// Provider represents a Git hosting provider (GitHub, GitLab, etc.).
type Provider interface {
	// Name returns the human-readable name of the provider (e.g., "GitHub", "GitLab").
	Name() string

	// Type returns the provider type identifier.
	Type() ProviderType

	// DetectFromURL determines if this provider can handle the given repository URL.
	DetectFromURL(url *url.URL) bool
}

// Client is the interface that wraps common Git hosting provider operations.
// This interface abstracts operations across different providers (GitHub, GitLab, etc.).
type Client interface {
	Provider

	// CreatePullRequest creates a new pull request (or merge request).
	CreatePullRequest(ctx context.Context, input CreatePullRequestInput) (*PullRequest, error)

	// UpdatePullRequest updates an existing pull request.
	UpdatePullRequest(ctx context.Context, input UpdatePullRequestInput) (*PullRequest, error)

	// GetPullRequest retrieves a pull request by its ID.
	GetPullRequest(ctx context.Context, id string) (*PullRequest, error)

	// ListPullRequests lists pull requests based on the given criteria.
	ListPullRequests(ctx context.Context, input ListPullRequestsInput) (*PullRequestsPage, error)

	// ConvertPullRequestToDraft converts a pull request to draft status.
	ConvertPullRequestToDraft(ctx context.Context, id string) (*PullRequest, error)

	// MarkPullRequestReadyForReview marks a pull request as ready for review.
	MarkPullRequestReadyForReview(ctx context.Context, id string) (*PullRequest, error)

	// RequestReviews requests reviews from users/teams on a pull request.
	RequestReviews(ctx context.Context, input RequestReviewsInput) (*PullRequest, error)

	// GetRepositoryBySlug retrieves repository information by owner/name slug.
	GetRepositoryBySlug(ctx context.Context, slug string) (*Repository, error)

	// GetUser retrieves user information by login.
	GetUser(ctx context.Context, login string) (*User, error)

	// GetTeam retrieves team information by organization and team slug.
	GetTeam(ctx context.Context, organizationLogin, teamSlug string) (*Team, error)

	// GetViewer retrieves information about the currently authenticated user.
	GetViewer(ctx context.Context) (*User, error)
}

// CreatePullRequestInput contains parameters for creating a pull request.
type CreatePullRequestInput struct {
	// RepositoryID is the provider-specific ID of the repository.
	RepositoryID string
	// Title is the title of the pull request.
	Title string
	// Body is the description/body of the pull request.
	Body string
	// HeadRefName is the name of the branch containing changes.
	HeadRefName string
	// BaseRefName is the name of the branch to merge into.
	BaseRefName string
	// Draft indicates whether the pull request should be created as a draft.
	Draft bool
}

// UpdatePullRequestInput contains parameters for updating a pull request.
type UpdatePullRequestInput struct {
	// ID is the provider-specific ID of the pull request.
	ID string
	// Title is the new title (optional, nil means no change).
	Title *string
	// Body is the new body (optional, nil means no change).
	Body *string
	// BaseRefName is the new base branch (optional, nil means no change).
	BaseRefName *string
}

// ListPullRequestsInput contains parameters for listing pull requests.
type ListPullRequestsInput struct {
	// Owner is the repository owner.
	Owner string
	// Repo is the repository name.
	Repo string
	// HeadRefName filters by head branch name (optional).
	HeadRefName string
	// BaseRefName filters by base branch name (optional).
	BaseRefName string
	// States filters by pull request states (optional).
	States []PullRequestState
	// First is the number of results to return (pagination).
	First int32
	// After is the cursor for pagination (optional).
	After string
}

// PullRequestsPage represents a paginated list of pull requests.
type PullRequestsPage struct {
	PullRequests []PullRequest
	PageInfo     PageInfo
}

// PageInfo contains pagination information.
type PageInfo struct {
	EndCursor       string
	HasNextPage     bool
	HasPreviousPage bool
	StartCursor     string
}

// RequestReviewsInput contains parameters for requesting reviews.
type RequestReviewsInput struct {
	// PullRequestID is the provider-specific ID of the pull request.
	PullRequestID string
	// UserIDs is a list of user IDs to request reviews from.
	UserIDs []string
	// TeamIDs is a list of team IDs to request reviews from.
	TeamIDs []string
	// Union indicates whether to add reviewers (true) or replace them (false).
	Union bool
}
