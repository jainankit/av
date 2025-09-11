package providers

import (
	"context"
	"strings"
	"time"
)

// ProviderType represents the type of Git hosting provider
type ProviderType string

const (
	ProviderTypeGitHub ProviderType = "github"
	ProviderTypeGitLab ProviderType = "gitlab"
)

// PullRequestState represents the state of a pull/merge request
type PullRequestState string

const (
	PullRequestStateOpen   PullRequestState = "OPEN"
	PullRequestStateClosed PullRequestState = "CLOSED"
	PullRequestStateMerged PullRequestState = "MERGED"
	PullRequestStateDraft  PullRequestState = "DRAFT"
)

// PageInfo represents pagination information for API responses
type PageInfo struct {
	HasNextPage     bool   `json:"hasNextPage"`
	HasPreviousPage bool   `json:"hasPreviousPage"`
	StartCursor     string `json:"startCursor"`
	EndCursor       string `json:"endCursor"`
}

// Common data structures that work across providers

// PullRequest represents a pull request (GitHub) or merge request (GitLab)
type PullRequest struct {
	ID            string
	Number        int64
	Title         string
	Body          string
	HeadRefName   string
	BaseRefName   string
	IsDraft       bool
	State         PullRequestState
	Permalink     string
	MergeCommitOID string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Author        User
}

// HeadBranchName returns the head branch name without refs prefix
func (pr *PullRequest) HeadBranchName() string {
	return trimRefsPrefix(pr.HeadRefName)
}

// BaseBranchName returns the base branch name without refs prefix  
func (pr *PullRequest) BaseBranchName() string {
	return trimRefsPrefix(pr.BaseRefName)
}

// GetMergeCommit returns the merge commit SHA if available
func (pr *PullRequest) GetMergeCommit() string {
	return pr.MergeCommitOID
}

// User represents a user on the Git hosting provider
type User struct {
	ID       string
	Username string
	Name     string
	Email    string
}

// Repository represents a repository on the Git hosting provider
type Repository struct {
	ID          string
	Name        string
	FullName    string
	Owner       User
	HTMLURL     string
	CloneURL    string
	SSHCloneURL string
	DefaultBranch string
	IsPrivate   bool
}

// CreatePullRequestInput represents input for creating a pull/merge request
type CreatePullRequestInput struct {
	RepositoryID string
	Title        string
	Body         string
	HeadRefName  string
	BaseRefName  string
	Draft        bool
}

// UpdatePullRequestInput represents input for updating a pull/merge request
type UpdatePullRequestInput struct {
	PullRequestID string
	Title         *string
	Body          *string
	State         *PullRequestState
}

// GetPullRequestsInput represents input for querying pull/merge requests
type GetPullRequestsInput struct {
	Owner       string
	Repo        string
	HeadRefName string
	BaseRefName string
	States      []PullRequestState
	First       int32
	After       string
}

// GetPullRequestsPage represents a page of pull/merge requests
type GetPullRequestsPage struct {
	PageInfo     PageInfo
	TotalCount   int64
	PullRequests []PullRequest
}

// RequestReviewsInput represents input for requesting reviews
type RequestReviewsInput struct {
	PullRequestID string
	UserIDs       []string
	TeamIDs       []string
	Union         bool // Add reviewers instead of replacing them
}

// Provider interface abstracts operations across Git hosting providers
type Provider interface {
	// Type returns the provider type
	Type() ProviderType

	// MergeRequestProvider returns the merge request provider
	MergeRequestProvider() MergeRequestProvider

	// UserProvider returns the user provider
	UserProvider() UserProvider

	// RepositoryProvider returns the repository provider
	RepositoryProvider() RepositoryProvider
}

// MergeRequestProvider interface abstracts pull/merge request operations
type MergeRequestProvider interface {
	// GetByID retrieves a pull request by its ID
	GetByID(ctx context.Context, id string) (*PullRequest, error)

	// List retrieves pull requests based on input criteria
	List(ctx context.Context, input GetPullRequestsInput) (*GetPullRequestsPage, error)

	// Create creates a new pull request
	Create(ctx context.Context, input CreatePullRequestInput) (*PullRequest, error)

	// Update updates an existing pull request
	Update(ctx context.Context, input UpdatePullRequestInput) (*PullRequest, error)

	// RequestReviews requests reviews on a pull request
	RequestReviews(ctx context.Context, input RequestReviewsInput) (*PullRequest, error)

	// ConvertToDraft converts a pull request to draft
	ConvertToDraft(ctx context.Context, id string) (*PullRequest, error)

	// MarkReadyForReview marks a draft pull request as ready for review
	MarkReadyForReview(ctx context.Context, id string) (*PullRequest, error)

	// GetByRepo retrieves all pull requests for a repository with pagination
	GetByRepo(ctx context.Context, owner, repo string, opts GetRepoItemsOpts) (*GetPullRequestsPage, error)
}

// UserProvider interface abstracts user operations
type UserProvider interface {
	// GetByUsername retrieves a user by username
	GetByUsername(ctx context.Context, username string) (*User, error)

	// GetCurrentUser retrieves information about the authenticated user
	GetCurrentUser(ctx context.Context) (*User, error)
}

// RepositoryProvider interface abstracts repository operations
type RepositoryProvider interface {
	// GetBySlug retrieves a repository by owner/name slug
	GetBySlug(ctx context.Context, slug string) (*Repository, error)

	// GetByID retrieves a repository by its ID
	GetByID(ctx context.Context, id string) (*Repository, error)
}

// GetRepoItemsOpts represents options for getting repository items with pagination
type GetRepoItemsOpts struct {
	First  int32
	After  string
	States []PullRequestState
}

// ProviderError interface for provider-agnostic error handling
type ProviderError interface {
	error
	// Type returns the error type for classification
	Type() ProviderErrorType
	// Provider returns which provider the error originated from
	Provider() ProviderType
	// IsRetryable indicates if the error can be retried
	IsRetryable() bool
	// IsAuthenticationError indicates if this is an authentication error
	IsAuthenticationError() bool
	// IsNotFoundError indicates if this is a not found error
	IsNotFoundError() bool
	// IsRateLimitError indicates if this is a rate limiting error
	IsRateLimitError() bool
}

// ProviderErrorType represents different categories of provider errors
type ProviderErrorType string

const (
	ProviderErrorTypeAuthentication ProviderErrorType = "authentication"
	ProviderErrorTypeNotFound       ProviderErrorType = "not_found"
	ProviderErrorTypeRateLimit      ProviderErrorType = "rate_limit"
	ProviderErrorTypePermission     ProviderErrorType = "permission"
	ProviderErrorTypeValidation     ProviderErrorType = "validation"
	ProviderErrorTypeNetwork        ProviderErrorType = "network"
	ProviderErrorTypeInternal       ProviderErrorType = "internal"
	ProviderErrorTypeUnknown        ProviderErrorType = "unknown"
)

// trimRefsPrefix removes the "refs/heads/" prefix from branch names if present
func trimRefsPrefix(ref string) string {
	return strings.TrimPrefix(ref, "refs/heads/")
}