package providers

import (
	"context"
)

// Provider defines the interface that all provider implementations must satisfy.
// This abstraction allows supporting multiple Git hosting providers (GitHub, GitLab, etc.)
// with a consistent API.
type Provider interface {
	// Pull Request operations
	PullRequest(ctx context.Context, id string) (*PullRequest, error)
	GetPullRequests(ctx context.Context, input GetPullRequestsInput) (*GetPullRequestsPage, error)
	CreatePullRequest(ctx context.Context, input CreatePullRequestInput) (*PullRequest, error)
	UpdatePullRequest(ctx context.Context, input UpdatePullRequestInput) (*PullRequest, error)
	RequestReviews(ctx context.Context, input RequestReviewsInput) (*PullRequest, error)
	ConvertPullRequestToDraft(ctx context.Context, id string) (*PullRequest, error)
	MarkPullRequestReadyForReview(ctx context.Context, id string) (*PullRequest, error)
	RepoPullRequests(ctx context.Context, input RepoPullRequestsInput) (*RepoPullRequestsPage, error)

	// Repository operations
	GetRepositoryBySlug(ctx context.Context, slug string) (*Repository, error)

	// User operations
	User(ctx context.Context, login string) (*User, error)
	Viewer(ctx context.Context) (*Viewer, error)

	// Team operations (organization/group management)
	OrganizationTeam(ctx context.Context, organizationLogin string, teamSlug string) (*Team, error)
}

// GetPullRequestsInput defines the input parameters for listing pull requests
type GetPullRequestsInput struct {
	// Required
	Owner string
	Repo  string
	// Optional
	HeadRefName string
	BaseRefName string
	States      []PullRequestState
	First       int32
	After       string
}

// CreatePullRequestInput defines the input parameters for creating a pull request
type CreatePullRequestInput struct {
	RepositoryID string
	BaseRefName  string
	HeadRefName  string
	Title        string
	Body         *string
	Draft        *bool
}

// UpdatePullRequestInput defines the input parameters for updating a pull request
type UpdatePullRequestInput struct {
	PullRequestID string
	BaseRefName   *string
	Title         *string
	Body          *string
}

// RequestReviewsInput defines the input parameters for requesting reviews
type RequestReviewsInput struct {
	PullRequestID string
	UserIDs       []string
	TeamIDs       []string
	Union         bool // true to replace existing reviewers, false to add to them
}

// RepoPullRequestsInput defines the input parameters for listing repository pull requests
type RepoPullRequestsInput struct {
	Owner       string
	Repo        string
	HeadRefName *string
	BaseRefName *string
	States      []PullRequestState
	First       int32
	After       string
}