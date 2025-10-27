package github

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/aviator-co/av/internal/provider"
	"github.com/shurcooL/githubv4"
)

// Ensure Client implements provider.Client interface at compile time.
var _ provider.Client = (*Client)(nil)

// Name returns the human-readable name of the provider.
func (c *Client) Name() string {
	return "GitHub"
}

// Type returns the provider type identifier.
func (c *Client) Type() provider.ProviderType {
	return provider.ProviderTypeGitHub
}

// DetectFromURL determines if this provider can handle the given repository URL.
// Returns true for github.com and GitHub Enterprise URLs.
func (c *Client) DetectFromURL(u *url.URL) bool {
	host := strings.ToLower(u.Host)
	// Handle github.com and common variations
	if host == "github.com" || host == "www.github.com" {
		return true
	}
	// GitHub Enterprise URLs typically have paths starting with /api/v3 or /api/graphql
	// We can't definitively detect GHE from URL alone without making a request,
	// but we can exclude known non-GitHub hosts (gitlab.com, bitbucket.org, etc.)
	if strings.Contains(host, "gitlab") || strings.Contains(host, "bitbucket") {
		return false
	}
	// For self-hosted instances, we return true as a fallback
	// The actual detection will be refined in later steps with configuration
	return true
}

// CreatePullRequest creates a new pull request (provider interface implementation).
func (c *Client) CreatePullRequest(
	ctx context.Context, input provider.CreatePullRequestInput,
) (*provider.PullRequest, error) {
	ghInput := githubv4.CreatePullRequestInput{
		RepositoryID: githubv4.ID(input.RepositoryID),
		Title:        githubv4.String(input.Title),
		Body:         Ptr(githubv4.String(input.Body)),
		HeadRefName:  githubv4.String(input.HeadRefName),
		BaseRefName:  githubv4.String(input.BaseRefName),
		Draft:        Ptr(githubv4.Boolean(input.Draft)),
	}

	ghPR, err := c.createPullRequestInternal(ctx, ghInput)
	if err != nil {
		return nil, err
	}

	return convertPullRequest(ghPR), nil
}

// UpdatePullRequest updates an existing pull request (provider interface implementation).
func (c *Client) UpdatePullRequest(
	ctx context.Context, input provider.UpdatePullRequestInput,
) (*provider.PullRequest, error) {
	ghInput := githubv4.UpdatePullRequestInput{
		PullRequestID: githubv4.ID(input.ID),
	}

	if input.Title != nil {
		ghInput.Title = Ptr(githubv4.String(*input.Title))
	}
	if input.Body != nil {
		ghInput.Body = Ptr(githubv4.String(*input.Body))
	}
	if input.BaseRefName != nil {
		ghInput.BaseRefName = Ptr(githubv4.String(*input.BaseRefName))
	}

	ghPR, err := c.updatePullRequestInternal(ctx, ghInput)
	if err != nil {
		return nil, err
	}

	return convertPullRequest(ghPR), nil
}

// GetPullRequest retrieves a pull request by its ID (provider interface implementation).
func (c *Client) GetPullRequest(ctx context.Context, id string) (*provider.PullRequest, error) {
	ghPR, err := c.PullRequest(ctx, id)
	if err != nil {
		return nil, err
	}

	return convertPullRequest(ghPR), nil
}

// ListPullRequests lists pull requests based on the given criteria (provider interface implementation).
func (c *Client) ListPullRequests(
	ctx context.Context, input provider.ListPullRequestsInput,
) (*provider.PullRequestsPage, error) {
	ghInput := GetPullRequestsInput{
		Owner:       input.Owner,
		Repo:        input.Repo,
		HeadRefName: input.HeadRefName,
		BaseRefName: input.BaseRefName,
		First:       input.First,
		After:       input.After,
	}

	// Convert provider states to GitHub states
	if len(input.States) > 0 {
		ghInput.States = make([]githubv4.PullRequestState, len(input.States))
		for i, state := range input.States {
			ghInput.States[i] = convertToGitHubState(state)
		}
	}

	ghPage, err := c.GetPullRequests(ctx, ghInput)
	if err != nil {
		return nil, err
	}

	// Convert results
	prs := make([]provider.PullRequest, len(ghPage.PullRequests))
	for i, ghPR := range ghPage.PullRequests {
		prs[i] = *convertPullRequest(&ghPR)
	}

	return &provider.PullRequestsPage{
		PullRequests: prs,
		PageInfo: provider.PageInfo{
			EndCursor:       ghPage.PageInfo.EndCursor,
			HasNextPage:     ghPage.PageInfo.HasNextPage,
			HasPreviousPage: ghPage.PageInfo.HasPreviousPage,
			StartCursor:     ghPage.PageInfo.StartCursor,
		},
	}, nil
}

