package gl

import (
	"context"

	"emperror.dev/errors"
)

type User struct {
	ID       string `graphql:"id"`
	Username string `graphql:"username"`
	Name     string `graphql:"name"`
	Email    string `graphql:"publicEmail"`
}

// User returns information about the given user.
func (c *Client) User(ctx context.Context, username string) (*User, error) {
	var query struct {
		User User `graphql:"user(username: $username)"`
	}
	if err := c.query(ctx, &query, map[string]any{
		"username": username,
	}); err != nil {
		return nil, err
	}
	if query.User.ID == "" {
		return nil, errors.Errorf("GitLab user %q not found", username)
	}
	return &query.User, nil
}

// CurrentUser returns information about the currently authenticated user.
func (c *Client) CurrentUser(ctx context.Context) (*User, error) {
	var query struct {
		CurrentUser User `graphql:"currentUser"`
	}
	if err := c.query(ctx, &query, nil); err != nil {
		return nil, errors.Wrap(err, "failed to get current user")
	}
	if query.CurrentUser.ID == "" {
		return nil, errors.New("no authenticated user found")
	}
	return &query.CurrentUser, nil
}