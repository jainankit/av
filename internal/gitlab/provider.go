// Package gitlab provides the GitLab implementation of the Provider interface.
// This adapter implements the unified Provider interface using GitLab's REST API,
// mapping GitLab-specific types and operations to the common provider abstractions.
package gitlab

import (
	"context"
	"strconv"

	"emperror.dev/errors"
	"github.com/aviator-co/av/internal/provider"
)

// GitLabProvider implements the Provider interface for GitLab
type GitLabProvider struct {
	client *Client
}

// NewProvider creates a new GitLab provider
func NewProvider(ctx context.Context, token, baseURL string) (*GitLabProvider, error) {
	client, err := NewClient(ctx, token, baseURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create GitLab client")
	}

	return &GitLabProvider{
		client: client,
	}, nil
}

// CreateMergeRequest implements Provider.CreateMergeRequest
func (p *GitLabProvider) CreateMergeRequest(ctx context.Context, input provider.CreateMRInput) (*provider.MergeRequest, error) {
	gitlabInput := CreateMergeRequestInput{
		ProjectID:    input.ProjectID,
		Title:        input.Title,
		Description:  input.Description,
		SourceBranch: input.SourceBranch,
		TargetBranch: input.TargetBranch,
		Draft:        input.Draft,
	}

	gitlabMR, err := p.client.CreateMergeRequest(ctx, gitlabInput)
	if err != nil {
		return nil, p.mapError(err)
	}

	return p.convertMergeRequest(gitlabMR, input.ProjectID), nil
}

// UpdateMergeRequest implements Provider.UpdateMergeRequest
func (p *GitLabProvider) UpdateMergeRequest(ctx context.Context, input provider.UpdateMRInput) (*provider.MergeRequest, error) {
	// Convert string MRID to int64 for GitLab IID
	iid, err := strconv.ParseInt(input.MRID, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "invalid merge request ID")
	}

	gitlabInput := UpdateMergeRequestInput{
		ProjectID:   input.ProjectID,
		IID:         iid,
		Title:       input.Title,
		Description: input.Description,
		Draft:       input.Draft,
	}

	// Handle state events
	if input.StateEvent != nil {
		gitlabInput.StateEvent = input.StateEvent
	}

	gitlabMR, err := p.client.UpdateMergeRequest(ctx, gitlabInput)
	if err != nil {
		return nil, p.mapError(err)
	}

	return p.convertMergeRequest(gitlabMR, input.ProjectID), nil
}

// GetMergeRequests implements Provider.GetMergeRequests
func (p *GitLabProvider) GetMergeRequests(ctx context.Context, input provider.GetMRsInput) (*provider.GetMRsPage, error) {
	// Convert unified states to GitLab states
	var gitlabStates []MergeRequestState
	for _, state := range input.States {
		gitlabStates = append(gitlabStates, p.convertStateToGitLab(state))
	}

	gitlabInput := GetMergeRequestsInput{
		ProjectID:    input.ProjectID,
		State:        gitlabStates,
		SourceBranch: input.SourceBranch,
		TargetBranch: input.TargetBranch,
		Page:         input.Page,
		PerPage:      input.PerPage,
	}

	gitlabPage, err := p.client.GetMergeRequests(ctx, gitlabInput)
	if err != nil {
		return nil, p.mapError(err)
	}

	// Convert GitLab MRs to unified format
	var unifiedMRs []provider.MergeRequest
	for _, gitlabMR := range gitlabPage.MergeRequests {
		unifiedMRs = append(unifiedMRs, *p.convertMergeRequest(&gitlabMR, input.ProjectID))
	}

	return &provider.GetMRsPage{
		MergeRequests: unifiedMRs,
		NextPage:      gitlabPage.NextPage,
		HasNextPage:   gitlabPage.HasNextPage,
		TotalPages:    gitlabPage.TotalPages,
		TotalCount:    gitlabPage.TotalCount,
	}, nil
}

// GetRepository implements Provider.GetRepository
func (p *GitLabProvider) GetRepository(ctx context.Context, owner, repo string) (*provider.Repository, error) {
	slug := owner + "/" + repo
	gitlabRepo, err := p.client.GetRepositoryBySlug(ctx, slug)
	if err != nil {
		return nil, p.mapError(err)
	}

	return &provider.Repository{
		ID:       strconv.FormatInt(gitlabRepo.ID, 10),
		Name:     gitlabRepo.GetName(),
		FullName: gitlabRepo.GetSlug(),
		Owner:    gitlabRepo.GetOwner(),
		WebURL:   gitlabRepo.WebURL,
	}, nil
}

