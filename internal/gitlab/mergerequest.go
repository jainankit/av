package gitlab

import (
	"context"
	"strings"

	"emperror.dev/errors"
)

// MergeRequestState represents the state of a GitLab merge request
type MergeRequestState string

const (
	MergeRequestStateOpened MergeRequestState = "opened"
	MergeRequestStateClosed MergeRequestState = "closed" 
	MergeRequestStateMerged MergeRequestState = "merged"
)

// MergeRequest represents a GitLab merge request
type MergeRequest struct {
	ID              string            `graphql:"id"`
	IID             string            `graphql:"iid"`
	Title           string            `graphql:"title"`
	Description     string            `graphql:"description"`
	State           MergeRequestState `graphql:"state"`
	Draft           bool              `graphql:"draft"`
	WebURL          string            `graphql:"webUrl"`
	SourceBranch    string            `graphql:"sourceBranch"`
	TargetBranch    string            `graphql:"targetBranch"`
	PRIVATE_MergeCommitSha struct {
		Sha string `graphql:"sha"`
	} `graphql:"mergeCommitSha"`
	PRIVATE_Pipelines struct {
		Nodes []struct {
			Commit struct {
				Sha string `graphql:"sha"`
			} `graphql:"commit"`
		} `graphql:"nodes"`
	} `graphql:"pipelines(first: 10)"`
}

func (mr *MergeRequest) HeadBranchName() string {
	// GitLab typically stores branch names without the "refs/heads/" prefix
	return strings.TrimPrefix(mr.SourceBranch, "refs/heads/")
}

func (mr *MergeRequest) BaseBranchName() string {
	return strings.TrimPrefix(mr.TargetBranch, "refs/heads/")
}

func (mr *MergeRequest) GetMergeCommit() string {
	if mr.State == MergeRequestStateOpened {
		return ""
	} else if mr.State == MergeRequestStateMerged && mr.PRIVATE_MergeCommitSha.Sha != "" {
		return mr.PRIVATE_MergeCommitSha.Sha
	}
	
	// For closed/merged MRs, check the pipeline commits
	for _, pipeline := range mr.PRIVATE_Pipelines.Nodes {
		if pipeline.Commit.Sha != "" {
			return pipeline.Commit.Sha
		}
	}
	return ""
}

// MergeRequestOpts contains options for fetching a specific merge request
type MergeRequestOpts struct {
	ProjectID string
	IID       string
}

// GetMergeRequest fetches a single merge request by ID
func (c *Client) GetMergeRequest(ctx context.Context, id string) (*MergeRequest, error) {
	var query struct {
		MergeRequest MergeRequest `graphql:"mergeRequest(id: $id)"`
	}
	
	variables := map[string]any{
		"id": id,
	}
	
	if err := c.query(ctx, &query, variables); err != nil {
		return nil, errors.Wrap(err, "failed to query merge request")
	}
	
	if query.MergeRequest.ID == "" {
		return nil, errors.Errorf("merge request %q not found", id)
	}
	
	return &query.MergeRequest, nil
}

// GetMergeRequestsInput contains parameters for fetching merge requests
type GetMergeRequestsInput struct {
	// REQUIRED
	ProjectPath string // namespace/project format
	// OPTIONAL
	SourceBranch string
	TargetBranch string
	State        MergeRequestState
	First        int32
	After        string
}

// GetMergeRequestsPage represents a paginated response of merge requests
type GetMergeRequestsPage struct {
	PageInfo     PageInfo
	MergeRequests []MergeRequest
}

// GetMergeRequests fetches merge requests with pagination support
func (c *Client) GetMergeRequests(
	ctx context.Context,
	input GetMergeRequestsInput,
) (*GetMergeRequestsPage, error) {
	if input.First == 0 {
		input.First = 50
	}
	
	var query struct {
		Project struct {
			MergeRequests struct {
				Nodes    []MergeRequest `graphql:"nodes"`
				PageInfo PageInfo       `graphql:"pageInfo"`
			} `graphql:"mergeRequests(sourceBranches: $sourceBranches, targetBranches: $targetBranches, state: $state, first: $first, after: $after)"`
		} `graphql:"project(fullPath: $projectPath)"`
	}
	
	variables := map[string]any{
		"projectPath": input.ProjectPath,
		"first":       input.First,
	}
	
	// Add optional filters
	if input.SourceBranch != "" {
		variables["sourceBranches"] = []string{input.SourceBranch}
	}
	if input.TargetBranch != "" {
		variables["targetBranches"] = []string{input.TargetBranch}
	}
	if input.State != "" {
		variables["state"] = input.State
	}
	if input.After != "" {
		variables["after"] = input.After
	}
	
	if err := c.query(ctx, &query, variables); err != nil {
		return nil, errors.Wrap(err, "failed to query merge requests")
	}
	
	return &GetMergeRequestsPage{
		PageInfo:      query.Project.MergeRequests.PageInfo,
		MergeRequests: query.Project.MergeRequests.Nodes,
	}, nil
}

