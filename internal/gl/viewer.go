package gl

import "context"

type Viewer struct {
	Name     string `graphql:"name"`
	Username string `graphql:"username"`
}

func (c *Client) Viewer(ctx context.Context) (*Viewer, error) {
	var query struct {
		CurrentUser struct {
			Name     string `graphql:"name"`
			Username string `graphql:"username"`
		} `graphql:"currentUser"`
	}
	err := c.query(ctx, &query, nil)
	if err != nil {
		return nil, err
	}
	return &Viewer{
		Name:     query.CurrentUser.Name,
		Username: query.CurrentUser.Username,
	}, nil
}