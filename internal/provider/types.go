package provider

// ProviderType represents the type of Git hosting provider
type ProviderType string

const (
	// GitHub represents GitHub.com and GitHub Enterprise Server
	GitHub ProviderType = "github"
	// GitLab represents GitLab.com and self-hosted GitLab instances
	GitLab ProviderType = "gitlab"
)

// String returns the string representation of the provider type
func (p ProviderType) String() string {
	return string(p)
}

// PullRequestState represents the state of a pull request across providers
type PullRequestState string

const (
	// PullRequestStateOpen represents an open pull request
	PullRequestStateOpen PullRequestState = "OPEN"
	// PullRequestStateClosed represents a closed pull request
	PullRequestStateClosed PullRequestState = "CLOSED"
	// PullRequestStateMerged represents a merged pull request
	PullRequestStateMerged PullRequestState = "MERGED"
)

// PullRequest represents a pull request in a provider-agnostic way
type PullRequest struct {
	ID          string
	Number      int64
	HeadRefName string
	BaseRefName string
	IsDraft     bool
	Permalink   string
	State       PullRequestState
	Title       string
	Body        string
	MergeCommit string
}

// HeadBranchName returns the head branch name without any refs/heads/ prefix
func (p *PullRequest) HeadBranchName() string {
	return trimRefsPrefix(p.HeadRefName)
}

// BaseBranchName returns the base branch name without any refs/heads/ prefix
func (p *PullRequest) BaseBranchName() string {
	return trimRefsPrefix(p.BaseRefName)
}

// GetMergeCommit returns the merge commit SHA if available
func (p *PullRequest) GetMergeCommit() string {
	return p.MergeCommit
}

// Repository represents a repository in a provider-agnostic way
type Repository struct {
	ID    string
	Owner string
	Name  string
}

// User represents a user in a provider-agnostic way
type User struct {
	ID    string
	Login string
}

// Viewer represents the authenticated user in a provider-agnostic way
type Viewer struct {
	Name  string
	Login string
}

// trimRefsPrefix removes the "refs/heads/" prefix from branch names if present
func trimRefsPrefix(refName string) string {
	const prefix = "refs/heads/"
	if len(refName) > len(prefix) && refName[:len(prefix)] == prefix {
		return refName[len(prefix):]
	}
	return refName
}