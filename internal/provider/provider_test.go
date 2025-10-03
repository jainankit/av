package provider

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseProviderType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ProviderType
		wantErr  bool
	}{
		{"empty", "", Unknown, false},
		{"github", "github", GitHub, false},
		{"gh", "gh", GitHub, false},
		{"GitHub", "GitHub", GitHub, false},
		{"gitlab", "gitlab", GitLab, false},
		{"gl", "gl", GitLab, false},
		{"GitLab", "GitLab", GitLab, false},
		{"invalid", "invalid", Unknown, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseProviderType(tt.input)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestDetectFromURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		repoSlug  string
		expected  ProviderType
		wantErr   bool
	}{
		{
			name:     "github https",
			url:      "https://github.com/owner/repo.git",
			repoSlug: "owner/repo",
			expected: GitHub,
			wantErr:  false,
		},
		{
			name:     "github ssh",
			url:      "git@github.com:owner/repo.git",
			repoSlug: "owner/repo",
			expected: GitHub,
			wantErr:  false,
		},
		{
			name:     "gitlab https",
			url:      "https://gitlab.com/namespace/project.git",
			repoSlug: "namespace/project",
			expected: GitLab,
			wantErr:  false,
		},
		{
			name:     "gitlab ssh",
			url:      "git@gitlab.com:namespace/project.git",
			repoSlug: "namespace/project",
			expected: GitLab,
			wantErr:  false,
		},
		{
			name:     "enterprise github",
			url:      "https://github.enterprise.com/owner/repo.git",
			repoSlug: "owner/repo",
			expected: GitHub,
			wantErr:  false,
		},
		{
			name:     "self-hosted gitlab",
			url:      "https://gitlab.example.com/namespace/project.git",
			repoSlug: "namespace/project",
			expected: GitLab,
			wantErr:  false,
		},
		{
			name:    "unsupported provider",
			url:     "https://bitbucket.org/owner/repo.git",
			wantErr: true,
		},
		{
			name:    "invalid url",
			url:     "not-a-url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := detectFromURL(tt.url, tt.repoSlug)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, provider.Type)
				assert.Equal(t, tt.repoSlug, provider.RepoSlug)
				assert.NotEmpty(t, provider.BaseURL)
			}
		})
	}
}

func TestIsGitHubURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{"github https", "https://github.com/owner/repo.git", true},
		{"github ssh", "git@github.com:owner/repo.git", true},
		{"enterprise github", "https://github.enterprise.com/owner/repo.git", true},
		{"gitlab https", "https://gitlab.com/namespace/project.git", false},
		{"gitlab ssh", "git@gitlab.com:namespace/project.git", false},
		{"invalid url", "not-a-url", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsGitHubURL(tt.url)
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
		{"gitlab https", "https://gitlab.com/namespace/project.git", true},
		{"gitlab ssh", "git@gitlab.com:namespace/project.git", true},
		{"self-hosted gitlab", "https://gitlab.example.com/namespace/project.git", true},
		{"github https", "https://github.com/owner/repo.git", false},
		{"github ssh", "git@github.com:owner/repo.git", false},
		{"invalid url", "not-a-url", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsGitLabURL(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseGitHubURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
		wantErr  bool
	}{
		{
			name:     "github https",
			url:      "https://github.com/owner/repo.git",
			expected: "owner/repo",
			wantErr:  false,
		},
		{
			name:     "github ssh",
			url:      "git@github.com:owner/repo.git",
			expected: "owner/repo",
			wantErr:  false,
		},
		{
			name:     "github without .git",
			url:      "https://github.com/owner/repo",
			expected: "owner/repo",
			wantErr:  false,
		},
		{
			name:    "gitlab url",
			url:     "https://gitlab.com/namespace/project.git",
			wantErr: true,
		},
		{
			name:    "invalid url",
			url:     "not-a-url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseGitHubURL(tt.url)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestProviderFlags_ValidateFlags(t *testing.T) {
	tests := []struct {
		name    string
		flags   ProviderFlags
		wantErr bool
	}{
		{
			name:    "empty flags",
			flags:   ProviderFlags{},
			wantErr: false,
		},
		{
			name: "valid github provider",
			flags: ProviderFlags{
				Provider:    "github",
				GitHubToken: "token",
			},
			wantErr: false,
		},
		{
			name: "valid gitlab provider",
			flags: ProviderFlags{
				Provider:    "gitlab",
				GitLabToken: "token",
			},
			wantErr: false,
		},
		{
			name: "invalid provider",
			flags: ProviderFlags{
				Provider: "invalid",
			},
			wantErr: true,
		},
		{
			name: "conflicting tokens",
			flags: ProviderFlags{
				GitHubToken: "token1",
				GitLabToken: "token2",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.flags.ValidateFlags()
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDetectionOptions(t *testing.T) {
	// Test default options
	options := DefaultDetectionOptions()
	assert.Equal(t, Unknown, options.ForceProvider)
	assert.True(t, options.AllowFallback)
	assert.False(t, options.RequireAuthentication)
	
	// Test flags conversion
	flags := ProviderFlags{
		Provider: "github",
	}
	options, err := flags.ToDetectionOptions()
	require.NoError(t, err)
	assert.Equal(t, GitHub, options.ForceProvider)
}

func TestValidateTokenFormat(t *testing.T) {
	tests := []struct {
		name     string
		provider Provider
		wantErr  bool
	}{
		{
			name: "valid github classic token",
			provider: Provider{
				Type:  GitHub,
				Token: "1234567890abcdef1234567890abcdef12345678",
			},
			wantErr: false,
		},
		{
			name: "valid github new token",
			provider: Provider{
				Type:  GitHub,
				Token: "ghp_abcdefghijklmnopqrstuvwxyz1234567890",
			},
			wantErr: false,
		},
		{
			name: "valid gitlab token",
			provider: Provider{
				Type:  GitLab,
				Token: "glpat_abcdefghijklmnopqrstuvwxyz",
			},
			wantErr: false,
		},
		{
			name: "invalid github token",
			provider: Provider{
				Type:  GitHub,
				Token: "invalid-token",
			},
			wantErr: true,
		},
		{
			name: "invalid gitlab token",
			provider: Provider{
				Type:  GitLab,
				Token: "invalid!@#$",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTokenFormat(&tt.provider)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateProviderComprehensive(t *testing.T) {
	ctx := context.Background()
	
	// Test nil provider
	result := ValidateProviderComprehensive(ctx, nil)
	assert.False(t, result.IsValid)
	assert.NotEmpty(t, result.Errors)
	
	// Test valid provider structure  
	provider := &Provider{
		Type:    GitHub,
		BaseURL: "https://github.com",
		Token:   "ghp_test_token_1234567890abcdef",
	}
	
	result = ValidateProviderComprehensive(ctx, provider)
	// Note: This will not be fully valid because actual API validation is not implemented
	assert.NotEmpty(t, result.AuthenticationStatus)
	assert.NotEmpty(t, result.ConnectivityStatus)
}