// ConvertToDraft implements Provider.ConvertToDraft
func (p *GitLabProvider) ConvertToDraft(ctx context.Context, projectID string, mrID string) (*provider.MergeRequest, error) {
	// Convert string mrID to int64 for GitLab IID
	iid, err := strconv.ParseInt(mrID, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "invalid merge request ID")
	}

	gitlabMR, err := p.client.ConvertToDraft(ctx, projectID, iid)
	if err != nil {
		return nil, p.mapError(err)
	}

	return p.convertMergeRequest(gitlabMR, projectID), nil
}

// MarkReadyForReview implements Provider.MarkReadyForReview
func (p *GitLabProvider) MarkReadyForReview(ctx context.Context, projectID string, mrID string) (*provider.MergeRequest, error) {
	// Convert string mrID to int64 for GitLab IID
	iid, err := strconv.ParseInt(mrID, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "invalid merge request ID")
	}

	gitlabMR, err := p.client.MarkReadyForReview(ctx, projectID, iid)
	if err != nil {
		return nil, p.mapError(err)
	}

	return p.convertMergeRequest(gitlabMR, projectID), nil
}

// RequestReviews implements Provider.RequestReviews
func (p *GitLabProvider) RequestReviews(ctx context.Context, projectID string, mrID string, reviewers []string) (*provider.MergeRequest, error) {
	// GitLab doesn't have a direct "request reviews" API like GitHub
	// Instead, we would need to mention users in comments or use approvers API
	// For now, return an error indicating this feature isn't implemented
	return nil, provider.ErrNotImplemented
}

// Helper methods for type conversion

// convertMergeRequest converts a GitLab MergeRequest to the unified format
func (p *GitLabProvider) convertMergeRequest(gitlabMR *MergeRequest, projectID string) *provider.MergeRequest {
	return &provider.MergeRequest{
		ID:           strconv.FormatInt(gitlabMR.IID, 10), // Use IID as the unified ID
		Number:       gitlabMR.IID,
		ProjectID:    projectID,
		Title:        gitlabMR.Title,
		Description:  gitlabMR.Description,
		State:        p.convertStateFromGitLab(MergeRequestState(gitlabMR.State)),
		IsDraft:      gitlabMR.IsDraft(),
		SourceBranch: gitlabMR.HeadBranchName(),
		TargetBranch: gitlabMR.BaseBranchName(),
		WebURL:       gitlabMR.WebURL,
		MergeCommit:  gitlabMR.GetMergeCommit(),
	}
}

// convertStateFromGitLab converts GitLab MergeRequestState to unified MRState
func (p *GitLabProvider) convertStateFromGitLab(gitlabState MergeRequestState) provider.MRState {
	switch gitlabState {
	case MergeRequestStateOpened:
		return provider.MRStateOpen
	case MergeRequestStateClosed:
		return provider.MRStateClosed
	case MergeRequestStateMerged:
		return provider.MRStateMerged
	default:
		return provider.MRStateOpen // default fallback
	}
}

// convertStateToGitLab converts unified MRState to GitLab MergeRequestState
func (p *GitLabProvider) convertStateToGitLab(unifiedState provider.MRState) MergeRequestState {
	switch unifiedState {
	case provider.MRStateOpen:
		return MergeRequestStateOpened
	case provider.MRStateClosed:
		return MergeRequestStateClosed
	case provider.MRStateMerged:
		return MergeRequestStateMerged
	default:
		return MergeRequestStateOpened // default fallback
	}
}

// mapError maps GitLab-specific errors to unified provider errors
func (p *GitLabProvider) mapError(err error) error {
	if IsHTTPNotFound(err) {
		return provider.ErrMergeRequestNotFound
	}
	if IsHTTPUnauthorized(err) {
		return provider.ErrUnauthorized
	}
	if IsHTTPForbidden(err) {
		return provider.ErrForbidden
	}
	if IsHTTPConflict(err) {
		return provider.ErrConflict
	}
	// Return the original error if it doesn't match known patterns
	return err
}