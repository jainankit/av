package provider

import "context"

// Provider is the interface that abstracts different git hosting providers
// (e.g., GitHub, GitLab). It defines all the operations needed by av CLI
// to interact with pull requests, repositories, users, and teams.
type Provider interface {
	// Name returns the name of the provider (e.g., "GitHub", "GitLab").
	Name() string

	// GetRepository returns information about the repository associated with
	// this provider instance.
	GetRepository(ctx context.Context) (*Repository, error)

	// GetPR returns information about a specific pull request by its ID.
	GetPR(ctx context.Context, id string) (*PullRequest, error)

	// GetPRs returns a paginated list of pull requests matching the given criteria.
	GetPRs(ctx context.Context, input GetPRsInput) (*GetPRsPage, error)

	// CreatePR creates a new pull request.
	CreatePR(ctx context.Context, input CreatePRInput) (*PullRequest, error)

	// UpdatePR updates an existing pull request.
	UpdatePR(ctx context.Context, input UpdatePRInput) (*PullRequest, error)

	// RequestReviews requests reviews from users and/or teams on a pull request.
	// The userIDs and teamIDs are provider-specific identifiers.
	RequestReviews(ctx context.Context, input RequestReviewsInput) (*PullRequest, error)

	// ConvertToDraft converts a pull request to draft status.
	ConvertToDraft(ctx context.Context, id string) (*PullRequest, error)

	// MarkReadyForReview marks a draft pull request as ready for review.
	MarkReadyForReview(ctx context.Context, id string) (*PullRequest, error)

	// GetViewer returns information about the currently authenticated user.
	GetViewer(ctx context.Context) (*Viewer, error)

	// GetUser returns information about a user by their login/username.
	GetUser(ctx context.Context, login string) (*User, error)

	// GetTeam returns information about a team within an organization.
	// Note: Not all providers support teams (e.g., GitLab does not).
	GetTeam(ctx context.Context, orgLogin string, teamSlug string) (*Team, error)
}

// RequestReviewsInput contains the parameters for requesting reviews on a pull request.
type RequestReviewsInput struct {
	// PullRequestID is the provider-specific identifier for the pull request.
	PullRequestID string
	// UserIDs is a list of user IDs to request reviews from.
	UserIDs []string
	// TeamIDs is a list of team IDs to request reviews from.
	// Note: Not all providers support team reviewers.
	TeamIDs []string
}
