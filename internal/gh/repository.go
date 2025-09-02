package gh

import (
	"context"

	"github.com/aviator-co/av/internal/provider/github"
)

// Repository maintains the same structure as before for backward compatibility
type Repository = github.Repository

// GetRepositoryBySlug retrieves a repository by its owner/name slug
func (c *Client) GetRepositoryBySlug(ctx context.Context, slug string) (*Repository, error) {
	return c.provider.GetRepositoryBySlugGithub(ctx, slug)
}