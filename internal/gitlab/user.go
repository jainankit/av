package gitlab

import (
	"context"

	"emperror.dev/errors"
)

type User struct {
	ID       string `graphql:"id"`
	Username string `graphql:"username"`
	Name     string `graphql:"name"`
	Email    string `graphql:"publicEmail"`
	WebURL   string `graphql:"webUrl"`
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

// UserByID returns information about the user with the given ID.
func (c *Client) UserByID(ctx context.Context, userID string) (*User, error) {
	var query struct {
		User User `graphql:"user(id: $id)"`
	}
	if err := c.query(ctx, &query, map[string]any{
		"id": userID,
	}); err != nil {
		return nil, err
	}
	if query.User.ID == "" {
		return nil, errors.Errorf("GitLab user with ID %q not found", userID)
	}
	return &query.User, nil
}

// IsAuthenticated checks if the client is properly authenticated by making a simple API call.
func (c *Client) IsAuthenticated(ctx context.Context) error {
	var query struct {
		CurrentUser struct {
			ID string `graphql:"id"`
		} `graphql:"currentUser"`
	}
	err := c.query(ctx, &query, nil)
	if err != nil {
		return errors.Wrap(err, "authentication check failed")
	}
	if query.CurrentUser.ID == "" {
		return errors.New("authentication check failed: no current user returned")
	}
	return nil
}