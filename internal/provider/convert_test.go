package provider

import (
	"testing"

	"github.com/shurcooL/githubv4"
)

func TestConvertGitHubPullRequestState(t *testing.T) {
	tests := []struct {
		name     string
		input    githubv4.PullRequestState
		expected PullRequestState
	}{
		{
			name:     "open",
			input:    githubv4.PullRequestStateOpen,
			expected: PullRequestStateOpen,
		},
		{
			name:     "closed",
			input:    githubv4.PullRequestStateClosed,
			expected: PullRequestStateClosed,
		},
		{
			name:     "merged",
			input:    githubv4.PullRequestStateMerged,
			expected: PullRequestStateMerged,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertGitHubPullRequestState(tt.input)
			if result != tt.expected {
				t.Errorf("ConvertGitHubPullRequestState(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConvertToGitHubPullRequestState(t *testing.T) {
	tests := []struct {
		name     string
		input    PullRequestState
		expected githubv4.PullRequestState
	}{
		{
			name:     "open",
			input:    PullRequestStateOpen,
			expected: githubv4.PullRequestStateOpen,
		},
		{
			name:     "closed",
			input:    PullRequestStateClosed,
			expected: githubv4.PullRequestStateClosed,
		},
		{
			name:     "merged",
			input:    PullRequestStateMerged,
			expected: githubv4.PullRequestStateMerged,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertToGitHubPullRequestState(tt.input)
			if result != tt.expected {
				t.Errorf("ConvertToGitHubPullRequestState(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConvertToGitHubPullRequestStates(t *testing.T) {
	input := []PullRequestState{
		PullRequestStateOpen,
		PullRequestStateClosed,
		PullRequestStateMerged,
	}
	expected := []githubv4.PullRequestState{
		githubv4.PullRequestStateOpen,
		githubv4.PullRequestStateClosed,
		githubv4.PullRequestStateMerged,
	}

	result := ConvertToGitHubPullRequestStates(input)
	if len(result) != len(expected) {
		t.Fatalf("ConvertToGitHubPullRequestStates() returned %d items, want %d", len(result), len(expected))
	}

	for i := range result {
		if result[i] != expected[i] {
			t.Errorf("ConvertToGitHubPullRequestStates()[%d] = %v, want %v", i, result[i], expected[i])
		}
	}

	// Test with empty slice
	emptyResult := ConvertToGitHubPullRequestStates(nil)
	if emptyResult != nil {
		t.Errorf("ConvertToGitHubPullRequestStates(nil) = %v, want nil", emptyResult)
	}
}

func TestParsePullRequestState(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected PullRequestState
	}{
		{
			name:     "uppercase open",
			input:    "OPEN",
			expected: PullRequestStateOpen,
		},
		{
			name:     "lowercase open",
			input:    "open",
			expected: PullRequestStateOpen,
		},
		{
			name:     "uppercase closed",
			input:    "CLOSED",
			expected: PullRequestStateClosed,
		},
		{
			name:     "lowercase closed",
			input:    "closed",
			expected: PullRequestStateClosed,
		},
		{
			name:     "uppercase merged",
			input:    "MERGED",
			expected: PullRequestStateMerged,
		},
		{
			name:     "lowercase merged",
			input:    "merged",
			expected: PullRequestStateMerged,
		},
		{
			name:     "unknown defaults to closed",
			input:    "unknown",
			expected: PullRequestStateClosed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParsePullRequestState(tt.input)
			if result != tt.expected {
				t.Errorf("ParsePullRequestState(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConvertGitLabMergeRequestState(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected PullRequestState
	}{
		{
			name:     "opened",
			input:    "opened",
			expected: PullRequestStateOpen,
		},
		{
			name:     "closed",
			input:    "closed",
			expected: PullRequestStateClosed,
		},
		{
			name:     "merged",
			input:    "merged",
			expected: PullRequestStateMerged,
		},
		{
			name:     "locked",
			input:    "locked",
			expected: PullRequestStateOpen,
		},
		{
			name:     "unknown defaults to closed",
			input:    "unknown",
			expected: PullRequestStateClosed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertGitLabMergeRequestState(tt.input)
			if result != tt.expected {
				t.Errorf("ConvertGitLabMergeRequestState(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConvertToGitLabMergeRequestState(t *testing.T) {
	tests := []struct {
		name     string
		input    PullRequestState
		expected string
	}{
		{
			name:     "open",
			input:    PullRequestStateOpen,
			expected: "opened",
		},
		{
			name:     "closed",
			input:    PullRequestStateClosed,
			expected: "closed",
		},
		{
			name:     "merged",
			input:    PullRequestStateMerged,
			expected: "merged",
		},
		{
			name:     "unknown defaults to closed",
			input:    PullRequestState("unknown"),
			expected: "closed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertToGitLabMergeRequestState(tt.input)
			if result != tt.expected {
				t.Errorf("ConvertToGitLabMergeRequestState(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBidirectionalConversion(t *testing.T) {
	// Test that converting back and forth maintains the value
	states := []PullRequestState{
		PullRequestStateOpen,
		PullRequestStateClosed,
		PullRequestStateMerged,
	}

	for _, state := range states {
		t.Run(string(state), func(t *testing.T) {
			// GitHub conversion
			ghState := ConvertToGitHubPullRequestState(state)
			backToProvider := ConvertGitHubPullRequestState(ghState)
			if backToProvider != state {
				t.Errorf("GitHub bidirectional conversion failed: %v -> %v -> %v", state, ghState, backToProvider)
			}

			// GitLab conversion
			glState := ConvertToGitLabMergeRequestState(state)
			backToProvider2 := ConvertGitLabMergeRequestState(glState)
			if backToProvider2 != state {
				t.Errorf("GitLab bidirectional conversion failed: %v -> %v -> %v", state, glState, backToProvider2)
			}
		})
	}
}
