package provider

import (
	"context"
)

// ProviderType represents the type of version control provider
type ProviderType string

const (
	ProviderTypeGitHub ProviderType = "github"
	ProviderTypeGitLab ProviderType = "gitlab"
)

// MRState represents the unified state of a merge request/pull request
type MRState string

const (
	MRStateOpen   MRState = "open"
	MRStateClosed MRState = "closed"
	MRStateMerged MRState = "merged"
)

// Provider defines the interface for version control providers that support merge requests/pull requests
type Provider interface {
	// CreateMergeRequest creates a new merge request/pull request
	CreateMergeRequest(ctx context.Context, input CreateMRInput) (*MergeRequest, error)
	
	// UpdateMergeRequest updates an existing merge request/pull request
	UpdateMergeRequest(ctx context.Context, input UpdateMRInput) (*MergeRequest, error)
	
	// GetMergeRequests retrieves merge requests/pull requests with filtering
	GetMergeRequests(ctx context.Context, input GetMRsInput) (*GetMRsPage, error)
	
	// GetRepository retrieves repository information
	GetRepository(ctx context.Context, owner, repo string) (*Repository, error)
	
	// ConvertToDraft converts a merge request/pull request to draft status
	ConvertToDraft(ctx context.Context, projectID string, mrID string) (*MergeRequest, error)
	
	// MarkReadyForReview marks a merge request/pull request as ready for review
	MarkReadyForReview(ctx context.Context, projectID string, mrID string) (*MergeRequest, error)
	
	// RequestReviews requests reviews from specified users
	RequestReviews(ctx context.Context, projectID string, mrID string, reviewers []string) (*MergeRequest, error)
}