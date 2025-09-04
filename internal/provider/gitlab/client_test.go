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

// Merge request parsing tests are now in mergerequest_test.go