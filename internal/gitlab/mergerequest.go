package gitlab

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
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
	ID           int64  `json:"id"`
	IID          int64  `json:"iid"`
	ProjectID    int64  `json:"project_id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	State        string `json:"state"`
	MergeStatus  string `json:"merge_status"`
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
	WebURL       string `json:"web_url"`
	DraftStatus  bool   `json:"work_in_progress"`
	MergeCommit  struct {
		ID string `json:"id"`
	} `json:"merge_commit_sha"`
}

// HeadBranchName returns the source branch name
func (mr *MergeRequest) HeadBranchName() string {
	return strings.TrimPrefix(mr.SourceBranch, "refs/heads/")
}

// BaseBranchName returns the target branch name
func (mr *MergeRequest) BaseBranchName() string {
	return strings.TrimPrefix(mr.TargetBranch, "refs/heads/")
}

// IsDraft returns true if the merge request is marked as work in progress/draft
func (mr *MergeRequest) IsDraft() bool {
	return mr.DraftStatus
}

// GetMergeCommit returns the merge commit SHA if the MR is merged
func (mr *MergeRequest) GetMergeCommit() string {
	if mr.State == string(MergeRequestStateMerged) && mr.MergeCommit.ID != "" {
		return mr.MergeCommit.ID
	}
	return ""
}

// CreateMergeRequestInput represents the input for creating a merge request
type CreateMergeRequestInput struct {
	ProjectID    string `json:"-"` // Used in URL, not body
	Title        string `json:"title"`
	Description  string `json:"description,omitempty"`
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
	Draft        bool   `json:"-"` // Handled separately with work_in_progress field
}

// UpdateMergeRequestInput represents the input for updating a merge request
type UpdateMergeRequestInput struct {
	ProjectID   string  `json:"-"` // Used in URL, not body
	IID         int64   `json:"-"` // Used in URL, not body
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	StateEvent  *string `json:"state_event,omitempty"` // "close", "reopen"
	Draft       *bool   `json:"-"`                     // Handled separately with work_in_progress field
}

// GetMergeRequestsInput represents the input for querying merge requests
type GetMergeRequestsInput struct {
	ProjectID    string              `json:"-"` // Used in URL, not body
	State        []MergeRequestState `json:"-"` // Used as query param
	SourceBranch string              `json:"-"` // Used as query param
	TargetBranch string              `json:"-"` // Used as query param
	Page         int                 `json:"-"` // Used as query param
	PerPage      int                 `json:"-"` // Used as query param
}

// GetMergeRequestsPage represents a page of merge requests with pagination info
type GetMergeRequestsPage struct {
	MergeRequests []MergeRequest
	NextPage      int
	HasNextPage   bool
	TotalPages    int
	TotalCount    int
}

// CreateMergeRequest creates a new merge request in GitLab
func (c *Client) CreateMergeRequest(ctx context.Context, input CreateMergeRequestInput) (*MergeRequest, error) {
	endpoint := fmt.Sprintf("/projects/%s/merge_requests", url.PathEscape(input.ProjectID))
	
	// Prepare the request body
	reqBody := map[string]interface{}{
		"title":         input.Title,
		"source_branch": input.SourceBranch,
		"target_branch": input.TargetBranch,
	}
	
	if input.Description != "" {
		reqBody["description"] = input.Description
	}
	
	// Handle draft status - GitLab uses work_in_progress field
	if input.Draft {
		reqBody["work_in_progress"] = true
	}

	var result MergeRequest
	if err := c.post(ctx, endpoint, reqBody, &result); err != nil {
		return nil, errors.Wrap(err, "failed to create merge request")
	}

	return &result, nil
}

// UpdateMergeRequest updates an existing merge request in GitLab
func (c *Client) UpdateMergeRequest(ctx context.Context, input UpdateMergeRequestInput) (*MergeRequest, error) {
	endpoint := fmt.Sprintf("/projects/%s/merge_requests/%d", url.PathEscape(input.ProjectID), input.IID)
	
	// Prepare the request body with only non-nil fields
	reqBody := make(map[string]interface{})
	
	if input.Title != nil {
		reqBody["title"] = *input.Title
	}
	
	if input.Description != nil {
		reqBody["description"] = *input.Description
	}
	
	if input.StateEvent != nil {
		reqBody["state_event"] = *input.StateEvent
	}
	
	// Handle draft status separately
	if input.Draft != nil {
		reqBody["work_in_progress"] = *input.Draft
	}

	var result MergeRequest
	if err := c.put(ctx, endpoint, reqBody, &result); err != nil {
		return nil, errors.Wrap(err, "failed to update merge request")
	}

	return &result, nil
}

// GetMergeRequests retrieves merge requests with filtering and pagination support
func (c *Client) GetMergeRequests(ctx context.Context, input GetMergeRequestsInput) (*GetMergeRequestsPage, error) {
	endpoint := fmt.Sprintf("/projects/%s/merge_requests", url.PathEscape(input.ProjectID))
	
	// Build query parameters
	params := make(map[string]string)
	
	if len(input.State) > 0 {
		// Convert states to comma-separated string
		states := make([]string, len(input.State))
		for i, state := range input.State {
			states[i] = string(state)
		}
		params["state"] = strings.Join(states, ",")
	}
	
	if input.SourceBranch != "" {
		params["source_branch"] = input.SourceBranch
	}
	
	if input.TargetBranch != "" {
		params["target_branch"] = input.TargetBranch
	}
	
	// Handle pagination
	if input.Page > 0 {
		params["page"] = strconv.Itoa(input.Page)
	} else {
		params["page"] = "1"
	}
	
	if input.PerPage > 0 {
		params["per_page"] = strconv.Itoa(input.PerPage)
	} else {
		params["per_page"] = "50" // Default page size
	}

	// Build the URL with query parameters
	fullEndpoint := c.buildURL(endpoint, params)

	var mergeRequests []MergeRequest
	if err := c.get(ctx, fullEndpoint, &mergeRequests); err != nil {
		return nil, errors.Wrap(err, "failed to get merge requests")
	}

	// TODO: Parse pagination headers from response to populate pagination info
	// For now, we'll return basic pagination info
	result := &GetMergeRequestsPage{
		MergeRequests: mergeRequests,
		HasNextPage:   len(mergeRequests) == input.PerPage, // Simple heuristic
		NextPage:      input.Page + 1,
	}

	return result, nil
}

// ConvertToDraft converts a merge request to draft status using the work_in_progress flag
func (c *Client) ConvertToDraft(ctx context.Context, projectID string, iid int64) (*MergeRequest, error) {
	draft := true
	input := UpdateMergeRequestInput{
		ProjectID: projectID,
		IID:       iid,
		Draft:     &draft,
	}
	
	return c.UpdateMergeRequest(ctx, input)
}

// MarkReadyForReview marks a merge request as ready for review by removing the work_in_progress flag
func (c *Client) MarkReadyForReview(ctx context.Context, projectID string, iid int64) (*MergeRequest, error) {
	draft := false
	input := UpdateMergeRequestInput{
		ProjectID: projectID,
		IID:       iid,
		Draft:     &draft,
	}
	
	return c.UpdateMergeRequest(ctx, input)
}

// GetMergeRequest retrieves a single merge request by project ID and IID
func (c *Client) GetMergeRequest(ctx context.Context, projectID string, iid int64) (*MergeRequest, error) {
	endpoint := fmt.Sprintf("/projects/%s/merge_requests/%d", url.PathEscape(projectID), iid)
	
	var result MergeRequest
	if err := c.get(ctx, endpoint, &result); err != nil {
		if IsHTTPNotFound(err) {
			return nil, errors.Errorf("merge request !%d not found in project %s", iid, projectID)
		}
		return nil, errors.Wrap(err, "failed to get merge request")
	}

	return &result, nil
}

// GetMergeRequestByID retrieves a single merge request by its global ID
func (c *Client) GetMergeRequestByID(ctx context.Context, projectID string, id int64) (*MergeRequest, error) {
	// First, get all MRs and find the one with matching ID
	// This is less efficient but GitLab doesn't have a direct endpoint for global ID lookup
	input := GetMergeRequestsInput{
		ProjectID: projectID,
		State:     []MergeRequestState{MergeRequestStateOpened, MergeRequestStateClosed, MergeRequestStateMerged},
		PerPage:   100, // Get a large page to search through
	}
	
	page := 1
	for {
		input.Page = page
		result, err := c.GetMergeRequests(ctx, input)
		if err != nil {
			return nil, errors.Wrap(err, "failed to search for merge request by ID")
		}
		
		// Search for the MR with matching ID
		for _, mr := range result.MergeRequests {
			if mr.ID == id {
				return &mr, nil
			}
		}
		
		// If we've reached the end or no more pages, break
		if !result.HasNextPage || len(result.MergeRequests) == 0 {
			break
		}
		
		page++
		if page > 10 { // Prevent infinite loop, limit search to first 10 pages
			break
		}
	}
	
	return nil, errors.Errorf("merge request with ID %d not found in project %s", id, projectID)
}