package gh

import (
	"context"

	"github.com/aviator-co/av/internal/provider/github"
)

// Client wraps the new GitHub provider implementation for backward compatibility
type Client struct {
	provider *github.Client
}

// NewClient creates a new GitHub client.
// It takes configuration from the global config.Av.GitHub variable.
func NewClient(ctx context.Context, token string) (*Client, error) {
	provider, err := github.NewClient(ctx, token)
	if err != nil {
		return nil, err
	}
	return &Client{provider: provider}, nil
}

// All methods now delegate to the provider implementation while maintaining the same API surface