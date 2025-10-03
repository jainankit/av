package gitlab

import (
	"context"
	"fmt"

	"emperror.dev/errors"
)

// Client represents a GitLab API client
type Client struct {
	token   string
	baseURL string
}

// ViewerInfo contains basic information about the authenticated user
type ViewerInfo struct {
	Name     string
	Username string
	Email    string
}

// NewClient creates a new GitLab client with the given token
func NewClient(ctx context.Context, token string) (*Client, error) {
	return NewClientWithBaseURL(ctx, token, "")
}

// NewClientWithBaseURL creates a new GitLab client with a custom base URL
func NewClientWithBaseURL(
	ctx context.Context, 
	token, baseURL string,
) (*Client, error) {
	if token == "" {
		return nil, errors.New("GitLab token is required")
	}
	
	if baseURL == "" {
		baseURL = "https://gitlab.com"
	}
	
	client := &Client{
		token:   token,
		baseURL: baseURL,
	}
	
	return client, nil
}

// Viewer returns information about the authenticated user
// This is a stub implementation for authentication validation
func (c *Client) Viewer(ctx context.Context) (*ViewerInfo, error) {
	// TODO: Implement actual GitLab API call to get user info
	// For now, this is a stub that will be implemented in a later step
	
	if c.token == "" {
		return nil, ErrUnauthorized
	}
	
	// Basic token format validation
	if err := validateGitLabTokenFormat(c.token); err != nil {
		return nil, ErrUnauthorized
	}
	
	// Return placeholder info - in a real implementation, this would make an API call
	return &ViewerInfo{
		Name:     "GitLab User",
		Username: "user",
		Email:    "",
	}, nil
}

// validateGitLabTokenFormat performs basic GitLab token format validation
func validateGitLabTokenFormat(token string) error {
	if len(token) == 0 {
		return errors.New("token is empty")
	}
	
	// GitLab personal access tokens usually start with "glpat_" or are legacy format
	// Legacy tokens are usually 20-64 characters long
	if len(token) < 20 {
		return errors.New(
			"token appears to be too short for a valid GitLab token",
		)
	}
	
	return nil
}
