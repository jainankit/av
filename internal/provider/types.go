package provider

import "strings"

// MRState represents the state of a merge request across different providers.
type MRState string

const (
	// MRStateOpen represents an open merge request.
	MRStateOpen MRState = "open"
	// MRStateClosed represents a closed merge request.
	MRStateClosed MRState = "closed"
	// MRStateMerged represents a merged merge request.
	MRStateMerged MRState = "merged"
	// MRStateDraft represents a draft merge request.
	MRStateDraft MRState = "draft"
)

// MergeRequest represents a merge request/pull request in a unified format.
type MergeRequest struct {
	// ID is the unique identifier for the merge request.
	ID string
	// Number is the merge request number.
	Number int64
	// Title is the merge request title.
	Title string
	// Body is the merge request description/body.
	Body string
	// HeadRefName is the name of the head branch.
	HeadRefName string
	// BaseRefName is the name of the base branch.
	BaseRefName string
	// IsDraft indicates if the merge request is in draft status.
	IsDraft bool
	// State represents the current state of the merge request.
	State MRState
	// Permalink is the web URL to the merge request.
	Permalink string
	// MergeCommit is the SHA of the merge commit (if merged).
	MergeCommit string
}

// HeadBranchName returns the head branch name without any "refs/heads/" prefix.
func (mr *MergeRequest) HeadBranchName() string {
	return strings.TrimPrefix(mr.HeadRefName, "refs/heads/")
}

// BaseBranchName returns the base branch name without any "refs/heads/" prefix.
func (mr *MergeRequest) BaseBranchName() string {
	return strings.TrimPrefix(mr.BaseRefName, "refs/heads/")
}

// Repository represents repository information in a unified format.
type Repository struct {
	// ID is the unique identifier for the repository.
	ID string
	// Name is the repository name.
	Name string
	// Owner is the repository owner/namespace.
	Owner string
	// FullName is the full repository name in "owner/name" format.
	FullName string
}

// CreateMRInput represents the input for creating a merge request.
type CreateMRInput struct {
	// RepositoryID is the unique identifier of the repository.
	RepositoryID string
	// BaseRefName is the name of the base branch.
	BaseRefName string
	// HeadRefName is the name of the head branch.
	HeadRefName string
	// Title is the merge request title.
	Title string
	// Body is the merge request description/body.
	Body string
	// Draft indicates if the merge request should be created as a draft.
	Draft bool
	// MaintainerCanModify indicates if maintainers can modify the merge request.
	MaintainerCanModify bool
}

// UpdateMRInput represents the input for updating a merge request.
type UpdateMRInput struct {
	// ID is the unique identifier of the merge request to update.
	ID string
	// Title is the new title (optional).
	Title *string
	// Body is the new body/description (optional).
	Body *string
	// BaseRefName is the new base branch (optional).
	BaseRefName *string
	// State is the new state (optional).
	State *MRState
	// MaintainerCanModify indicates if maintainers can modify the merge request (optional).
	MaintainerCanModify *bool
}

// GetMRsInput represents the input for querying merge requests.
type GetMRsInput struct {
	// Owner is the repository owner.
	Owner string
	// Repo is the repository name.
	Repo string
	// HeadRefName filters by head branch name (optional).
	HeadRefName string
	// BaseRefName filters by base branch name (optional).
	BaseRefName string
	// States filters by merge request states (optional).
	States []MRState
	// First specifies the maximum number of results to return.
	First int32
	// After is the cursor for pagination (optional).
	After string
}

// PageInfo represents pagination information.
type PageInfo struct {
	// HasNextPage indicates if there are more pages available.
	HasNextPage bool
	// HasPreviousPage indicates if there are previous pages available.
	HasPreviousPage bool
	// StartCursor is the cursor pointing to the start of the current page.
	StartCursor string
	// EndCursor is the cursor pointing to the end of the current page.
	EndCursor string
}

// MergeRequestsPage represents a paginated response of merge requests.
type MergeRequestsPage struct {
	// PageInfo contains pagination information.
	PageInfo PageInfo
	// TotalCount is the total number of merge requests (optional, not all providers support this).
	TotalCount int64
	// MergeRequests contains the merge requests in this page.
	MergeRequests []MergeRequest
}