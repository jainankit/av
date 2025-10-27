package provider

import (
	"strings"
)

// PullRequestState represents the state of a pull request.
type PullRequestState string

const (
	// PullRequestStateOpen indicates an open pull request.
	PullRequestStateOpen PullRequestState = "OPEN"
	// PullRequestStateClosed indicates a closed (not merged) pull request.
	PullRequestStateClosed PullRequestState = "CLOSED"
	// PullRequestStateMerged indicates a merged pull request.
	PullRequestStateMerged PullRequestState = "MERGED"
)

// PullRequest represents a provider-agnostic pull request.
type PullRequest struct {
	// ID is the provider-specific identifier for the pull request.
	ID string
	// Number is the pull request number (typically displayed in the UI).
	Number int64
	// Title is the title of the pull request.
	Title string
	// Body is the description/body text of the pull request.
	Body string
	// State is the current state of the pull request (open, closed, merged).
	State PullRequestState
	// HeadRefName is the name of the branch containing changes.
	// May include "refs/heads/" prefix depending on the provider.
	HeadRefName string
	// BaseRefName is the name of the branch to merge into.
	// May include "refs/heads/" prefix depending on the provider.
	BaseRefName string
	// Permalink is the web URL to view the pull request.
	Permalink string
	// IsDraft indicates whether the pull request is in draft status.
	IsDraft bool
	// MergeCommitOID is the commit hash of the merge commit (if merged).
	MergeCommitOID string
}

// HeadBranchName returns the head branch name without any "refs/heads/" prefix.
func (p *PullRequest) HeadBranchName() string {
	return strings.TrimPrefix(p.HeadRefName, "refs/heads/")
}

// BaseBranchName returns the base branch name without any "refs/heads/" prefix.
func (p *PullRequest) BaseBranchName() string {
	return strings.TrimPrefix(p.BaseRefName, "refs/heads/")
}

// Repository represents a provider-agnostic repository.
type Repository struct {
	// ID is the provider-specific identifier for the repository.
	ID string
	// Owner is the username or organization that owns the repository.
	Owner string
	// Name is the name of the repository.
	Name string
	// Provider is the type of provider hosting this repository.
	Provider ProviderType
}

// User represents a provider-agnostic user.
type User struct {
	// ID is the provider-specific identifier for the user.
	ID string
	// Login is the username/handle of the user.
	Login string
	// Name is the display name of the user (may be empty).
	Name string
}

// Team represents a provider-agnostic team or group.
type Team struct {
	// ID is the provider-specific identifier for the team.
	ID string
	// Name is the display name of the team.
	Name string
	// Slug is the URL-friendly identifier for the team.
	Slug string
}
