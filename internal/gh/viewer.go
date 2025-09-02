package gh

import (
	"context"

	"github.com/aviator-co/av/internal/provider/github"
)

// Viewer maintains the same structure as before for backward compatibility
type Viewer = github.Viewer

// Viewer returns information about the authenticated user
func (c *Client) Viewer(ctx context.Context) (*Viewer, error) {
	return c.provider.ViewerGithub(ctx)
}