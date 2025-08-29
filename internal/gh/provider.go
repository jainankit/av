package gh

import (
	"context"
	"strconv"
	"strings"

	"emperror.dev/errors"
	"github.com/aviator-co/av/internal/provider"
	"github.com/shurcooL/githubv4"
)

// GitHubProvider wraps the GitHub client to implement the Provider interface
type GitHubProvider struct {
	client *Client
}

// NewProvider creates a new GitHub provider instance
func NewProvider(client *Client) provider.Provider {
	return &GitHubProvider{
		client: client,
	}
}

// CreateMergeRequest creates a new pull request on GitHub
func (p *GitHubProvider) CreateMergeRequest(ctx context.Context, input provider.CreateMRInput) (*provider.MergeRequest, error) {
	// Parse projectID to extract owner/repo
	owner, repo, err := parseProjectID(input.ProjectID)
	if err != nil {
		return nil, err
	}

	// Get repository to get the proper repository ID
	ghRepo, err := p.client.GetRepositoryBySlug(ctx, input.ProjectID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get repository")
	}

	// Create GitHub pull request input
	ghInput := githubv4.CreatePullRequestInput{
		RepositoryID: githubv4.ID(ghRepo.ID),
		Title:        githubv4.String(input.Title),
		Body:         nullable(githubv4.String(input.Description)),
		HeadRefName:  githubv4.String(input.SourceBranch),
		BaseRefName:  githubv4.String(input.TargetBranch),
		Draft:        nullable(githubv4.Boolean(input.Draft)),
	}

	ghPR, err := p.client.CreatePullRequest(ctx, ghInput)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create GitHub pull request")
	}

	return p.convertGitHubPRToMR(ghPR, owner, repo), nil
}

// UpdateMergeRequest updates an existing pull request on GitHub
func (p *GitHubProvider) UpdateMergeRequest(ctx context.Context, input provider.UpdateMRInput) (*provider.MergeRequest, error) {
	// Parse projectID to extract owner/repo
	owner, repo, err := parseProjectID(input.ProjectID)
	if err != nil {
		return nil, err
	}

	// Create GitHub update input
	ghInput := githubv4.UpdatePullRequestInput{
		PullRequestID: githubv4.ID(input.MRID),
	}

	if input.Title != nil {
		ghInput.Title = nullable(githubv4.String(*input.Title))
	}
	if input.Description != nil {
		ghInput.Body = nullable(githubv4.String(*input.Description))
	}

	// Handle state changes
	if input.StateEvent != nil {
		switch *input.StateEvent {
		case "close":
			ghInput.State = nullable(githubv4.PullRequestState(githubv4.PullRequestStateClosed))
		case "reopen":
			ghInput.State = nullable(githubv4.PullRequestState(githubv4.PullRequestStateOpen))
		}
	}

	ghPR, err := p.client.UpdatePullRequest(ctx, ghInput)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update GitHub pull request")
	}

	// Handle draft status separately if needed
	if input.Draft != nil {
		if *input.Draft && !ghPR.IsDraft {
			ghPR, err = p.client.ConvertPullRequestToDraft(ctx, ghPR.ID)
			if err != nil {
				return nil, errors.Wrap(err, "failed to convert pull request to draft")
			}
		} else if !*input.Draft && ghPR.IsDraft {
			ghPR, err = p.client.MarkPullRequestReadyForReview(ctx, ghPR.ID)
			if err != nil {
				return nil, errors.Wrap(err, "failed to mark pull request ready for review")
			}
		}
	}

	return p.convertGitHubPRToMR(ghPR, owner, repo), nil
}

// GetMergeRequests retrieves pull requests from GitHub with filtering
func (p *GitHubProvider) GetMergeRequests(ctx context.Context, input provider.GetMRsInput) (*provider.GetMRsPage, error) {
	// Parse projectID to extract owner/repo
	owner, repo, err := parseProjectID(input.ProjectID)
	if err != nil {
		return nil, err
	}

	// Convert unified states to GitHub states
	var ghStates []githubv4.PullRequestState
	for _, state := range input.States {
		ghState := convertMRStateToGitHub(state)
		ghStates = append(ghStates, ghState)
	}

	// Create GitHub input
	ghInput := GetPullRequestsInput{
		Owner:       owner,
		Repo:        repo,
		HeadRefName: input.SourceBranch,
		BaseRefName: input.TargetBranch,
		States:      ghStates,
		First:       int32(input.PerPage),
	}

	// Calculate after cursor based on page
	if input.Page > 1 {
		// For simplicity, we'll use a basic page offset approach
		// In a real implementation, you'd want to properly handle cursors
		ghInput.After = ""
	}

	ghPage, err := p.client.GetPullRequests(ctx, ghInput)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get GitHub pull requests")
	}

	// Convert GitHub PRs to unified MRs
	mrs := make([]provider.MergeRequest, len(ghPage.PullRequests))
	for i, ghPR := range ghPage.PullRequests {
		mrs[i] = *p.convertGitHubPRToMR(&ghPR, owner, repo)
	}

	return &provider.GetMRsPage{
		MergeRequests: mrs,
		NextPage:      input.Page + 1,
		HasNextPage:   ghPage.HasNextPage,
		TotalPages:    0, // GitHub doesn't provide total pages in this format
		TotalCount:    0, // GitHub doesn't provide total count in this format
	}, nil
}

