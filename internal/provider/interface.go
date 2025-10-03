package provider

import (
	"context"
)

// ClientInterface defines the common interface that both GitHub and GitLab
// clients should implement. This allows for provider-agnostic operations
// where the same code can work with either provider.
//
// Note: This is an optional abstraction for future extensibility. The current
// implementation may choose to use provider-specific clients directly.
type ClientInterface interface {
	// GetUser returns information about the authenticated user
	GetUser(ctx context.Context) (*User, error)

	// GetRepository returns information about a repository
	GetRepository(ctx context.Context, repoSlug string) (*Repository, error)

	// CreatePullRequest creates a new pull request or merge request
	CreatePullRequest(ctx context.Context, input *CreatePullRequestInput) (*PullRequest, error)

	// GetPullRequest retrieves a pull request or merge request by ID
	GetPullRequest(ctx context.Context, repoSlug string, number int) (*PullRequest, error)

	// UpdatePullRequest updates an existing pull request or merge request
	UpdatePullRequest(ctx context.Context, input *UpdatePullRequestInput) (*PullRequest, error)

	// ListPullRequests lists pull requests or merge requests for a repository
	ListPullRequests(
		ctx context.Context,
		repoSlug string,
		options *ListOptions,
	) ([]*PullRequest, error)
}

// User represents a user account across providers
type User struct {
	ID        string
	Username  string
	Name      string
	Email     string
	AvatarURL string
	HTMLURL   string
}

// Repository represents a repository across providers
type Repository struct {
	ID            string
	Slug          string
	Name          string
	Description   string
	HTMLURL       string
	CloneURL      string
	SSHCloneURL   string
	DefaultBranch string
	IsPrivate     bool
	IsFork        bool
}

// PullRequest represents a pull request or merge request across providers
type PullRequest struct {
	ID       string
	Number   int
	Title    string
	Body     string
	State    PullRequestState
	HTMLURL  string

	// Branch information
	SourceBranch string
	TargetBranch string

	// Author information
	Author *User

	// Status information
	IsDraft        bool
	IsMergeable    bool
	MergeCommitSHA string

	// Metadata
	CreatedAt string
	UpdatedAt string
	MergedAt  string
}

// PullRequestState represents the state of a pull request
type PullRequestState string

const (
	PullRequestStateOpen   PullRequestState = "open"
	PullRequestStateClosed PullRequestState = "closed"
	PullRequestStateMerged PullRequestState = "merged"
	PullRequestStateDraft  PullRequestState = "draft"
)

// CreatePullRequestInput contains the input for creating a pull request
type CreatePullRequestInput struct {
	RepoSlug     string
	Title        string
	Body         string
	SourceBranch string
	TargetBranch string
	IsDraft      bool
	Reviewers    []string
}

// UpdatePullRequestInput contains the input for updating a pull request
type UpdatePullRequestInput struct {
	RepoSlug     string
	Number       int
	Title        *string
	Body         *string
	State        *PullRequestState
	TargetBranch *string
	IsDraft      *bool
}

// ListOptions contains options for listing operations
type ListOptions struct {
	State      *PullRequestState
	Author     *string
	Limit      int
	Cursor     string
	OrderBy    string
	Direction  string
}

// ClientFactory creates provider-specific clients
type ClientFactory interface {
	// CreateGitHubClient creates a GitHub client
	CreateGitHubClient(ctx context.Context, provider *Provider) (ClientInterface, error)

	// CreateGitLabClient creates a GitLab client
	CreateGitLabClient(ctx context.Context, provider *Provider) (ClientInterface, error)
}

// DefaultClientFactory is the default implementation of ClientFactory
type DefaultClientFactory struct{}

// CreateGitHubClient creates a GitHub client
func (f *DefaultClientFactory) CreateGitHubClient(
	ctx context.Context,
	provider *Provider,
) (ClientInterface, error) {
	// TODO: Implement GitHub client wrapper
	return nil, ErrNotImplemented
}

// CreateGitLabClient creates a GitLab client
func (f *DefaultClientFactory) CreateGitLabClient(
	ctx context.Context,
	provider *Provider,
) (ClientInterface, error) {
	// TODO: Implement GitLab client wrapper
	return nil, ErrNotImplemented
}

// GetClient returns a client for the specified provider
func (f *DefaultClientFactory) GetClient(
	ctx context.Context,
	provider *Provider,
) (ClientInterface, error) {
	switch provider.Type {
	case GitHub:
		return f.CreateGitHubClient(ctx, provider)
	case GitLab:
		return f.CreateGitLabClient(ctx, provider)
	default:
		return nil, ErrUnsupportedProvider
	}
}

// Global client factory instance
var clientFactory ClientFactory = &DefaultClientFactory{}

// SetClientFactory sets the global client factory
func SetClientFactory(factory ClientFactory) {
	clientFactory = factory
}

// GetClientFactory returns the global client factory
func GetClientFactory() ClientFactory {
	return clientFactory
}

// GetClient is a convenience function to get a client for a provider
func GetClient(ctx context.Context, provider *Provider) (ClientInterface, error) {
	return clientFactory.GetClient(ctx, provider)
}