// CreateMergeRequestInput represents the input for creating a merge request
type CreateMergeRequestInput struct {
	ProjectPath  string `json:"projectPath"`
	Title        string `json:"title"`
	Description  string `json:"description,omitempty"`
	SourceBranch string `json:"sourceBranch"`
	TargetBranch string `json:"targetBranch"`
	Draft        bool   `json:"draft,omitempty"`
}

// CreateMergeRequest creates a new merge request
func (c *Client) CreateMergeRequest(
	ctx context.Context,
	input CreateMergeRequestInput,
) (*MergeRequest, error) {
	var mutation struct {
		MergeRequestCreate struct {
			MergeRequest MergeRequest `graphql:"mergeRequest"`
			Errors       []string     `graphql:"errors"`
		} `graphql:"mergeRequestCreate(input: $input)"`
	}
	
	variables := map[string]any{
		"input": map[string]any{
			"projectPath":  input.ProjectPath,
			"title":        input.Title,
			"sourceBranch": input.SourceBranch,
			"targetBranch": input.TargetBranch,
		},
	}
	
	// Add optional fields
	if input.Description != "" {
		variables["input"].(map[string]any)["description"] = input.Description
	}
	if input.Draft {
		variables["input"].(map[string]any)["draft"] = input.Draft
	}
	
	if err := c.mutate(ctx, &mutation, variables); err != nil {
		return nil, errors.Wrap(err, "failed to create merge request: gitlab error")
	}
	
	if len(mutation.MergeRequestCreate.Errors) > 0 {
		return nil, errors.Errorf("failed to create merge request: %v", mutation.MergeRequestCreate.Errors)
	}
	
	return &mutation.MergeRequestCreate.MergeRequest, nil
}

// UpdateMergeRequestInput represents the input for updating a merge request
type UpdateMergeRequestInput struct {
	ProjectPath string                `json:"projectPath"`
	IID         string                `json:"iid"`
	Title       *string               `json:"title,omitempty"`
	Description *string               `json:"description,omitempty"`
	State       *MergeRequestState    `json:"state,omitempty"`
	TargetBranch *string              `json:"targetBranch,omitempty"`
}

// UpdateMergeRequest updates an existing merge request
func (c *Client) UpdateMergeRequest(
	ctx context.Context,
	input UpdateMergeRequestInput,
) (*MergeRequest, error) {
	var mutation struct {
		MergeRequestUpdate struct {
			MergeRequest MergeRequest `graphql:"mergeRequest"`
			Errors       []string     `graphql:"errors"`
		} `graphql:"mergeRequestUpdate(input: $input)"`
	}
	
	mutationInput := map[string]any{
		"projectPath": input.ProjectPath,
		"iid":         input.IID,
	}
	
	// Add optional fields
	if input.Title != nil {
		mutationInput["title"] = *input.Title
	}
	if input.Description != nil {
		mutationInput["description"] = *input.Description
	}
	if input.State != nil {
		mutationInput["state"] = *input.State
	}
	if input.TargetBranch != nil {
		mutationInput["targetBranch"] = *input.TargetBranch
	}
	
	variables := map[string]any{
		"input": mutationInput,
	}
	
	if err := c.mutate(ctx, &mutation, variables); err != nil {
		return nil, errors.Wrap(err, "failed to update merge request: gitlab error")
	}
	
	if len(mutation.MergeRequestUpdate.Errors) > 0 {
		return nil, errors.Errorf("failed to update merge request: %v", mutation.MergeRequestUpdate.Errors)
	}
	
	return &mutation.MergeRequestUpdate.MergeRequest, nil
}

// RequestReviewsInput represents input for requesting reviews on a merge request
type RequestReviewsInput struct {
	ProjectPath string   `json:"projectPath"`
	IID         string   `json:"iid"`
	UserIDs     []string `json:"userIds"`
}

