package provider

// MergeRequest represents a unified merge request/pull request
type MergeRequest struct {
	ID           string
	Number       int64
	ProjectID    string
	Title        string
	Description  string
	State        MRState
	IsDraft      bool
	SourceBranch string
	TargetBranch string
	WebURL       string
	MergeCommit  string
}

// Repository represents a unified repository
type Repository struct {
	ID       string
	Name     string
	FullName string
	Owner    string
	WebURL   string
}

// CreateMRInput represents input for creating a merge request/pull request
type CreateMRInput struct {
	ProjectID    string
	Title        string
	Description  string
	SourceBranch string
	TargetBranch string
	Draft        bool
}

// UpdateMRInput represents input for updating a merge request/pull request
type UpdateMRInput struct {
	ProjectID   string
	MRID        string
	Title       *string
	Description *string
	Draft       *bool
	StateEvent  *string // "close", "reopen"
}

// GetMRsInput represents input for querying merge requests/pull requests
type GetMRsInput struct {
	ProjectID    string
	States       []MRState
	SourceBranch string
	TargetBranch string
	Page         int
	PerPage      int
}

// GetMRsPage represents a page of merge requests/pull requests with pagination
type GetMRsPage struct {
	MergeRequests []MergeRequest
	NextPage      int
	HasNextPage   bool
	TotalPages    int
	TotalCount    int
}