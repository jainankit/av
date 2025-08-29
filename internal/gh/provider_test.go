package gh

import (
	"testing"

	"github.com/aviator-co/av/internal/provider"
	"github.com/shurcooL/githubv4"
)

func TestGitHubProvider_ImplementsInterface(t *testing.T) {
	// This test ensures GitHubProvider implements the Provider interface
	var _ provider.Provider = (*GitHubProvider)(nil)
}

func TestConvertGitHubStateToMR(t *testing.T) {
	tests := []struct {
		name          string
		githubState   githubv4.PullRequestState
		expectedState provider.MRState
	}{
		{
			name:          "open state",
			githubState:   githubv4.PullRequestStateOpen,
			expectedState: provider.MRStateOpen,
		},
		{
			name:          "closed state",
			githubState:   githubv4.PullRequestStateClosed,
			expectedState: provider.MRStateClosed,
		},
		{
			name:          "merged state",
			githubState:   githubv4.PullRequestStateMerged,
			expectedState: provider.MRStateMerged,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertGitHubStateToMR(tt.githubState)
			if result != tt.expectedState {
				t.Errorf("convertGitHubStateToMR() = %v, want %v", result, tt.expectedState)
			}
		})
	}
}

func TestConvertMRStateToGitHub(t *testing.T) {
	tests := []struct {
		name          string
		unifiedState  provider.MRState
		expectedState githubv4.PullRequestState
	}{
		{
			name:          "open state",
			unifiedState:  provider.MRStateOpen,
			expectedState: githubv4.PullRequestStateOpen,
		},
		{
			name:          "closed state",
			unifiedState:  provider.MRStateClosed,
			expectedState: githubv4.PullRequestStateClosed,
		},
		{
			name:          "merged state",
			unifiedState:  provider.MRStateMerged,
			expectedState: githubv4.PullRequestStateMerged,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertMRStateToGitHub(tt.unifiedState)
			if result != tt.expectedState {
				t.Errorf("convertMRStateToGitHub() = %v, want %v", result, tt.expectedState)
			}
		})
	}
}

func TestParseProjectID(t *testing.T) {
	tests := []struct {
		name        string
		projectID   string
		expectedOwner string
		expectedRepo  string
		expectError bool
	}{
		{
			name:          "valid project ID",
			projectID:     "owner/repo",
			expectedOwner: "owner",
			expectedRepo:  "repo",
			expectError:   false,
		},
		{
			name:          "complex repo name",
			projectID:     "my-org/my-repo-name",
			expectedOwner: "my-org",
			expectedRepo:  "my-repo-name",
			expectError:   false,
		},
		{
			name:        "invalid format - no slash",
			projectID:   "invalidformat",
			expectError: true,
		},
		{
			name:        "invalid format - too many slashes",
			projectID:   "owner/repo/extra",
			expectError: true,
		},
		{
			name:        "empty string",
			projectID:   "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := parseProjectID(tt.projectID)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("parseProjectID() expected error, but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("parseProjectID() unexpected error: %v", err)
				return
			}
			
			if owner != tt.expectedOwner {
				t.Errorf("parseProjectID() owner = %v, want %v", owner, tt.expectedOwner)
			}
			if repo != tt.expectedRepo {
				t.Errorf("parseProjectID() repo = %v, want %v", repo, tt.expectedRepo)
			}
		})
	}
}

func TestConvertGitHubPRToMR(t *testing.T) {
	p := &GitHubProvider{}
	
	ghPR := &PullRequest{
		ID:          "PR_123",
		Number:      456,
		HeadRefName: "feature-branch",
		BaseRefName: "main",
		IsDraft:     false,
		Permalink:   "https://github.com/owner/repo/pull/456",
		State:       githubv4.PullRequestStateOpen,
		Title:       "Test PR",
		Body:        "Test description",
	}

	owner := "owner"
	repo := "repo"
	result := p.convertGitHubPRToMR(ghPR, owner, repo)

	if result.ID != "PR_123" {
		t.Errorf("expected ID to be 'PR_123', got %s", result.ID)
	}
	if result.Number != 456 {
		t.Errorf("expected Number to be 456, got %d", result.Number)
	}
	if result.ProjectID != "owner/repo" {
		t.Errorf("expected ProjectID to be 'owner/repo', got %s", result.ProjectID)
	}
	if result.Title != "Test PR" {
		t.Errorf("expected Title to be 'Test PR', got %s", result.Title)
	}
	if result.State != provider.MRStateOpen {
		t.Errorf("expected State to be %v, got %v", provider.MRStateOpen, result.State)
	}
	if result.IsDraft != false {
		t.Errorf("expected IsDraft to be false, got %t", result.IsDraft)
	}
	if result.SourceBranch != "feature-branch" {
		t.Errorf("expected SourceBranch to be 'feature-branch', got %s", result.SourceBranch)
	}
	if result.TargetBranch != "main" {
		t.Errorf("expected TargetBranch to be 'main', got %s", result.TargetBranch)
	}
	if result.WebURL != "https://github.com/owner/repo/pull/456" {
		t.Errorf("expected WebURL to be 'https://github.com/owner/repo/pull/456', got %s", result.WebURL)
	}
}