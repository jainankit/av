package provider

import (
	"testing"
)

func TestPullRequestHelperMethods(t *testing.T) {
	pr := &PullRequest{
		HeadRefName: "refs/heads/feature-branch",
		BaseRefName: "refs/heads/main",
	}

	if pr.HeadBranchName() != "feature-branch" {
		t.Errorf("HeadBranchName() = %q, want %q", pr.HeadBranchName(), "feature-branch")
	}

	if pr.BaseBranchName() != "main" {
		t.Errorf("BaseBranchName() = %q, want %q", pr.BaseBranchName(), "main")
	}

	// Test without prefix
	pr2 := &PullRequest{
		HeadRefName: "feature",
		BaseRefName: "develop",
	}

	if pr2.HeadBranchName() != "feature" {
		t.Errorf("HeadBranchName() = %q, want %q", pr2.HeadBranchName(), "feature")
	}

	if pr2.BaseBranchName() != "develop" {
		t.Errorf("BaseBranchName() = %q, want %q", pr2.BaseBranchName(), "develop")
	}
}

func TestProviderType(t *testing.T) {
	if ProviderTypeGitHub != "github" {
		t.Errorf("ProviderTypeGitHub = %q, want %q", ProviderTypeGitHub, "github")
	}

	if ProviderTypeGitLab != "gitlab" {
		t.Errorf("ProviderTypeGitLab = %q, want %q", ProviderTypeGitLab, "gitlab")
	}
}

func TestPullRequestState(t *testing.T) {
	if PullRequestStateOpen != "OPEN" {
		t.Errorf("PullRequestStateOpen = %q, want %q", PullRequestStateOpen, "OPEN")
	}

	if PullRequestStateClosed != "CLOSED" {
		t.Errorf("PullRequestStateClosed = %q, want %q", PullRequestStateClosed, "CLOSED")
	}

	if PullRequestStateMerged != "MERGED" {
		t.Errorf("PullRequestStateMerged = %q, want %q", PullRequestStateMerged, "MERGED")
	}
}