// ConvertPullRequestToDraft converts a pull request to draft status (provider interface implementation).
func (c *Client) ConvertPullRequestToDraft(ctx context.Context, id string) (*provider.PullRequest, error) {
	ghPR, err := c.convertPullRequestToDraftInternal(ctx, id)
	if err != nil {
		return nil, err
	}

	return convertPullRequest(ghPR), nil
}

// MarkPullRequestReadyForReview marks a pull request as ready for review (provider interface implementation).
func (c *Client) MarkPullRequestReadyForReview(ctx context.Context, id string) (*provider.PullRequest, error) {
	ghPR, err := c.markPullRequestReadyForReviewInternal(ctx, id)
	if err != nil {
		return nil, err
	}

	return convertPullRequest(ghPR), nil
}

// RequestReviews requests reviews from users/teams on a pull request (provider interface implementation).
func (c *Client) RequestReviews(
	ctx context.Context, input provider.RequestReviewsInput,
) (*provider.PullRequest, error) {
	ghInput := githubv4.RequestReviewsInput{
		PullRequestID: githubv4.ID(input.PullRequestID),
		Union:         Ptr(githubv4.Boolean(input.Union)),
	}

	if len(input.UserIDs) > 0 {
		userIDs := make([]githubv4.ID, len(input.UserIDs))
		for i, id := range input.UserIDs {
			userIDs[i] = githubv4.ID(id)
		}
		ghInput.UserIDs = &userIDs
	}

	if len(input.TeamIDs) > 0 {
		teamIDs := make([]githubv4.ID, len(input.TeamIDs))
		for i, id := range input.TeamIDs {
			teamIDs[i] = githubv4.ID(id)
		}
		ghInput.TeamIDs = &teamIDs
	}

	ghPR, err := c.requestReviewsInternal(ctx, ghInput)
	if err != nil {
		return nil, err
	}

	return convertPullRequest(ghPR), nil
}

// GetRepositoryBySlug retrieves repository information by owner/name slug (provider interface implementation).
func (c *Client) GetRepositoryBySlug(ctx context.Context, slug string) (*provider.Repository, error) {
	ghRepo, err := c.getRepositoryBySlugInternal(ctx, slug)
	if err != nil {
		return nil, err
	}

	return &provider.Repository{
		ID:       ghRepo.ID,
		Owner:    ghRepo.Owner.Login,
		Name:     ghRepo.Name,
		Provider: provider.ProviderTypeGitHub,
	}, nil
}

// GetUser retrieves user information by login (provider interface implementation).
func (c *Client) GetUser(ctx context.Context, login string) (*provider.User, error) {
	ghUser, err := c.User(ctx, login)
	if err != nil {
		return nil, err
	}

	return &provider.User{
		ID:    string(ghUser.ID),
		Login: ghUser.Login,
		Name:  "", // GitHub user query doesn't include name field by default
	}, nil
}

// GetTeam retrieves team information by organization and team slug (provider interface implementation).
func (c *Client) GetTeam(ctx context.Context, organizationLogin, teamSlug string) (*provider.Team, error) {
	ghTeam, err := c.OrganizationTeam(ctx, organizationLogin, teamSlug)
	if err != nil {
		return nil, err
	}

	return &provider.Team{
		ID:   string(ghTeam.ID),
		Name: ghTeam.Name,
		Slug: ghTeam.Slug,
	}, nil
}

// GetViewer retrieves information about the currently authenticated user (provider interface implementation).
func (c *Client) GetViewer(ctx context.Context) (*provider.User, error) {
	ghViewer, err := c.Viewer(ctx)
	if err != nil {
		return nil, err
	}

	return &provider.User{
		ID:    "", // Viewer query doesn't include ID by default
		Login: ghViewer.Login,
		Name:  ghViewer.Name,
	}, nil
}

// Internal wrapper functions to avoid method name conflicts

func (c *Client) createPullRequestInternal(
	ctx context.Context, input githubv4.CreatePullRequestInput,
) (*PullRequest, error) {
	var mutation struct {
		CreatePullRequest struct {
			PullRequest PullRequest
		} `graphql:"createPullRequest(input: $input)"`
	}
	if err := c.mutate(ctx, &mutation, input, nil); err != nil {
		return nil, err
	}
	return &mutation.CreatePullRequest.PullRequest, nil
}

