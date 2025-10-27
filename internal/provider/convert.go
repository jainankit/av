package provider

import (
	"github.com/shurcooL/githubv4"
)

// ConvertGitHubPullRequestState converts a GitHub pull request state to the provider-agnostic state.
func ConvertGitHubPullRequestState(state githubv4.PullRequestState) PullRequestState {
	switch state {
	case githubv4.PullRequestStateOpen:
		return PullRequestStateOpen
	case githubv4.PullRequestStateClosed:
		return PullRequestStateClosed
	case githubv4.PullRequestStateMerged:
		return PullRequestStateMerged
	default:
		// Default to closed for unknown states
		return PullRequestStateClosed
	}
}

// ConvertToGitHubPullRequestState converts a provider-agnostic state to GitHub pull request state.
func ConvertToGitHubPullRequestState(state PullRequestState) githubv4.PullRequestState {
	switch state {
	case PullRequestStateOpen:
		return githubv4.PullRequestStateOpen
	case PullRequestStateClosed:
		return githubv4.PullRequestStateClosed
	case PullRequestStateMerged:
		return githubv4.PullRequestStateMerged
	default:
		// Default to closed for unknown states
		return githubv4.PullRequestStateClosed
	}
}

// ConvertToGitHubPullRequestStates converts a slice of provider-agnostic states to GitHub states.
func ConvertToGitHubPullRequestStates(states []PullRequestState) []githubv4.PullRequestState {
	if len(states) == 0 {
		return nil
	}
	result := make([]githubv4.PullRequestState, len(states))
	for i, state := range states {
		result[i] = ConvertToGitHubPullRequestState(state)
	}
	return result
}

// ParsePullRequestState parses a string representation of a pull request state.
// This supports both provider-agnostic format and GitHub's format for backwards compatibility.
func ParsePullRequestState(s string) PullRequestState {
	switch s {
	case "OPEN", "open":
		return PullRequestStateOpen
	case "CLOSED", "closed":
		return PullRequestStateClosed
	case "MERGED", "merged":
		return PullRequestStateMerged
	default:
		// Try to handle GitHub-specific state strings for backwards compatibility
		return ConvertGitHubPullRequestState(githubv4.PullRequestState(s))
	}
}

// ConvertGitLabMergeRequestState converts a GitLab merge request state to the provider-agnostic state.
// GitLab states: "opened", "closed", "merged", "locked"
func ConvertGitLabMergeRequestState(state string) PullRequestState {
	switch state {
	case "opened":
		return PullRequestStateOpen
	case "closed":
		return PullRequestStateClosed
	case "merged":
		return PullRequestStateMerged
	case "locked":
		// Locked MRs are still open but locked for discussion
		return PullRequestStateOpen
	default:
		return PullRequestStateClosed
	}
}

// ConvertToGitLabMergeRequestState converts a provider-agnostic state to GitLab merge request state.
func ConvertToGitLabMergeRequestState(state PullRequestState) string {
	switch state {
	case PullRequestStateOpen:
		return "opened"
	case PullRequestStateClosed:
		return "closed"
	case PullRequestStateMerged:
		return "merged"
	default:
		return "closed"
	}
}
