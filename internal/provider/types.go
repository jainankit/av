package provider

// ProviderType represents the type of git hosting provider.
type ProviderType string

const (
	// ProviderTypeGitHub represents GitHub (including GitHub Enterprise).
	ProviderTypeGitHub ProviderType = "github"
	// ProviderTypeGitLab represents GitLab (including self-hosted instances).
	ProviderTypeGitLab ProviderType = "gitlab"
)

// PRState represents the state of a pull request.
type PRState string

const (
	// PRStateOpen indicates the pull request is open.
	PRStateOpen PRState = "OPEN"
	// PRStateClosed indicates the pull request is closed without being merged.
	PRStateClosed PRState = "CLOSED"
	// PRStateMerged indicates the pull request has been merged.
	PRStateMerged PRState = "MERGED"
)

// PullRequest represents a pull request/merge request in a provider-agnostic way.
type PullRequest struct {
	// ID is the provider-specific identifier for the pull request.
	// For GitHub: GraphQL node ID (e.g., "PR_kwDOHMmHms5...")
	// For GitLab: Global MR ID as string (e.g., "12345")
	ID string

	// Number is the user-visible pull request number.
	// For GitHub: PR number (e.g., 123)
	// For GitLab: MR IID (internal ID, per-project)
	Number int64

	// Title is the title of the pull request.
	Title string

	// Body is the description/body text of the pull request.
	Body string

	// HeadRefName is the name of the source branch (without "refs/heads/" prefix).
	HeadRefName string

	// BaseRefName is the name of the target branch (without "refs/heads/" prefix).
	BaseRefName string

	// State is the current state of the pull request.
	State PRState

	// IsDraft indicates whether the pull request is in draft status.
	IsDraft bool

	// Permalink is the web URL for the pull request.
	Permalink string

	// MergeCommitSHA is the SHA of the merge commit, if the PR has been merged.
	// Empty string if not merged or if merge commit is not available.
	MergeCommitSHA string
}

// Repository represents a git repository in a provider-agnostic way.
type Repository struct {
	// ID is the provider-specific identifier for the repository.
	// For GitHub: GraphQL node ID (e.g., "R_kgDOHMmHmg")
	// For GitLab: Project ID as string (e.g., "12345")
	ID string

	// Owner is the owner/namespace of the repository.
	// For GitHub: Organization or user login (e.g., "aviator-co")
	// For GitLab: Top-level namespace (e.g., "aviator-co")
	Owner string

	// Name is the name of the repository.
	// For GitHub: Repository name (e.g., "av")
	// For GitLab: Project name (e.g., "av")
	Name string
}

// User represents a user in a provider-agnostic way.
type User struct {
	// ID is the provider-specific identifier for the user.
	ID string

	// Login is the username/login of the user.
	Login string
}

// Team represents a team within an organization in a provider-agnostic way.
// Note: Not all providers support teams (e.g., GitLab does not have team reviewers).
type Team struct {
	// ID is the provider-specific identifier for the team.
	ID string

	// Name is the display name of the team.
	Name string

	// Slug is the URL-friendly identifier for the team.
	Slug string
}

// Viewer represents the currently authenticated user.
type Viewer struct {
	// Name is the display name of the user.
	Name string

	// Login is the username/login of the user.
	Login string
}

// PageInfo contains pagination information for paginated API responses.
type PageInfo struct {
	// HasNextPage indicates whether there are more results available.
	HasNextPage bool

	// HasPreviousPage indicates whether there are previous results available.
	HasPreviousPage bool

	// StartCursor is the cursor for the first item in the current page.
	StartCursor string

	// EndCursor is the cursor for the last item in the current page.
	EndCursor string
}

// CreatePRInput contains the parameters for creating a pull request.
type CreatePRInput struct {
	// RepositoryID is the provider-specific identifier for the repository.
	RepositoryID string

	// Title is the title of the pull request.
	Title string

	// Body is the description/body text of the pull request.
	Body string

	// HeadRefName is the name of the source branch.
	HeadRefName string

	// BaseRefName is the name of the target branch.
	BaseRefName string

	// Draft indicates whether to create the PR as a draft.
	Draft bool
}

// UpdatePRInput contains the parameters for updating a pull request.
type UpdatePRInput struct {
	// ID is the provider-specific identifier for the pull request.
	ID string

	// Title is the new title (nil means no change).
	Title *string

	// Body is the new body text (nil means no change).
	Body *string

	// BaseRefName is the new target branch (nil means no change).
	BaseRefName *string
}

// GetPRsInput contains the parameters for querying pull requests.
type GetPRsInput struct {
	// Owner is the owner/namespace of the repository.
	Owner string

	// Repo is the name of the repository.
	Repo string

	// HeadRefName filters by source branch name (optional).
	HeadRefName string

	// BaseRefName filters by target branch name (optional).
	BaseRefName string

	// States filters by pull request states (optional, empty means all states).
	States []PRState

	// First is the maximum number of results to return (pagination limit).
	First int32

	// After is the cursor for pagination (optional).
	After string
}

// GetPRsPage represents a page of pull request results.
type GetPRsPage struct {
	// PageInfo contains pagination information.
	PageInfo PageInfo

	// PullRequests is the list of pull requests in this page.
	PullRequests []PullRequest
}
