package gitlab

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewURLBuilder(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		expected string
	}{
		{
			name:     "empty base URL defaults to GitLab.com",
			baseURL:  "",
			expected: "https://gitlab.com",
		},
		{
			name:     "base URL with trailing slash",
			baseURL:  "https://gitlab.example.com/",
			expected: "https://gitlab.example.com",
		},
		{
			name:     "base URL with /api/v4 suffix",
			baseURL:  "https://gitlab.example.com/api/v4",
			expected: "https://gitlab.example.com",
		},
		{
			name:     "base URL with both trailing slash and /api/v4",
			baseURL:  "https://gitlab.example.com/api/v4/",
			expected: "https://gitlab.example.com",
		},
		{
			name:     "clean base URL",
			baseURL:  "https://gitlab.example.com",
			expected: "https://gitlab.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewURLBuilder(tt.baseURL)
			assert.Equal(t, tt.expected, builder.baseURL)
		})
	}
}

func TestURLBuilder_ProjectURL(t *testing.T) {
	builder := NewURLBuilder("https://gitlab.example.com")
	url := builder.ProjectURL("owner", "repo")
	assert.Equal(t, "https://gitlab.example.com/owner/repo", url)
}

func TestURLBuilder_MergeRequestURL(t *testing.T) {
	builder := NewURLBuilder("https://gitlab.example.com")
	url := builder.MergeRequestURL("owner", "repo", 123)
	assert.Equal(t, "https://gitlab.example.com/owner/repo/-/merge_requests/123", url)
}

func TestURLBuilder_IssueURL(t *testing.T) {
	builder := NewURLBuilder("https://gitlab.example.com") 
	url := builder.IssueURL("owner", "repo", 456)
	assert.Equal(t, "https://gitlab.example.com/owner/repo/-/issues/456", url)
}

func TestURLBuilder_BranchURL(t *testing.T) {
	builder := NewURLBuilder("https://gitlab.example.com")
	
	tests := []struct {
		name     string
		branch   string
		expected string
	}{
		{
			name:     "simple branch name",
			branch:   "main",
			expected: "https://gitlab.example.com/owner/repo/-/tree/main",
		},
		{
			name:     "branch with slash",
			branch:   "feature/new-feature",
			expected: "https://gitlab.example.com/owner/repo/-/tree/feature%2Fnew-feature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := builder.BranchURL("owner", "repo", tt.branch)
			assert.Equal(t, tt.expected, url)
		})
	}
}

func TestURLBuilder_CommitURL(t *testing.T) {
	builder := NewURLBuilder("https://gitlab.example.com")
	url := builder.CommitURL("owner", "repo", "abc123")
	assert.Equal(t, "https://gitlab.example.com/owner/repo/-/commit/abc123", url)
}

func TestURLBuilder_CompareURL(t *testing.T) {
	builder := NewURLBuilder("https://gitlab.example.com")
	
	tests := []struct {
		name     string
		from     string
		to       string
		expected string
	}{
		{
			name:     "simple branch names",
			from:     "main",
			to:       "develop",
			expected: "https://gitlab.example.com/owner/repo/-/compare/main...develop",
		},
		{
			name:     "branches with slashes", 
			from:     "feature/branch-a",
			to:       "feature/branch-b",
			expected: "https://gitlab.example.com/owner/repo/-/compare/feature%2Fbranch-a...feature%2Fbranch-b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := builder.CompareURL("owner", "repo", tt.from, tt.to)
			assert.Equal(t, tt.expected, url)
		})
	}
}

func TestURLBuilder_FileURL(t *testing.T) {
	builder := NewURLBuilder("https://gitlab.example.com")
	
	tests := []struct {
		name     string
		branch   string
		filePath string
		expected string
	}{
		{
			name:     "file in root",
			branch:   "main",
			filePath: "README.md",
			expected: "https://gitlab.example.com/owner/repo/-/blob/main/README.md",
		},
		{
			name:     "file in subdirectory",
			branch:   "main",
			filePath: "src/main.go",
			expected: "https://gitlab.example.com/owner/repo/-/blob/main/src/main.go",
		},
		{
			name:     "file with leading slash",
			branch:   "main",
			filePath: "/src/main.go",
			expected: "https://gitlab.example.com/owner/repo/-/blob/main/src/main.go",
		},
		{
			name:     "branch with slash",
			branch:   "feature/branch",
			filePath: "file.txt",
			expected: "https://gitlab.example.com/owner/repo/-/blob/feature%2Fbranch/file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := builder.FileURL("owner", "repo", tt.branch, tt.filePath)
			assert.Equal(t, tt.expected, url)
		})
	}
}

func TestURLBuilder_UserURL(t *testing.T) {
	builder := NewURLBuilder("https://gitlab.example.com")
	url := builder.UserURL("username")
	assert.Equal(t, "https://gitlab.example.com/username", url)
}

