package providers

import (
	"strings"
	"time"
)

// PullRequestState represents the state of a pull request across providers
type PullRequestState string

const (
	PullRequestStateClosed PullRequestState = "CLOSED"
	PullRequestStateMerged PullRequestState = "MERGED"
	PullRequestStateOpen   PullRequestState = "OPEN"
)

// PullRequest represents a pull request/merge request abstracted across providers
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
	CreatedAt   time.Time
	UpdatedAt   time.Time
	// Provider-specific merge commit information
	MergeCommitSHA *string
	// Timeline events for determining merge/close commits
	ClosedBySHA *string
	MergedBySHA *string
}

// HeadBranchName returns the head branch name with refs/heads/ prefix stripped
func (p *PullRequest) HeadBranchName() string {
	return strings.TrimPrefix(p.HeadRefName, "refs/heads/")
}

// BaseBranchName returns the base branch name with refs/heads/ prefix stripped
func (p *PullRequest) BaseBranchName() string {
	return strings.TrimPrefix(p.BaseRefName, "refs/heads/")
}

// Repository represents a repository abstracted across providers
type Repository struct {
	ID    string
	Owner RepositoryOwner
	Name  string
}

// RepositoryOwner represents the owner of a repository
type RepositoryOwner struct {
	Login string
}

// User represents a user abstracted across providers
type User struct {
	ID    string
	Login string
}

// Team represents a team/group abstracted across providers
type Team struct {
	ID   string
	Name string
	Slug string
}

// Viewer represents the currently authenticated user
type Viewer struct {
	Name  string
	Login string
}

// GetPullRequestsPage represents a paginated list of pull requests
type GetPullRequestsPage struct {
	PageInfo
	PullRequests []PullRequest
}

// RepoPullRequestsPage represents a paginated list of repository pull requests
type RepoPullRequestsPage struct {
	PageInfo
	PullRequests []PullRequest
}