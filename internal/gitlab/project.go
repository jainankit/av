package gitlab

import (
	"context"
	"strings"

	"emperror.dev/errors"
)

type Project struct {
	ID   string `graphql:"id"`
	Name string `graphql:"name"`
	Path string `graphql:"path"`
	Namespace struct {
		Path string `graphql:"path"`
	} `graphql:"namespace"`
	WebURL string `graphql:"webUrl"`
}

func (c *Client) GetProjectBySlug(ctx context.Context, slug string) (*Project, error) {
	namespace, project, ok := strings.Cut(slug, "/")
	if !ok {
		return nil, errors.Errorf(
			"unable to parse project slug (expected <namespace>/<project>): %q",
			slug,
		)
	}

	var query struct {
		Project Project `graphql:"project(fullPath: $fullPath)"`
	}
	err := c.query(ctx, &query, map[string]interface{}{
		"fullPath": slug,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch project from GitLab")
	}

	// Check if project was found (GitLab returns empty struct if not found)
	if query.Project.ID == "" {
		return nil, errors.Errorf("GitLab project %q not found", slug)
	}

	return &query.Project, nil
}