// RequestReviews requests reviews from specified users on a merge request
func (c *Client) RequestReviews(
	ctx context.Context,
	input RequestReviewsInput,
) (*MergeRequest, error) {
	var mutation struct {
		MergeRequestSetReviewers struct {
			MergeRequest MergeRequest `graphql:"mergeRequest"`
			Errors       []string     `graphql:"errors"`
		} `graphql:"mergeRequestSetReviewers(input: $input)"`
	}
	
	variables := map[string]any{
		"input": map[string]any{
			"projectPath": input.ProjectPath,
			"iid":         input.IID,
			"reviewerIds": input.UserIDs,
		},
	}
	
	if err := c.mutate(ctx, &mutation, variables); err != nil {
		return nil, errors.Wrap(err, "failed to request merge request reviews")
	}
	
	if len(mutation.MergeRequestSetReviewers.Errors) > 0 {
		return nil, errors.Errorf("failed to request merge request reviews: %v", mutation.MergeRequestSetReviewers.Errors)
	}
	
	return &mutation.MergeRequestSetReviewers.MergeRequest, nil
}

// ConvertMergeRequestToDraft converts a merge request to draft status
func (c *Client) ConvertMergeRequestToDraft(ctx context.Context, projectPath, iid string) (*MergeRequest, error) {
	return c.UpdateMergeRequest(ctx, UpdateMergeRequestInput{
		ProjectPath: projectPath,
		IID:         iid,
		State:       Ptr(MergeRequestStateOpened), // GitLab uses draft field separately
	})
}

// MarkMergeRequestReadyForReview marks a merge request as ready for review (removes draft status)
func (c *Client) MarkMergeRequestReadyForReview(ctx context.Context, projectPath, iid string) (*MergeRequest, error) {
	var mutation struct {
		MergeRequestSetDraft struct {
			MergeRequest MergeRequest `graphql:"mergeRequest"`
			Errors       []string     `graphql:"errors"`
		} `graphql:"mergeRequestSetDraft(input: $input)"`
	}
	
	variables := map[string]any{
		"input": map[string]any{
			"projectPath": projectPath,
			"iid":         iid,
			"draft":       false,
		},
	}
	
	if err := c.mutate(ctx, &mutation, variables); err != nil {
		return nil, errors.Wrap(err, "failed to mark merge request ready for review: gitlab error")
	}
	
	if len(mutation.MergeRequestSetDraft.Errors) > 0 {
		return nil, errors.Errorf("failed to mark merge request ready for review: %v", mutation.MergeRequestSetDraft.Errors)
	}
	
	return &mutation.MergeRequestSetDraft.MergeRequest, nil
}

// RepoMergeRequestOpts contains options for fetching repository merge requests
type RepoMergeRequestOpts struct {
	ProjectPath string
	First       int32
	After       string
	State       MergeRequestState
}

// RepoMergeRequestsResponse represents the response for repository merge requests
type RepoMergeRequestsResponse struct {
	PageInfo      PageInfo
	TotalCount    int64
	MergeRequests []MergeRequest
}

// RepoMergeRequests fetches merge requests for a repository with pagination
func (c *Client) RepoMergeRequests(
	ctx context.Context,
	opts RepoMergeRequestOpts,
) (RepoMergeRequestsResponse, error) {
	var query struct {
		Project struct {
			MergeRequests struct {
				Count    int64         `graphql:"count"`
				PageInfo PageInfo      `graphql:"pageInfo"`
				Nodes    []MergeRequest `graphql:"nodes"`
			} `graphql:"mergeRequests(state: $state, first: $first, after: $after)"`
		} `graphql:"project(fullPath: $projectPath)"`
	}
	
	if opts.First == 0 {
		opts.First = 100
	}
	
	variables := map[string]any{
		"projectPath": opts.ProjectPath,
		"first":       opts.First,
	}
	
	if opts.After != "" {
		variables["after"] = opts.After
	}
	if opts.State != "" {
		variables["state"] = opts.State
	}
	
	if err := c.query(ctx, &query, variables); err != nil {
		return RepoMergeRequestsResponse{}, errors.Wrap(err, "failed to query merge requests")
	}
	
	return RepoMergeRequestsResponse{
		PageInfo:      query.Project.MergeRequests.PageInfo,
		TotalCount:    query.Project.MergeRequests.Count,
		MergeRequests: query.Project.MergeRequests.Nodes,
	}, nil
}