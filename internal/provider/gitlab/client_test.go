package gitlab

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		baseURL     string
		expectError bool
	}{
		{
			name:        "no token provided",
			token:       "",
			baseURL:     "",
			expectError: true,
		},
		{
			name:        "valid token for GitLab.com",
			token:       "dummy-token",
			baseURL:     "",
			expectError: false,
		},
		{
			name:        "valid token with custom base URL",
			token:       "dummy-token",
			baseURL:     "https://gitlab.example.com",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(context.Background(), tt.token, tt.baseURL)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, tt.baseURL, client.baseURL)
				assert.Equal(t, tt.token, client.token)
				assert.NotNil(t, client.gitlab)
			}
		})
	}
}

func TestParseProjectID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{
			name:     "numeric project ID",
			input:    "123",
			expected: 123,
		},
		{
			name:     "project path",
			input:    "group/project",
			expected: "group/project",
		},
		{
			name:     "complex project path",
			input:    "group/subgroup/project",
			expected: "group/subgroup/project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseProjectID(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseMergeRequestID(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectProject interface{}
		expectIID     int
		expectError   bool
	}{
		{
			name:          "numeric project ID with IID",
			input:         "123:456",
			expectProject: 123,
			expectIID:     456,
			expectError:   false,
		},
		{
			name:          "project path with IID",
			input:         "group/project:789",
			expectProject: "group/project",
			expectIID:     789,
			expectError:   false,
		},
		{
			name:        "invalid format - no colon",
			input:       "123456",
			expectError: true,
		},
		{
			name:        "invalid format - non-numeric IID",
			input:       "123:abc",
			expectError: true,
		},
		{
			name:        "empty input",
			input:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project, iid, err := parseMergeRequestID(tt.input)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectProject, project)
				assert.Equal(t, tt.expectIID, iid)
			}
		})
	}
}