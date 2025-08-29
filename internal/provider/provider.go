package provider

import "context"

// ProviderType represents the type of Git hosting provider.
type ProviderType string

const (
	// GitHub represents GitHub as the hosting provider.
	GitHub ProviderType = "github"
	// GitLab represents GitLab as the hosting provider.
	GitLab ProviderType = "gitlab"
)

// Provider defines the interface for interacting with Git hosting providers.
// This abstraction allows the application to work with both GitHub PRs and GitLab MRs
// through a unified interface.
type Provider interface {
	// CreateMergeRequest creates a new merge request (PR for GitHub, MR for GitLab).
	CreateMergeRequest(ctx context.Context, input CreateMRInput) (*MergeRequest, error)

	// UpdateMergeRequest updates an existing merge request.
	UpdateMergeRequest(ctx context.Context, input UpdateMRInput) (*MergeRequest, error)

	// GetMergeRequests retrieves merge requests based on the provided criteria.
	GetMergeRequests(ctx context.Context, input GetMRsInput) (*MergeRequestsPage, error)

	// GetRepository retrieves repository information.
	GetRepository(ctx context.Context, owner, repo string) (*Repository, error)

	// ConvertToDraft converts a merge request to draft status.
	ConvertToDraft(ctx context.Context, id string) (*MergeRequest, error)

	// MarkReadyForReview marks a merge request as ready for review (removes draft status).
	MarkReadyForReview(ctx context.Context, id string) (*MergeRequest, error)

	// RequestReviews requests reviews from specified users on the merge request.
	RequestReviews(ctx context.Context, id string, reviewers []string) (*MergeRequest, error)
}