func TestURLBuilder_GroupURL(t *testing.T) {
	builder := NewURLBuilder("https://gitlab.example.com")
	url := builder.GroupURL("group-name")
	assert.Equal(t, "https://gitlab.example.com/group-name", url)
}

func TestParseGitLabURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected *GitLabURLComponents
	}{
		{
			name: "project URL",
			url:  "https://gitlab.com/owner/repo",
			expected: &GitLabURLComponents{
				BaseURL: "https://gitlab.com",
				Owner:   "owner",
				Name:    "repo",
				Type:    "project",
			},
		},
		{
			name: "merge request URL",
			url:  "https://gitlab.com/owner/repo/-/merge_requests/123",
			expected: &GitLabURLComponents{
				BaseURL: "https://gitlab.com",
				Owner:   "owner",
				Name:    "repo",
				Type:    "merge_request",
				Number:  "123",
			},
		},
		{
			name: "issue URL",
			url:  "https://gitlab.com/owner/repo/-/issues/456",
			expected: &GitLabURLComponents{
				BaseURL: "https://gitlab.com",
				Owner:   "owner",
				Name:    "repo",
				Type:    "issue",
				Number:  "456",
			},
		},
		{
			name: "branch URL",
			url:  "https://gitlab.com/owner/repo/-/tree/main",
			expected: &GitLabURLComponents{
				BaseURL: "https://gitlab.com",
				Owner:   "owner",
				Name:    "repo",
				Type:    "branch",
				Branch:  "main",
			},
		},
		{
			name: "branch URL with slashes",
			url:  "https://gitlab.com/owner/repo/-/tree/feature/branch",
			expected: &GitLabURLComponents{
				BaseURL: "https://gitlab.com",
				Owner:   "owner",
				Name:    "repo",
				Type:    "branch",
				Branch:  "feature/branch",
			},
		},
		{
			name: "file URL",
			url:  "https://gitlab.com/owner/repo/-/blob/main/README.md",
			expected: &GitLabURLComponents{
				BaseURL:  "https://gitlab.com",
				Owner:    "owner",
				Name:     "repo",
				Type:     "file",
				Branch:   "main",
				FilePath: "README.md",
			},
		},
		{
			name: "commit URL",
			url:  "https://gitlab.com/owner/repo/-/commit/abc123",
			expected: &GitLabURLComponents{
				BaseURL: "https://gitlab.com",
				Owner:   "owner",
				Name:    "repo",
				Type:    "commit",
				SHA:     "abc123",
			},
		},
		{
			name: "compare URL",
			url:  "https://gitlab.com/owner/repo/-/compare/main...develop",
			expected: &GitLabURLComponents{
				BaseURL: "https://gitlab.com",
				Owner:   "owner",
				Name:    "repo",
				Type:    "compare",
				FromRef: "main",
				ToRef:   "develop",
			},
		},
		{
			name: "user URL",
			url:  "https://gitlab.com/username",
			expected: &GitLabURLComponents{
				BaseURL: "https://gitlab.com",
				Owner:   "",
				Name:    "username",
				Type:    "user_or_group",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseGitLabURL(tt.url)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsGitLabURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "GitLab.com URL",
			url:      "https://gitlab.com/owner/repo",
			expected: true,
		},
		{
			name:     "www.GitLab.com URL",
			url:      "https://www.gitlab.com/owner/repo",
			expected: true,
		},
		{
			name:     "self-hosted GitLab",
			url:      "https://gitlab.example.com/owner/repo",
			expected: true,
		},
		{
			name:     "GitHub URL",
			url:      "https://github.com/owner/repo",
			expected: false,
		},
		{
			name:     "other domain",
			url:      "https://example.com/owner/repo",
			expected: false,
		},
		{
			name:     "invalid URL",
			url:      "not-a-url",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsGitLabURL(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeGitLabBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		expected string
	}{
		{
			name:     "empty URL",
			baseURL:  "",
			expected: "",
		},
		{
			name:     "URL without /api/v4",
			baseURL:  "https://gitlab.example.com",
			expected: "https://gitlab.example.com/api/v4",
		},
		{
			name:     "URL with /api/v4",
			baseURL:  "https://gitlab.example.com/api/v4",
			expected: "https://gitlab.example.com/api/v4",
		},
		{
			name:     "URL with trailing slash",
			baseURL:  "https://gitlab.example.com/",
			expected: "https://gitlab.example.com/api/v4",
		},
		{
			name:     "URL with trailing slash and /api/v4",
			baseURL:  "https://gitlab.example.com/api/v4/",
			expected: "https://gitlab.example.com/api/v4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeGitLabBaseURL(tt.baseURL)
			assert.Equal(t, tt.expected, result)
		})
	}
}