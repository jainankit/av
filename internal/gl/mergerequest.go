package gl

import (
	"context"
	"fmt"
	"strings"
	"time"

	"emperror.dev/errors"
)

type MergeRequestState string

const (
	MergeRequestStateOpened MergeRequestState = "opened"
	MergeRequestStateClosed MergeRequestState = "closed"
	MergeRequestStateMerged MergeRequestState = "merged"
)

type MergeRequest struct {
	ID                    int                `json:"id"`
	IID                   int                `json:"iid"` // Project-scoped ID
	Title                 string             `json:"title"`
	Description           string             `json:"description"`
	State                 MergeRequestState  `json:"state"`
	CreatedAt             time.Time          `json:"created_at"`
	UpdatedAt             time.Time          `json:"updated_at"`
	MergedAt              *time.Time         `json:"merged_at"`
	ClosedAt              *time.Time         `json:"closed_at"`
	SourceBranch          string             `json:"source_branch"`
	TargetBranch          string             `json:"target_branch"`
	WebURL                string             `json:"web_url"`
	MergeCommitSHA        *string            `json:"merge_commit_sha"`
	SquashCommitSHA       *string            `json:"squash_commit_sha"`
	WorkInProgress        bool               `json:"work_in_progress"`
	Draft                 bool               `json:"draft"`
	Author                User               `json:"author"`
	Assignees             []User             `json:"assignees"`
	Assignee              *User              `json:"assignee"`
	Reviewers             []User             `json:"reviewers"`
	SourceProjectID       int                `json:"source_project_id"`
	TargetProjectID       int                `json:"target_project_id"`
	Labels                []string           `json:"labels"`
	MilestoneID           *int               `json:"milestone_id"`
	ForceRemoveSourceBranch bool             `json:"force_remove_source_branch"`
	AllowCollaboration    bool               `json:"allow_collaboration"`
	AllowMaintainerToPush bool               `json:"allow_maintainer_to_push"`
	SubscribedToDiscussion bool              `json:"subscribed_to_discussion"`
	ChangesCount          string             `json:"changes_count"`
	UserNotesCount        int                `json:"user_notes_count"`
	UpvotesCount          int                `json:"upvotes_count"`
	DownvotesCount        int                `json:"downvotes_count"`
	MergeStatus           string             `json:"merge_status"`
	SHA                   string             `json:"sha"`
	MergeCommitMessage    *string            `json:"merge_commit_message"`
	SquashCommitMessage   *string            `json:"squash_commit_message"`
	ProjectID             int                `json:"project_id"`
	HasConflicts          bool               `json:"has_conflicts"`
	BlockingDiscussionsResolved bool         `json:"blocking_discussions_resolved"`
}

func (mr *MergeRequest) HeadBranchName() string {
	return strings.TrimPrefix(mr.SourceBranch, "refs/heads/")
}

func (mr *MergeRequest) BaseBranchName() string {
	return strings.TrimPrefix(mr.TargetBranch, "refs/heads/")
}

func (mr *MergeRequest) GetMergeCommit() string {
	if mr.State != MergeRequestStateMerged {
		return ""
	}
	
	if mr.MergeCommitSHA != nil && *mr.MergeCommitSHA != "" {
		return *mr.MergeCommitSHA
	}
	
	if mr.SquashCommitSHA != nil && *mr.SquashCommitSHA != "" {
		return *mr.SquashCommitSHA
	}
	
	return ""
}

type GetMergeRequestsInput struct {
	// REQUIRED
	ProjectID int
	// OPTIONAL
	SourceBranch *string
	TargetBranch *string
	State        *MergeRequestState
	PerPage      int
	Page         int
}

type GetMergeRequestsPage struct {
	PageInfo
	MergeRequests []MergeRequest
}

type CreateMergeRequestInput struct {
	ProjectID                 int     `json:"project_id"`
	SourceBranch             string  `json:"source_branch"`
	TargetBranch             string  `json:"target_branch"`
	Title                    string  `json:"title"`
	Description              *string `json:"description,omitempty"`
	TargetProjectID          *int    `json:"target_project_id,omitempty"`
	AssigneeID               *int    `json:"assignee_id,omitempty"`
	AssigneeIDs              []int   `json:"assignee_ids,omitempty"`
	ReviewerIDs              []int   `json:"reviewer_ids,omitempty"`
	MilestoneID              *int    `json:"milestone_id,omitempty"`
	Labels                   []string `json:"labels,omitempty"`
	RemoveSourceBranch       *bool   `json:"remove_source_branch,omitempty"`
	AllowCollaboration       *bool   `json:"allow_collaboration,omitempty"`
	AllowMaintainerToPush    *bool   `json:"allow_maintainer_to_push,omitempty"`
	Squash                   *bool   `json:"squash,omitempty"`
}

