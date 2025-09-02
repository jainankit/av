package github

import (
	"context"

	"emperror.dev/errors"
	"github.com/shurcooL/githubv4"
)

// PullRequest represents a GitHub pull request with all the GitHub-specific details
// This type is used for backward compatibility with existing code
type PullRequest = githubPullRequest

// PullRequestOpts represents options for getting a pull request by number
type PullRequestOpts struct {
	Owner  string
	Repo   string
	Number int64
}

// PullRequest retrieves a pull request by its ID
func (c *Client) PullRequest(ctx context.Context, id string) (*PullRequest, error) {
	var query struct {
		Node struct {
			PullRequest PullRequest `graphql:"... on PullRequest"`
		} `graphql:"node(id: $id)"`
	}
	if err := c.query(ctx, &query, map[string]interface{}{
		"id": githubv4.ID(id),
	}); err != nil {
		return nil, errors.Wrap(err, "failed to query pull request")
	}
	if query.Node.PullRequest.ID == "" {
		return nil, errors.Errorf("pull request %q not found", id)
	}
	return &query.Node.PullRequest, nil
}

// GetPullRequestsInput represents the input for getting pull requests (GitHub-specific version)
type GetPullRequestsInput struct {
	// REQUIRED
	Owner string
	Repo  string
	// OPTIONAL
	HeadRefName string
	BaseRefName string
	States      []githubv4.PullRequestState
	First       int32
	After       string
}

// GetPullRequestsPage represents a page of pull requests (GitHub-specific version)
type GetPullRequestsPage struct {
	PageInfo     PageInfo
	PullRequests []PullRequest
}

// PageInfo represents pagination information for GitHub API
type PageInfo struct {
	EndCursor       string
	HasNextPage     bool
	HasPreviousPage bool
	StartCursor     string
}

