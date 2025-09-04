package provider

import "context"

// GitProvider defines the interface for interacting with Git hosting providers (GitHub, GitLab, etc.)
type GitProvider interface {
	// Pull Request operations
	CreatePullRequest(ctx context.Context, opts CreatePullRequestOpts) (*PullRequest, error)
	UpdatePullRequest(ctx context.Context, opts UpdatePullRequestOpts) (*PullRequest, error)
	GetPullRequests(ctx context.Context, repoSlug string, opts GetPullRequestsOpts) ([]*PullRequest, error)
	PullRequest(ctx context.Context, id string) (*PullRequest, error)
	
	// Repository operations
	GetRepositoryBySlug(ctx context.Context, slug string) (*Repository, error)
	
	// User operations
	User(ctx context.Context, login string) (*User, error)
	Viewer(ctx context.Context) (*Viewer, error)
}

// CreatePullRequestOpts contains options for creating a pull request
type CreatePullRequestOpts struct {
	RepositoryID string
	HeadRefName  string
	BaseRefName  string
	Title        string
	Body         string
	Draft        bool
}

// UpdatePullRequestOpts contains options for updating a pull request
type UpdatePullRequestOpts struct {
	PullRequestID string
	Title         *string
	Body          *string
	Draft         *bool
}

// GetPullRequestsOpts contains options for getting pull requests
type GetPullRequestsOpts struct {
	Head  string
	Base  string
	State string
}