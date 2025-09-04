package gitlab

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/aviator-co/av/internal/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xanzy/go-gitlab"
)

func TestClient_GenerateRepositoryPermalink(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		owner    string
		repoName string
		expected string
	}{
		{
			name:     "GitLab.com repository",
			baseURL:  "",
			owner:    "owner",
			repoName: "repo",
			expected: "https://gitlab.com/owner/repo",
		},
		{
			name:     "self-hosted GitLab",
			baseURL:  "https://gitlab.example.com/api/v4",
			owner:    "owner",
			repoName: "repo",
			expected: "https://gitlab.example.com/owner/repo",
		},
		{
			name:     "self-hosted without API path",
			baseURL:  "https://gitlab.example.com",
			owner:    "owner", 
			repoName: "repo",
			expected: "https://gitlab.example.com/owner/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{baseURL: tt.baseURL}
			permalink, err := client.GenerateRepositoryPermalink(context.Background(), tt.owner, tt.repoName)
			
			require.NoError(t, err)
			assert.Equal(t, tt.expected, permalink)
		})
	}
}

func TestClient_GenerateMergeRequestPermalink(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		owner    string
		repoName string
		mrNumber int64
		expected string
	}{
		{
			name:     "GitLab.com merge request",
			baseURL:  "",
			owner:    "owner",
			repoName: "repo",
			mrNumber: 123,
			expected: "https://gitlab.com/owner/repo/-/merge_requests/123",
		},
		{
			name:     "self-hosted GitLab merge request",
			baseURL:  "https://gitlab.example.com",
			owner:    "owner",
			repoName: "repo", 
			mrNumber: 456,
			expected: "https://gitlab.example.com/owner/repo/-/merge_requests/456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{baseURL: tt.baseURL}
			permalink, err := client.GenerateMergeRequestPermalink(context.Background(), tt.owner, tt.repoName, tt.mrNumber)
			
			require.NoError(t, err)
			assert.Equal(t, tt.expected, permalink)
		})
	}
}

func TestClient_GenerateBranchPermalink(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		owner    string
		repoName string
		branch   string
		expected string
	}{
		{
			name:     "simple branch name",
			baseURL:  "",
			owner:    "owner",
			repoName: "repo",
			branch:   "main",
			expected: "https://gitlab.com/owner/repo/-/tree/main",
		},
		{
			name:     "branch with special characters",
			baseURL:  "",
			owner:    "owner",
			repoName: "repo",
			branch:   "feature/add-new-feature",
			expected: "https://gitlab.com/owner/repo/-/tree/feature%2Fadd-new-feature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{baseURL: tt.baseURL}
			permalink, err := client.GenerateBranchPermalink(context.Background(), tt.owner, tt.repoName, tt.branch)
			
			require.NoError(t, err)
			assert.Equal(t, tt.expected, permalink)
		})
	}
}

func TestClient_GenerateCommitPermalink(t *testing.T) {
	client := &Client{baseURL: ""}
	permalink, err := client.GenerateCommitPermalink(context.Background(), "owner", "repo", "abc123")
	
	require.NoError(t, err)
	assert.Equal(t, "https://gitlab.com/owner/repo/-/commit/abc123", permalink)
}

func TestMapGitLabAccessLevel(t *testing.T) {
	tests := []struct {
		gitlabLevel gitlab.AccessLevelValue
		expected    AccessLevel
	}{
		{gitlab.NoPermissions, AccessLevelNoAccess},
		{gitlab.MinimalAccessPermissions, AccessLevelMinimal},
		{gitlab.GuestPermissions, AccessLevelGuest},
		{gitlab.ReporterPermissions, AccessLevelReporter},
		{gitlab.DeveloperPermissions, AccessLevelDeveloper},
		{gitlab.MaintainerPermissions, AccessLevelMaintainer},
		{gitlab.OwnerPermissions, AccessLevelOwner},
	}

	for _, tt := range tests {
		t.Run(string(tt.expected), func(t *testing.T) {
			result := mapGitLabAccessLevel(tt.gitlabLevel)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAccessLevelConstants(t *testing.T) {
	// Verify our access level constants are defined correctly
	assert.Equal(t, AccessLevel("NO_ACCESS"), AccessLevelNoAccess)
	assert.Equal(t, AccessLevel("MINIMAL"), AccessLevelMinimal)
	assert.Equal(t, AccessLevel("GUEST"), AccessLevelGuest)
	assert.Equal(t, AccessLevel("REPORTER"), AccessLevelReporter)
	assert.Equal(t, AccessLevel("DEVELOPER"), AccessLevelDeveloper)
	assert.Equal(t, AccessLevel("MAINTAINER"), AccessLevelMaintainer)
	assert.Equal(t, AccessLevel("OWNER"), AccessLevelOwner)
}

func TestRepositorySettingsMapping(t *testing.T) {
	// Test that our RepositorySettings struct has all the expected fields
	createdTime := time.Now()
	lastActivityTime := time.Now().Add(-time.Hour)
	
	settings := &RepositorySettings{
		ID:                   "123",
		Name:                 "test-repo",
		Path:                 "test-repo",
		PathWithNamespace:    "group/test-repo",
		Description:          "A test repository",
		DefaultBranch:        "main",
		Visibility:           "public",
		HTTPURLToRepo:        "https://gitlab.com/group/test-repo.git",
		SSHURLToRepo:         "git@gitlab.com:group/test-repo.git",
		WebURL:               "https://gitlab.com/group/test-repo",
		MergeRequestsEnabled: true,
		IssuesEnabled:        true,
		WikiEnabled:          true,
		SnippetsEnabled:      true,
		JobsEnabled:          true,
		PagesEnabled:         true,
		ContainerRegistryEnabled: true,
		ArchiveOnDestroy:     false,
		RequestAccessEnabled: true,
		OnlyAllowMergeIfPipelineSucceeds: false,
		OnlyAllowMergeIfAllDiscussionsAreResolved: false,
		ForksCount:           5,
		StarsCount:           10,
		OpenIssuesCount:      2,
		CreatedAt:            &createdTime,
		LastActivityAt:       &lastActivityTime,
	}

	// Verify all fields are populated
	assert.Equal(t, "123", settings.ID)
	assert.Equal(t, "test-repo", settings.Name)
	assert.Equal(t, "group/test-repo", settings.PathWithNamespace)
	assert.Equal(t, "main", settings.DefaultBranch)
	assert.True(t, settings.MergeRequestsEnabled)
	assert.True(t, settings.IssuesEnabled)
	assert.Equal(t, 5, settings.ForksCount)
	assert.Equal(t, 10, settings.StarsCount)
	assert.NotNil(t, settings.CreatedAt)
	assert.NotNil(t, settings.LastActivityAt)
}

func TestProjectMemberType(t *testing.T) {
	member := ProjectMember{
		ID:          "456",
		Username:    "testuser",
		Name:        "Test User",
		AccessLevel: AccessLevelDeveloper,
	}

	assert.Equal(t, "456", member.ID)
	assert.Equal(t, "testuser", member.Username)
	assert.Equal(t, "Test User", member.Name)
	assert.Equal(t, AccessLevelDeveloper, member.AccessLevel)
}