// GetRepository retrieves repository information from GitHub
func (p *GitHubProvider) GetRepository(ctx context.Context, owner, repo string) (*provider.Repository, error) {
	projectID := owner + "/" + repo
	ghRepo, err := p.client.GetRepositoryBySlug(ctx, projectID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get GitHub repository")
	}

	return &provider.Repository{
		ID:       ghRepo.ID,
		Name:     ghRepo.Name,
		FullName: ghRepo.Owner.Login + "/" + ghRepo.Name,
		Owner:    ghRepo.Owner.Login,
		WebURL:   "https://github.com/" + ghRepo.Owner.Login + "/" + ghRepo.Name,
	}, nil
}

// ConvertToDraft converts a pull request to draft status
func (p *GitHubProvider) ConvertToDraft(ctx context.Context, projectID string, mrID string) (*provider.MergeRequest, error) {
	// Parse projectID to extract owner/repo
	owner, repo, err := parseProjectID(projectID)
	if err != nil {
		return nil, err
	}

	ghPR, err := p.client.ConvertPullRequestToDraft(ctx, mrID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert GitHub pull request to draft")
	}

	return p.convertGitHubPRToMR(ghPR, owner, repo), nil
}

// MarkReadyForReview marks a pull request as ready for review
func (p *GitHubProvider) MarkReadyForReview(ctx context.Context, projectID string, mrID string) (*provider.MergeRequest, error) {
	// Parse projectID to extract owner/repo
	owner, repo, err := parseProjectID(projectID)
	if err != nil {
		return nil, err
	}

	ghPR, err := p.client.MarkPullRequestReadyForReview(ctx, mrID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to mark GitHub pull request ready for review")
	}

	return p.convertGitHubPRToMR(ghPR, owner, repo), nil
}

// RequestReviews requests reviews from specified users
func (p *GitHubProvider) RequestReviews(ctx context.Context, projectID string, mrID string, reviewers []string) (*provider.MergeRequest, error) {
	// Parse projectID to extract owner/repo
	owner, repo, err := parseProjectID(projectID)
	if err != nil {
		return nil, err
	}

	// Convert reviewer strings to GitHub user IDs
	var userIDs []githubv4.ID
	for _, reviewer := range reviewers {
		userIDs = append(userIDs, githubv4.ID(reviewer))
	}

	ghInput := githubv4.RequestReviewsInput{
		PullRequestID: githubv4.ID(mrID),
		UserIDs:       &userIDs,
		Union:         Ptr[githubv4.Boolean](true), // Add reviewers instead of replacing
	}

	ghPR, err := p.client.RequestReviews(ctx, ghInput)
	if err != nil {
		return nil, errors.Wrap(err, "failed to request reviews on GitHub pull request")
	}

	return p.convertGitHubPRToMR(ghPR, owner, repo), nil
}

// Helper functions

// parseProjectID parses a "owner/repo" format project ID
func parseProjectID(projectID string) (owner, repo string, err error) {
	parts := strings.Split(projectID, "/")
	if len(parts) != 2 {
		return "", "", errors.Errorf("invalid project ID format, expected 'owner/repo': %s", projectID)
	}
	return parts[0], parts[1], nil
}

// convertGitHubPRToMR converts a GitHub PullRequest to a unified MergeRequest
func (p *GitHubProvider) convertGitHubPRToMR(ghPR *PullRequest, owner, repo string) *provider.MergeRequest {
	return &provider.MergeRequest{
		ID:           ghPR.ID,
		Number:       ghPR.Number,
		ProjectID:    owner + "/" + repo,
		Title:        ghPR.Title,
		Description:  ghPR.Body,
		State:        convertGitHubStateToMR(ghPR.State),
		IsDraft:      ghPR.IsDraft,
		SourceBranch: ghPR.HeadBranchName(),
		TargetBranch: ghPR.BaseBranchName(),
		WebURL:       ghPR.Permalink,
		MergeCommit:  ghPR.GetMergeCommit(),
	}
}

// convertGitHubStateToMR converts GitHub PR state to unified MR state
func convertGitHubStateToMR(state githubv4.PullRequestState) provider.MRState {
	switch state {
	case githubv4.PullRequestStateOpen:
		return provider.MRStateOpen
	case githubv4.PullRequestStateClosed:
		return provider.MRStateClosed
	case githubv4.PullRequestStateMerged:
		return provider.MRStateMerged
	default:
		return provider.MRStateOpen
	}
}

// convertMRStateToGitHub converts unified MR state to GitHub PR state
func convertMRStateToGitHub(state provider.MRState) githubv4.PullRequestState {
	switch state {
	case provider.MRStateOpen:
		return githubv4.PullRequestStateOpen
	case provider.MRStateClosed:
		return githubv4.PullRequestStateClosed
	case provider.MRStateMerged:
		return githubv4.PullRequestStateMerged
	default:
		return githubv4.PullRequestStateOpen
	}
}