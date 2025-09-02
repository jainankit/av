package provider

import (
	"context"
)

// GitProvider defines the interface for interacting with Git hosting providers
// like GitHub and GitLab. It abstracts common operations across different
// platforms while maintaining consistent API patterns.
type GitProvider interface {
	// Pull Request Operations
	CreatePullRequest(ctx context.Context, input CreatePullRequestInput) (*PullRequest, error)
	UpdatePullRequest(ctx context.Context, input UpdatePullRequestInput) (*PullRequest, error)
	GetPullRequests(ctx context.Context, input GetPullRequestsInput) (*GetPullRequestsPage, error)

	// Repository Operations
	GetRepositoryBySlug(ctx context.Context, slug string) (*Repository, error)

	// User Operations
	User(ctx context.Context, login string) (*User, error)
	Viewer(ctx context.Context) (*Viewer, error)

	// Provider Identification
	Type() ProviderType
}

// CreatePullRequestInput contains the parameters for creating a pull request
type CreatePullRequestInput struct {
	RepositoryID string
	Title        string
	Body         string
	HeadRefName  string
	BaseRefName  string
	IsDraft      bool
}

// UpdatePullRequestInput contains the parameters for updating a pull request
type UpdatePullRequestInput struct {
	PullRequestID string
	Title         *string
	Body          *string
	IsDraft       *bool
}

// GetPullRequestsInput contains the parameters for querying pull requests
type GetPullRequestsInput struct {
	Owner       string
	Repo        string
	HeadRefName string
	BaseRefName string
	States      []PullRequestState
	First       int32
	After       string
}

// GetPullRequestsPage represents a paginated response of pull requests
type GetPullRequestsPage struct {
	PageInfo     PageInfo
	PullRequests []PullRequest
}

// PageInfo represents pagination information
type PageInfo struct {
	HasNextPage     bool
	HasPreviousPage bool
	StartCursor     string
	EndCursor       string
}