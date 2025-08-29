package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectProviderFromURL(t *testing.T) {
	tests := []struct {
		name        string
		remoteURL   string
		expected    ProviderType
		expectError bool
	}{
		{
			name:        "GitHub HTTPS URL",
			remoteURL:   "https://github.com/owner/repo.git",
			expected:    GitHub,
			expectError: false,
		},
		{
			name:        "GitHub SSH URL",
			remoteURL:   "git@github.com:owner/repo.git",
			expected:    GitHub,
			expectError: false,
		},
		{
			name:        "GitLab.com HTTPS URL",
			remoteURL:   "https://gitlab.com/owner/repo.git",
			expected:    GitLab,
			expectError: false,
		},
		{
			name:        "GitLab.com SSH URL",
			remoteURL:   "git@gitlab.com:owner/repo.git",
			expected:    GitLab,
			expectError: false,
		},
		{
			name:        "Self-hosted GitLab URL",
			remoteURL:   "https://gitlab.example.com/owner/repo.git",
			expected:    GitLab,
			expectError: false,
		},
		{
			name:        "Unknown provider defaults to GitHub",
			remoteURL:   "https://bitbucket.org/owner/repo.git",
			expected:    GitHub,
			expectError: false,
		},
		{
			name:        "Empty URL returns error",
			remoteURL:   "",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DetectProviderFromURL(tt.remoteURL)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, ProviderType(""), result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestValidateProviderConfig(t *testing.T) {
	tests := []struct {
		name         string
		providerType ProviderType
		config       ProviderConfig
		expectError  bool
	}{
		{
			name:         "Valid GitHub config",
			providerType: GitHub,
			config: ProviderConfig{
				Type:  GitHub,
				Token: "ghp_1234567890abcdef1234567890abcdef12345678",
			},
			expectError: false,
		},
		{
			name:         "Valid GitHub classic token",
			providerType: GitHub,
			config: ProviderConfig{
				Type:  GitHub,
				Token: "1234567890abcdef1234567890abcdef12345678",
			},
			expectError: false,
		},
		{
			name:         "Valid GitLab config",
			providerType: GitLab,
			config: ProviderConfig{
				Type:  GitLab,
				Token: "glpat-1234567890abcdef1234",
			},
			expectError: false,
		},
		{
			name:         "Empty provider type",
			providerType: "",
			config: ProviderConfig{
				Token: "some-token",
			},
			expectError: true,
		},
		{
			name:         "Empty token",
			providerType: GitHub,
			config: ProviderConfig{
				Type: GitHub,
			},
			expectError: true,
		},
		{
			name:         "Invalid GitHub token format",
			providerType: GitHub,
			config: ProviderConfig{
				Type:  GitHub,
				Token: "invalid-token",
			},
			expectError: true,
		},
		{
			name:         "Invalid GitLab token format (too short)",
			providerType: GitLab,
			config: ProviderConfig{
				Type:  GitLab,
				Token: "short",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProviderConfig(tt.providerType, tt.config)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseGitURL(t *testing.T) {
	tests := []struct {
		name        string
		gitURL      string
		expectedURL string
		expectError bool
	}{
		{
			name:        "HTTPS URL",
			gitURL:      "https://github.com/owner/repo.git",
			expectedURL: "https://github.com/owner/repo.git",
			expectError: false,
		},
		{
			name:        "SSH URL",
			gitURL:      "git@github.com:owner/repo.git",
			expectedURL: "https://github.com/owner/repo.git",
			expectError: false,
		},
		{
			name:        "Invalid SSH format",
			gitURL:      "git@invalid",
			expectedURL: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseGitURL(tt.gitURL)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedURL, result.String())
			}
		})
	}
}

func TestTokenValidation(t *testing.T) {
	t.Run("GitHub token validation", func(t *testing.T) {
		validTokens := []string{
			"ghp_1234567890abcdef1234567890abcdef12345678",
			"gho_1234567890abcdef1234567890abcdef12345678", 
			"ghu_1234567890abcdef1234567890abcdef12345678",
			"ghs_1234567890abcdef1234567890abcdef12345678",
			"ghr_1234567890abcdef1234567890abcdef12345678",
			"1234567890abcdef1234567890abcdef12345678", // Classic 40-char hex
		}

		invalidTokens := []string{
			"",
			"invalid",
			"ghp_short",
			"1234567890abcdef1234567890abcdef1234567g", // Invalid hex character
		}

		for _, token := range validTokens {
			assert.True(t, isValidGitHubToken(token), "Token should be valid: %s", token)
		}

		for _, token := range invalidTokens {
			assert.False(t, isValidGitHubToken(token), "Token should be invalid: %s", token)
		}
	})

	t.Run("GitLab token validation", func(t *testing.T) {
		validTokens := []string{
			"glpat-1234567890abcdef1234",
			"12345678901234567890abcdef", // Long alphanumeric
			"token_with_underscores_1234567890",
			"token-with-hyphens-1234567890",
		}

		invalidTokens := []string{
			"",
			"short", // Too short
			"token@with@invalid@chars", // Invalid characters
			"token with spaces", // Invalid characters
		}

		for _, token := range validTokens {
			assert.True(t, isValidGitLabToken(token), "Token should be valid: %s", token)
		}

		for _, token := range invalidTokens {
			assert.False(t, isValidGitLabToken(token), "Token should be invalid: %s", token)
		}
	})
}