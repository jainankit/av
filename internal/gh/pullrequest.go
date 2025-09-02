package gh

import (
	"context"

	"github.com/aviator-co/av/internal/provider/github"
	"github.com/shurcooL/githubv4"
)

// PullRequest maintains the same structure as before for backward compatibility
type PullRequest = github.PullRequest

// PullRequestOpts represents options for getting a pull request by number
type PullRequestOpts = github.PullRequestOpts

// PullRequest retrieves a pull request by its ID
func (c *Client) PullRequest(ctx context.Context, id string) (*PullRequest, error) {
	return c.provider.PullRequest(ctx, id)
}

// GetPullRequestsInput represents the input for getting pull requests
type GetPullRequestsInput = github.GetPullRequestsInput

// GetPullRequestsPage represents a page of pull requests
type GetPullRequestsPage = github.GetPullRequestsPage

// GetPullRequests retrieves pull requests matching the given criteria
func (c *Client) GetPullRequests(
	ctx context.Context,
	input GetPullRequestsInput,
) (*GetPullRequestsPage, error) {
	return c.provider.GetPullRequestsGithub(ctx, input)
}

// CreatePullRequest creates a new pull request
func (c *Client) CreatePullRequest(
	ctx context.Context,
	input githubv4.CreatePullRequestInput,
) (*PullRequest, error) {
	return c.provider.CreatePullRequestGithub(ctx, input)
}

// UpdatePullRequest updates an existing pull request
func (c *Client) UpdatePullRequest(
	ctx context.Context,
	input githubv4.UpdatePullRequestInput,
) (*PullRequest, error) {
	return c.provider.UpdatePullRequestGithub(ctx, input)
}

// RequestReviews requests reviews from the given users on the given pull request
func (c *Client) RequestReviews(
	ctx context.Context,
	input githubv4.RequestReviewsInput,
) (*PullRequest, error) {
	return c.provider.RequestReviews(ctx, input)
}

// ConvertPullRequestToDraft converts a pull request to draft status
func (c *Client) ConvertPullRequestToDraft(ctx context.Context, id string) (*PullRequest, error) {
	return c.provider.ConvertPullRequestToDraft(ctx, id)
}

// MarkPullRequestReadyForReview marks a pull request as ready for review
func (c *Client) MarkPullRequestReadyForReview(
	ctx context.Context,
	id string,
) (*PullRequest, error) {
	return c.provider.MarkPullRequestReadyForReview(ctx, id)
}

// RepoPullRequestOpts represents options for repository pull request queries
type RepoPullRequestOpts = github.RepoPullRequestOpts

// RepoPullRequestsResponse represents the response for repository pull requests
type RepoPullRequestsResponse = github.RepoPullRequestsResponse

// RepoPullRequests retrieves all pull requests for a repository
func (c *Client) RepoPullRequests(
	ctx context.Context,
	opts RepoPullRequestOpts,
) (RepoPullRequestsResponse, error) {
	return c.provider.RepoPullRequests(ctx, opts)
}