func (c *Client) updatePullRequestInternal(
	ctx context.Context, input githubv4.UpdatePullRequestInput,
) (*PullRequest, error) {
	var mutation struct {
		UpdatePullRequest struct {
			PullRequest PullRequest
		} `graphql:"updatePullRequest(input: $input)"`
	}
	if err := c.mutate(ctx, &mutation, input, nil); err != nil {
		return nil, err
	}
	return &mutation.UpdatePullRequest.PullRequest, nil
}

func (c *Client) convertPullRequestToDraftInternal(ctx context.Context, id string) (*PullRequest, error) {
	var mutation struct {
		ConvertPullRequestToDraft struct {
			PullRequest PullRequest
		} `graphql:"convertPullRequestToDraft(input: $input)"`
	}
	if err := c.mutate(ctx, &mutation, githubv4.ConvertPullRequestToDraftInput{PullRequestID: id}, nil); err != nil {
		return nil, err
	}
	return &mutation.ConvertPullRequestToDraft.PullRequest, nil
}

func (c *Client) markPullRequestReadyForReviewInternal(ctx context.Context, id string) (*PullRequest, error) {
	var mutation struct {
		MarkPullRequestReadyForReview struct {
			PullRequest PullRequest
		} `graphql:"markPullRequestReadyForReview(input: $input)"`
	}
	if err := c.mutate(ctx, &mutation, githubv4.MarkPullRequestReadyForReviewInput{PullRequestID: id}, nil); err != nil {
		return nil, err
	}
	return &mutation.MarkPullRequestReadyForReview.PullRequest, nil
}

func (c *Client) requestReviewsInternal(ctx context.Context, input githubv4.RequestReviewsInput) (*PullRequest, error) {
	if input.Union == nil {
		input.Union = Ptr[githubv4.Boolean](true)
	}
	var mutation struct {
		RequestReviews struct {
			PullRequest PullRequest
		} `graphql:"requestReviews(input: $input)"`
	}
	if err := c.mutate(ctx, &mutation, input, nil); err != nil {
		return nil, err
	}
	return &mutation.RequestReviews.PullRequest, nil
}

func (c *Client) getRepositoryBySlugInternal(ctx context.Context, slug string) (*Repository, error) {
	owner, name, ok := strings.Cut(slug, "/")
	if !ok {
		return nil, errors.New("invalid repository slug format")
	}

	var query struct {
		Repository Repository `graphql:"repository(owner: $owner, name: $name)"`
	}
	err := c.query(ctx, &query, map[string]interface{}{
		"owner": githubv4.String(owner),
		"name":  githubv4.String(name),
	})
	if err != nil {
		return nil, err
	}

	return &query.Repository, nil
}

// Conversion helper functions

// convertPullRequest converts a GitHub PullRequest to a provider PullRequest.
func convertPullRequest(ghPR *PullRequest) *provider.PullRequest {
	return &provider.PullRequest{
		ID:             ghPR.ID,
		Number:         ghPR.Number,
		Title:          ghPR.Title,
		Body:           ghPR.Body,
		State:          convertFromGitHubState(ghPR.State),
		HeadRefName:    ghPR.HeadRefName,
		BaseRefName:    ghPR.BaseRefName,
		Permalink:      ghPR.Permalink,
		IsDraft:        ghPR.IsDraft,
		MergeCommitOID: ghPR.GetMergeCommit(),
	}
}

// convertFromGitHubState converts a githubv4.PullRequestState to provider.PullRequestState.
func convertFromGitHubState(state githubv4.PullRequestState) provider.PullRequestState {
	switch state {
	case githubv4.PullRequestStateOpen:
		return provider.PullRequestStateOpen
	case githubv4.PullRequestStateClosed:
		return provider.PullRequestStateClosed
	case githubv4.PullRequestStateMerged:
		return provider.PullRequestStateMerged
	default:
		return provider.PullRequestStateOpen
	}
}

// convertToGitHubState converts a provider.PullRequestState to githubv4.PullRequestState.
func convertToGitHubState(state provider.PullRequestState) githubv4.PullRequestState {
	switch state {
	case provider.PullRequestStateOpen:
		return githubv4.PullRequestStateOpen
	case provider.PullRequestStateClosed:
		return githubv4.PullRequestStateClosed
	case provider.PullRequestStateMerged:
		return githubv4.PullRequestStateMerged
	default:
		return githubv4.PullRequestStateOpen
	}
}
