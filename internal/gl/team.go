package gl

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
}

// GroupByPath returns information about the given group by its path.
func (c *Client) GroupByPath(
	ctx context.Context,
	groupPath string,
) (*Group, error) {
	var query struct {
		Group Group `graphql:"group(fullPath: $fullPath)"`
	}
	if err := c.query(ctx, &query, map[string]any{
		"fullPath": groupPath,
	}); err != nil {
		return nil, err
	}
	if query.Group.ID == "" {
		return nil, errors.Errorf("GitLab group %q not found", groupPath)
	}
	return &query.Group, nil
}

// Note: GitLab doesn't have the exact same concept of "teams" as GitHub.
// GitLab uses Groups and Subgroups for organization, and users can have different
// access levels within groups. For compatibility, we're treating groups as teams.
type Team struct {
	ID          string `graphql:"id"`
	Name        string `graphql:"name"`
	Path        string `graphql:"path"`
	FullPath    string `graphql:"fullPath"`
	Description string `graphql:"description"`
}

// GroupTeam returns information about a subgroup within a parent group.
// This provides similar functionality to GitHub's organization teams.
func (c *Client) GroupTeam(
	ctx context.Context,
	parentGroupPath string,
	teamPath string,
) (*Team, error) {
	fullTeamPath := parentGroupPath + "/" + teamPath
	var query struct {
		Group struct {
			ID          string `graphql:"id"`
			Name        string `graphql:"name"`
			Path        string `graphql:"path"`
			FullPath    string `graphql:"fullPath"`
			Description string `graphql:"description"`
		} `graphql:"group(fullPath: $fullPath)"`
	}
	if err := c.query(ctx, &query, map[string]any{
		"fullPath": fullTeamPath,
	}); err != nil {
		return nil, err
	}
	if query.Group.ID == "" {
		return nil, errors.Errorf(
			"GitLab subgroup %q not found within group %q",
			teamPath,
			parentGroupPath,
		)
	}
	return &Team{
		ID:          query.Group.ID,
		Name:        query.Group.Name,
		Path:        query.Group.Path,
		FullPath:    query.Group.FullPath,
		Description: query.Group.Description,
	}, nil
}