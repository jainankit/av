package gl

import (
	"context"
	"strings"

	"emperror.dev/errors"
)

type Repository struct {
	ID                string
	Name              string
	Path              string
	NameWithNamespace string
	Namespace         struct {
		Name string
		Path string
	}
}

func (c *Client) GetRepositoryBySlug(ctx context.Context, slug string) (*Repository, error) {
	// GitLab uses full path format like "group/subgroup/project" or "user/project"
	if slug == "" {
		return nil, errors.Errorf("repository slug cannot be empty")
	}

	var query struct {
		Project struct {
			ID                string `graphql:"id"`
			Name              string `graphql:"name"`
			Path              string `graphql:"path"`
			NameWithNamespace string `graphql:"nameWithNamespace"`
			Namespace         struct {
				Name string `graphql:"name"`
				Path string `graphql:"path"`
			} `graphql:"namespace"`
		} `graphql:"project(fullPath: $fullPath)"`
	}
	
	err := c.query(ctx, &query, map[string]interface{}{
		"fullPath": slug,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch project from GitLab")
	}

	if query.Project.ID == "" {
		return nil, errors.Errorf("GitLab project %q not found", slug)
	}

	return &Repository{
		ID:                query.Project.ID,
		Name:              query.Project.Name,
		Path:              query.Project.Path,
		NameWithNamespace: query.Project.NameWithNamespace,
		Namespace:         query.Project.Namespace,
	}, nil
}

// GetRepositoryByOwnerAndName provides compatibility with GitHub-style owner/repo format
func (c *Client) GetRepositoryByOwnerAndName(ctx context.Context, owner, name string) (*Repository, error) {
	slug := owner + "/" + name
	return c.GetRepositoryBySlug(ctx, slug)
}

// ParseRepositorySlug parses a GitLab project path and returns the namespace and project name
func ParseRepositorySlug(slug string) (namespace, project string, ok bool) {
	// Find the last '/' to separate the project name from the namespace path
	lastSlash := strings.LastIndex(slug, "/")
	if lastSlash == -1 {
		return "", "", false
	}
	return slug[:lastSlash], slug[lastSlash+1:], true
}