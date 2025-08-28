package provider

import (
	"time"
)

// ProviderName represents the type of version control provider
type ProviderName string

const (
	ProviderGitHub ProviderName = "github"
	ProviderGitLab ProviderName = "gitlab"
)

// MergeRequestState represents the state of a merge request/pull request
type MergeRequestState string

const (
	MergeRequestStateOpen   MergeRequestState = "OPEN"
	MergeRequestStateClosed MergeRequestState = "CLOSED"
	MergeRequestStateMerged MergeRequestState = "MERGED"
)

// User represents a user on the version control platform
type User struct {
	ID    string `json:"id"`
	Login string `json:"login"`
}

// Repository represents a repository on the version control platform
type Repository struct {
	ID       string `json:"id"`
	Owner    string `json:"owner"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
}

// GetOwner returns the repository owner
func (r *Repository) GetOwner() string {
	return r.Owner
}

// GetName returns the repository name
func (r *Repository) GetName() string {
	return r.Name
}

// GetFullName returns the full repository name (owner/name)
func (r *Repository) GetFullName() string {
	if r.FullName != "" {
		return r.FullName
	}
	return r.Owner + "/" + r.Name
}

// MergeRequest represents a merge request/pull request with provider-agnostic fields
type MergeRequest struct {
	ID              string            `json:"id"`
	Number          int64             `json:"number"`
	Title           string            `json:"title"`
	Body            string            `json:"body"`
	State           MergeRequestState `json:"state"`
	IsDraft         bool              `json:"is_draft"`
	HeadRefName     string            `json:"head_ref_name"`
	BaseRefName     string            `json:"base_ref_name"`
	Permalink       string            `json:"permalink"`
	MergeCommitSha  string            `json:"merge_commit_sha,omitempty"`
	CreatedAt       *time.Time        `json:"created_at,omitempty"`
	UpdatedAt       *time.Time        `json:"updated_at,omitempty"`
	Author          *User             `json:"author,omitempty"`
}

// HeadBranchName returns the head branch name without refs/heads/ prefix
func (mr *MergeRequest) HeadBranchName() string {
	if mr.HeadRefName == "" {
		return ""
	}
	// Remove refs/heads/ prefix if present
	const prefix = "refs/heads/"
	if len(mr.HeadRefName) > len(prefix) && mr.HeadRefName[:len(prefix)] == prefix {
		return mr.HeadRefName[len(prefix):]
	}
	return mr.HeadRefName
}

// BaseBranchName returns the base branch name without refs/heads/ prefix  
func (mr *MergeRequest) BaseBranchName() string {
	if mr.BaseRefName == "" {
		return ""
	}
	// Remove refs/heads/ prefix if present
	const prefix = "refs/heads/"
	if len(mr.BaseRefName) > len(prefix) && mr.BaseRefName[:len(prefix)] == prefix {
		return mr.BaseRefName[len(prefix):]
	}
	return mr.BaseRefName
}

// GetMergeCommit returns the merge commit SHA if the MR is merged
func (mr *MergeRequest) GetMergeCommit() string {
	if mr.State == MergeRequestStateMerged {
		return mr.MergeCommitSha
	}
	return ""
}

// PageInfo contains pagination information
type PageInfo struct {
	EndCursor       string `json:"end_cursor"`
	HasNextPage     bool   `json:"has_next_page"`
	HasPreviousPage bool   `json:"has_previous_page"`
	StartCursor     string `json:"start_cursor"`
}

// GetMergeRequestsInput defines parameters for querying merge requests
type GetMergeRequestsInput struct {
	// Required
	Owner string `json:"owner"`
	Repo  string `json:"repo"`
	// Optional filters
	HeadRefName string              `json:"head_ref_name,omitempty"`
	BaseRefName string              `json:"base_ref_name,omitempty"`
	States      []MergeRequestState `json:"states,omitempty"`
	First       int32               `json:"first,omitempty"`
	After       string              `json:"after,omitempty"`
}

// GetMergeRequestsPage represents a page of merge requests with pagination info
type GetMergeRequestsPage struct {
	PageInfo     PageInfo       `json:"page_info"`
	MergeRequests []MergeRequest `json:"merge_requests"`
}

// CreateMergeRequestInput defines parameters for creating a new merge request
type CreateMergeRequestInput struct {
	RepositoryID string  `json:"repository_id"`
	Title        string  `json:"title"`
	Body         *string `json:"body,omitempty"`
	BaseRefName  string  `json:"base_ref_name"`
	HeadRefName  string  `json:"head_ref_name"`
	Draft        *bool   `json:"draft,omitempty"`
}

// UpdateMergeRequestInput defines parameters for updating a merge request
type UpdateMergeRequestInput struct {
	PullRequestID string  `json:"pull_request_id"`
	Title         *string `json:"title,omitempty"`
	Body          *string `json:"body,omitempty"`
	BaseRefName   *string `json:"base_ref_name,omitempty"`
}

// RequestReviewsInput defines parameters for requesting reviews
type RequestReviewsInput struct {
	PullRequestID string   `json:"pull_request_id"`
	UserIDs       []string `json:"user_ids,omitempty"`
	TeamIDs       []string `json:"team_ids,omitempty"`
	Union         *bool    `json:"union,omitempty"` // Add reviewers vs replace them
}

// RepositoryMergeRequestsInput defines parameters for fetching repository merge requests
type RepositoryMergeRequestsInput struct {
	Owner  string              `json:"owner"`
	Repo   string              `json:"repo"`
	First  int32               `json:"first,omitempty"`
	After  string              `json:"after,omitempty"`
	States []MergeRequestState `json:"states,omitempty"`
}

// RepositoryMergeRequestsPage represents a page of repository merge requests
type RepositoryMergeRequestsPage struct {
	PageInfo      PageInfo       `json:"page_info"`
	TotalCount    int64          `json:"total_count"`
	MergeRequests []MergeRequest `json:"merge_requests"`
}