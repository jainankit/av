package gitlab

import "context"

type Viewer struct {
	ID       string `graphql:"id"`
	Username string `graphql:"username"`
	Name     string `graphql:"name"`
	Email    string `graphql:"email"`
	WebURL   string `graphql:"webUrl"`
}

func (c *Client) Viewer(ctx context.Context) (*Viewer, error) {
	var query struct {
		CurrentUser Viewer `graphql:"currentUser"`
	}
	err := c.query(ctx, &query, nil)
	if err != nil {
		return nil, err
	}
	return &query.CurrentUser, nil
}