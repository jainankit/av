package github

import (
	"context"

	"emperror.dev/errors"
	"github.com/shurcooL/githubv4"
)

// User represents a GitHub user with GitHub-specific details
type User = githubUser

// Viewer represents a GitHub viewer with GitHub-specific details
type Viewer = githubViewer

// UserGithub returns information about the given user using GitHub-specific types
func (c *Client) UserGithub(ctx context.Context, login string) (*User, error) {
	var query struct {
		User User `graphql:"user(login: $login)"`
	}
	if err := c.query(ctx, &query, map[string]any{
		"login": githubv4.String(login),
	}); err != nil {
		return nil, err
	}
	if query.User.ID == "" {
		return nil, errors.Errorf("GitHub user %q not found", login)
	}
	return &query.User, nil
}

// ViewerGithub returns information about the authenticated user using GitHub-specific types
func (c *Client) ViewerGithub(ctx context.Context) (*Viewer, error) {
	var query struct {
		Viewer Viewer `graphql:"viewer"`
	}
	err := c.query(ctx, &query, nil)
	if err != nil {
		return nil, err
	}
	return &query.Viewer, nil
}