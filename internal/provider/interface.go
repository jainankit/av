package provider

import (
	"context"
)

// Provider defines the interface for version control hosting providers (GitHub, GitLab, etc.)
type Provider interface {
	// Authentication and basic info
	GetViewer(ctx context.Context) (*User, error)
	
	// Repository operations
	GetRepository(ctx context.Context, slug string) (*Repository, error)
	
	// User operations  
	GetUser(ctx context.Context, login string) (*User, error)
	
	// MergeRequest/PullRequest operations
	GetMergeRequest(ctx context.Context, id string) (*MergeRequest, error)
	GetMergeRequests(ctx context.Context, input GetMergeRequestsInput) (*GetMergeRequestsPage, error)
	CreateMergeRequest(ctx context.Context, input CreateMergeRequestInput) (*MergeRequest, error)
	UpdateMergeRequest(ctx context.Context, input UpdateMergeRequestInput) (*MergeRequest, error)
	RequestReviews(ctx context.Context, input RequestReviewsInput) (*MergeRequest, error)
	ConvertToDraft(ctx context.Context, id string) (*MergeRequest, error)
	MarkReadyForReview(ctx context.Context, id string) (*MergeRequest, error)
	GetRepositoryMergeRequests(ctx context.Context, opts RepositoryMergeRequestsInput) (*RepositoryMergeRequestsPage, error)
	
	// Provider identification
	Name() ProviderName
}

// MergeRequestOperations defines operations specific to merge requests/pull requests
type MergeRequestOperations interface {
	HeadBranchName() string
	BaseBranchName() string
	GetMergeCommit() string
}

// RepositoryOperations defines operations specific to repositories
type RepositoryOperations interface {
	GetOwner() string
	GetName() string
	GetFullName() string
}