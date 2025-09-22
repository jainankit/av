package gitlab

import (
	"context"

	"emperror.dev/errors"
)

type Group struct {
	ID          string `graphql:"id"`
	Name        string `graphql:"name"`
	Path        string `graphql:"path"`
	FullPath    string `graphql:"fullPath"`
	Description string `graphql:"description"`
	WebURL      string `graphql:"webUrl"`
	Visibility  string `graphql:"visibility"`
}

type GroupMember struct {
	ID          string `graphql:"id"`
	Username    string `graphql:"username"`
	Name        string `graphql:"name"`
	AccessLevel struct {
		IntegerValue int    `graphql:"integerValue"`
		StringValue  string `graphql:"stringValue"`
	} `graphql:"accessLevel"`
}

// Group returns information about the given group by its full path.
func (c *Client) Group(ctx context.Context, fullPath string) (*Group, error) {
	var query struct {
		Group Group `graphql:"group(fullPath: $fullPath)"`
	}
	if err := c.query(ctx, &query, map[string]any{
		"fullPath": fullPath,
	}); err != nil {
		return nil, err
	}
	if query.Group.ID == "" {
		return nil, errors.Errorf("GitLab group %q not found", fullPath)
	}
	return &query.Group, nil
}

// GroupByID returns information about the group with the given ID.
func (c *Client) GroupByID(ctx context.Context, groupID string) (*Group, error) {
	var query struct {
		Group Group `graphql:"group(id: $id)"`
	}
	if err := c.query(ctx, &query, map[string]any{
		"id": groupID,
	}); err != nil {
		return nil, err
	}
	if query.Group.ID == "" {
		return nil, errors.Errorf("GitLab group with ID %q not found", groupID)
	}
	return &query.Group, nil
}

// GroupMembers returns a list of members for the given group.
func (c *Client) GroupMembers(ctx context.Context, fullPath string) ([]GroupMember, error) {
	var query struct {
		Group struct {
			ID      string `graphql:"id"`
			Members struct {
				Nodes []GroupMember `graphql:"nodes"`
			} `graphql:"groupMembers"`
		} `graphql:"group(fullPath: $fullPath)"`
	}
	if err := c.query(ctx, &query, map[string]any{
		"fullPath": fullPath,
	}); err != nil {
		return nil, err
	}
	if query.Group.ID == "" {
		return nil, errors.Errorf("GitLab group %q not found", fullPath)
	}
	return query.Group.Members.Nodes, nil
}

// IsGroupMember checks if a user is a member of the specified group.
func (c *Client) IsGroupMember(ctx context.Context, fullPath, username string) (bool, error) {
	members, err := c.GroupMembers(ctx, fullPath)
	if err != nil {
		return false, err
	}
	
	for _, member := range members {
		if member.Username == username {
			return true, nil
		}
	}
	return false, nil
}