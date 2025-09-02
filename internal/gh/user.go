package gh

import (
	"context"

	"github.com/aviator-co/av/internal/provider/github"
)

// User maintains the same structure as before for backward compatibility
type User = github.User

// User returns information about the given user.
func (c *Client) User(ctx context.Context, login string) (*User, error) {
	return c.provider.UserGithub(ctx, login)
}