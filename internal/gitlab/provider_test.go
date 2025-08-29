package gitlab

import (
	"testing"

	"github.com/aviator-co/av/internal/provider"
)

func TestGitLabProvider_ImplementsInterface(t *testing.T) {
	// This test ensures GitLabProvider implements the Provider interface
	var _ provider.Provider = (*GitLabProvider)(nil)
}

func TestConvertStateFromGitLab(t *testing.T) {
	p := &GitLabProvider{}

	tests := []struct {
		name         string
		gitlabState  MergeRequestState
		expectedState provider.MRState
	}{
		{
			name:         "opened state",
			gitlabState:  MergeRequestStateOpened,
			expectedState: provider.MRStateOpen,
		},
		{
			name:         "closed state",
			gitlabState:  MergeRequestStateClosed,
			expectedState: provider.MRStateClosed,
		},
		{
			name:         "merged state",
			gitlabState:  MergeRequestStateMerged,
			expectedState: provider.MRStateMerged,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.convertStateFromGitLab(tt.gitlabState)
			if result != tt.expectedState {
				t.Errorf("convertStateFromGitLab() = %v, want %v", result, tt.expectedState)
			}
		})
	}
}

func TestConvertStateToGitLab(t *testing.T) {
	p := &GitLabProvider{}

	tests := []struct {
		name           string
		unifiedState   provider.MRState
		expectedState  MergeRequestState
	}{
		{
			name:          "open state",
			unifiedState:  provider.MRStateOpen,
			expectedState: MergeRequestStateOpened,
		},
		{
			name:          "closed state",
			unifiedState:  provider.MRStateClosed,
			expectedState: MergeRequestStateClosed,
		},
		{
			name:          "merged state",
			unifiedState:  provider.MRStateMerged,
			expectedState: MergeRequestStateMerged,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.convertStateToGitLab(tt.unifiedState)
			if result != tt.expectedState {
				t.Errorf("convertStateToGitLab() = %v, want %v", result, tt.expectedState)
			}
		})
	}
}

func TestConvertMergeRequest(t *testing.T) {
	p := &GitLabProvider{}

	gitlabMR := &MergeRequest{
		ID:           123,
		IID:          45,
		ProjectID:    67,
		Title:        "Test MR",
		Description:  "Test description",
		State:        string(MergeRequestStateOpened),
		SourceBranch: "feature-branch",
		TargetBranch: "main",
		WebURL:       "https://gitlab.com/owner/repo/-/merge_requests/45",
		DraftStatus:  false,
	}

	projectID := "67"
	result := p.convertMergeRequest(gitlabMR, projectID)

	if result.ID != "45" {
		t.Errorf("expected ID to be '45', got %s", result.ID)
	}
	if result.Number != 45 {
		t.Errorf("expected Number to be 45, got %d", result.Number)
	}
	if result.ProjectID != projectID {
		t.Errorf("expected ProjectID to be %s, got %s", projectID, result.ProjectID)
	}
	if result.Title != "Test MR" {
		t.Errorf("expected Title to be 'Test MR', got %s", result.Title)
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
}