// GetPullRequestsGithub retrieves pull requests using GitHub-specific types for backward compatibility
func (c *Client) GetPullRequestsGithub(
	ctx context.Context,
	input GetPullRequestsInput,
) (*GetPullRequestsPage, error) {
	if input.First == 0 {
		input.First = 50
	}
	var query struct {
		Repository struct {
			PullRequests struct {
				Nodes    []PullRequest
				PageInfo PageInfo
			} `graphql:"pullRequests(states: $states, headRefName: $headRefName, baseRefName: $baseRefName, first: $first, after: $after)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}
	if err := c.query(ctx, &query, map[string]interface{}{
		"owner":       githubv4.String(input.Owner),
		"repo":        githubv4.String(input.Repo),
		"headRefName": nullable(githubv4.String(input.HeadRefName)),
		"baseRefName": nullable(githubv4.String(input.BaseRefName)),
		"states":      &input.States,
		"first":       githubv4.Int(input.First),
		"after":       nullable(githubv4.String(input.After)),
	}); err != nil {
		return nil, errors.Wrap(err, "failed to query pull requests")
	}
	return &GetPullRequestsPage{
		PageInfo:     query.Repository.PullRequests.PageInfo,
		PullRequests: query.Repository.PullRequests.Nodes,
	}, nil
}

// CreatePullRequestGithub creates a pull request using GitHub-specific input types
func (c *Client) CreatePullRequestGithub(
	ctx context.Context,
	input githubv4.CreatePullRequestInput,
) (*PullRequest, error) {
	var mutation struct {
		CreatePullRequest struct {
			PullRequest PullRequest
		} `graphql:"createPullRequest(input: $input)"`
	}
	if err := c.mutate(ctx, &mutation, input, nil); err != nil {
		return nil, errors.Wrap(err, "failed to create pull request: github error")
	}
	return &mutation.CreatePullRequest.PullRequest, nil
}

// UpdatePullRequestGithub updates a pull request using GitHub-specific input types
func (c *Client) UpdatePullRequestGithub(
	ctx context.Context,
	input githubv4.UpdatePullRequestInput,
) (*PullRequest, error) {
	var mutation struct {
		UpdatePullRequest struct {
			PullRequest PullRequest
		} `graphql:"updatePullRequest(input: $input)"`
	}
	if err := c.mutate(ctx, &mutation, input, nil); err != nil {
		return nil, errors.Wrap(err, "failed to update pull request: github error")
	}
	return &mutation.UpdatePullRequest.PullRequest, nil
}

// RequestReviews requests reviews from the given users on the given pull request
func (c *Client) RequestReviews(
	ctx context.Context,
	input githubv4.RequestReviewsInput,
) (*PullRequest, error) {
	if input.Union == nil {
		// Add reviewers instead of replacing them by default.
		input.Union = Ptr[githubv4.Boolean](true)
	}
	var mutation struct {
		RequestReviews struct {
			PullRequest PullRequest
		} `graphql:"requestReviews(input: $input)"`
	}
	if err := c.mutate(ctx, &mutation, input, nil); err != nil {
		return nil, errors.Wrap(err, "failed to request pull request reviews")
	}
	return &mutation.RequestReviews.PullRequest, nil
}

// ConvertPullRequestToDraft converts a pull request to draft status
func (c *Client) ConvertPullRequestToDraft(ctx context.Context, id string) (*PullRequest, error) {
	var mutation struct {
		ConvertPullRequestToDraft struct {
			PullRequest PullRequest
		} `graphql:"convertPullRequestToDraft(input: $input)"`
	}
	if err := c.mutate(ctx, &mutation, githubv4.ConvertPullRequestToDraftInput{PullRequestID: id}, nil); err != nil {
		return nil, errors.Wrap(err, "failed to convert pull request to draft: github error")
	}
	return &mutation.ConvertPullRequestToDraft.PullRequest, nil
}

// MarkPullRequestReadyForReview marks a pull request as ready for review
func (c *Client) MarkPullRequestReadyForReview(
	ctx context.Context,
	id string,
) (*PullRequest, error) {
	var mutation struct {
		MarkPullRequestReadyForReview struct {
			PullRequest PullRequest
		} `graphql:"markPullRequestReadyForReview(input: $input)"`
	}
	if err := c.mutate(ctx, &mutation, githubv4.MarkPullRequestReadyForReviewInput{PullRequestID: id}, nil); err != nil {
		return nil, errors.Wrap(err, "failed to mark pull request ready for review: github error")
	}
	return &mutation.MarkPullRequestReadyForReview.PullRequest, nil
}

// RepoPullRequestOpts represents options for repository pull request queries
type RepoPullRequestOpts struct {
	Owner  string
	Repo   string
	First  int32
	After  string
	States []githubv4.PullRequestState
}

// RepoPullRequestsResponse represents the response for repository pull requests
type RepoPullRequestsResponse struct {
	PageInfo
	TotalCount   int64
	PullRequests []PullRequest
}

// RepoPullRequests retrieves all pull requests for a repository
func (c *Client) RepoPullRequests(
	ctx context.Context,
	opts RepoPullRequestOpts,
) (RepoPullRequestsResponse, error) {
	var query struct {
		Repository struct {
			PullRequests struct {
				TotalCount int64
				PageInfo   PageInfo
				Nodes      []PullRequest
			} `graphql:"pullRequests(states: $states, first: $first, after: $after)"`
		} `graphql:"repository(owner:$owner, name:$repo)"`
	}

	if opts.First == 0 {
		opts.First = 100
	}
	vars := map[string]any{
		"owner":  githubv4.String(opts.Owner),
		"repo":   githubv4.String(opts.Repo),
		"first":  githubv4.Int(opts.First),
		"after":  nullable(githubv4.String(opts.After)),
		"states": opts.States,
	}
	if opts.After != "" {
		vars["after"] = githubv4.String(opts.After)
	}
	if len(opts.States) > 0 {
		vars["states"] = opts.States
	}
	if err := c.query(ctx, &query, vars); err != nil {
		return RepoPullRequestsResponse{}, errors.Wrap(err, "failed to query pull requests")
	}
	return RepoPullRequestsResponse{
		PageInfo:     query.Repository.PullRequests.PageInfo,
		TotalCount:   query.Repository.PullRequests.TotalCount,
		PullRequests: query.Repository.PullRequests.Nodes,
	}, nil
}

// Ptr returns a pointer to the argument.
//
// It's a convenience function to make working with the API easier: since Go
// disallows pointers-to-literals, and optional input fields are expressed as
// pointers, this function can be used to easily set optional fields to non-nil
// primitives.
//
// For example, `githubv4.CreatePullRequestInput{Draft: Ptr(true)}`.
func Ptr[T any](v T) *T {
	return &v
}

// Helper methods for PullRequest type compatibility
func (p *PullRequest) HeadBranchName() string {
	// Note: GH sometimes includes the "refs/heads/" prefix and sometimes it doesn't.
	// I think(?) it might just return exactly what is given to the API during
	// creation.
	return trimRefsPrefix(p.HeadRefName)
}

func (p *PullRequest) BaseBranchName() string {
	// See comment in HeadBranchName above.
	return trimRefsPrefix(p.BaseRefName)
}

func (p *PullRequest) GetMergeCommit() string {
	return getMergeCommitFromGithubPR(p)
}

// trimRefsPrefix removes the "refs/heads/" prefix from branch names if present
func trimRefsPrefix(refName string) string {
	const prefix = "refs/heads/"
	if len(refName) > len(prefix) && refName[:len(prefix)] == prefix {
		return refName[len(prefix):]
	}
	return refName
}