type UpdateMergeRequestInput struct {
	ProjectID                int     `json:"-"` // Used in URL path, not JSON body
	MergeRequestIID          int     `json:"-"` // Used in URL path, not JSON body
	Title                    *string `json:"title,omitempty"`
	Description              *string `json:"description,omitempty"`
	TargetBranch             *string `json:"target_branch,omitempty"`
	AssigneeID               *int    `json:"assignee_id,omitempty"`
	AssigneeIDs              []int   `json:"assignee_ids,omitempty"`
	ReviewerIDs              []int   `json:"reviewer_ids,omitempty"`
	MilestoneID              *int    `json:"milestone_id,omitempty"`
	Labels                   []string `json:"labels,omitempty"`
	StateEvent               *string `json:"state_event,omitempty"` // close, reopen, merge
	RemoveSourceBranch       *bool   `json:"remove_source_branch,omitempty"`
	Squash                   *bool   `json:"squash,omitempty"`
}

func (c *Client) GetMergeRequest(ctx context.Context, id int) (*MergeRequest, error) {
	path := fmt.Sprintf("/merge_requests/%d", id)
	
	var mr MergeRequest
	if err := c.get(ctx, path, nil, &mr); err != nil {
		return nil, errors.Wrapf(err, "failed to get merge request %d", id)
	}
	
	return &mr, nil
}

func (c *Client) GetMergeRequestByProject(ctx context.Context, projectID, mergeRequestIID int) (*MergeRequest, error) {
	path := fmt.Sprintf("/projects/%d/merge_requests/%d", projectID, mergeRequestIID)
	
	var mr MergeRequest
	if err := c.get(ctx, path, nil, &mr); err != nil {
		return nil, errors.Wrapf(err, "failed to get merge request %d from project %d", mergeRequestIID, projectID)
	}
	
	return &mr, nil
}

func (c *Client) GetMergeRequests(ctx context.Context, input GetMergeRequestsInput) (*GetMergeRequestsPage, error) {
	path := fmt.Sprintf("/projects/%d/merge_requests", input.ProjectID)
	
	params := make(map[string]string)
	if input.SourceBranch != nil {
		params["source_branch"] = *input.SourceBranch
	}
	if input.TargetBranch != nil {
		params["target_branch"] = *input.TargetBranch
	}
	if input.State != nil {
		params["state"] = string(*input.State)
	}
	
	perPage := input.PerPage
	if perPage == 0 {
		perPage = 50
	}
	params["per_page"] = fmt.Sprintf("%d", perPage)
	
	if input.Page > 0 {
		params["page"] = fmt.Sprintf("%d", input.Page)
	}
	
	var mrs []MergeRequest
	pageInfo, err := c.getPaginated(ctx, path, params, &mrs)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get merge requests for project %d", input.ProjectID)
	}
	
	return &GetMergeRequestsPage{
		PageInfo:      *pageInfo,
		MergeRequests: mrs,
	}, nil
}

func (c *Client) CreateMergeRequest(ctx context.Context, input CreateMergeRequestInput) (*MergeRequest, error) {
	path := fmt.Sprintf("/projects/%d/merge_requests", input.ProjectID)
	
	var mr MergeRequest
	if err := c.post(ctx, path, input, &mr); err != nil {
		return nil, errors.Wrap(err, "failed to create merge request")
	}
	
	return &mr, nil
}

func (c *Client) UpdateMergeRequest(ctx context.Context, input UpdateMergeRequestInput) (*MergeRequest, error) {
	path := fmt.Sprintf("/projects/%d/merge_requests/%d", input.ProjectID, input.MergeRequestIID)
	
	var mr MergeRequest
	if err := c.put(ctx, path, input, &mr); err != nil {
		return nil, errors.Wrapf(err, "failed to update merge request %d", input.MergeRequestIID)
	}
	
	return &mr, nil
}