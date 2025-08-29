package gitlab

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"emperror.dev/errors"
)

// Repository represents a GitLab project/repository
type Repository struct {
	ID                int64  `json:"id"`
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"`
	WebURL            string `json:"web_url"`
	HTTPURLToRepo     string `json:"http_url_to_repo"`
	SSHURLToRepo      string `json:"ssh_url_to_repo"`
	Namespace         struct {
		Name string `json:"name"`
		Path string `json:"path"`
	} `json:"namespace"`
	DefaultBranch string `json:"default_branch"`
	Visibility    string `json:"visibility"`
}

// GetOwner returns the namespace (owner) of the repository
func (r *Repository) GetOwner() string {
	return r.Namespace.Path
}

// GetName returns the name of the repository  
func (r *Repository) GetName() string {
	return r.Name
}

// GetSlug returns the full repository slug (owner/repo)
func (r *Repository) GetSlug() string {
	return r.PathWithNamespace
}

// GetRepository retrieves a single repository/project by its ID or path
func (c *Client) GetRepository(ctx context.Context, projectID string) (*Repository, error) {
	endpoint := "/projects/" + url.PathEscape(projectID)
	
	var result Repository
	if err := c.get(ctx, endpoint, &result); err != nil {
		if IsHTTPNotFound(err) {
			return nil, errors.Errorf("project %s not found", projectID)
		}
		return nil, errors.Wrap(err, "failed to get repository")
	}

	return &result, nil
}

// GetRepositoryBySlug retrieves a repository by its slug (owner/repo format)
func (c *Client) GetRepositoryBySlug(ctx context.Context, slug string) (*Repository, error) {
	// Validate the slug format
	owner, name, ok := strings.Cut(slug, "/")
	if !ok {
		return nil, errors.Errorf(
			"unable to parse repository slug (expected <owner>/<repo>): %q",
			slug,
		)
	}
	
	// GitLab uses URL-encoded path format for API calls
	encodedSlug := url.PathEscape(slug)
	
	return c.GetRepository(ctx, encodedSlug)
}

// ResolveProjectID resolves a project ID from various inputs (ID, slug, etc.)
// Returns the project ID as a string that can be used in API calls
func (c *Client) ResolveProjectID(ctx context.Context, input string) (string, error) {
	// If input is a numeric string, it's likely already a project ID
	if _, err := strconv.ParseInt(input, 10, 64); err == nil {
		// Validate that this project ID exists
		_, err := c.GetRepository(ctx, input)
		if err != nil {
			return "", errors.Wrap(err, "project ID does not exist")
		}
		return input, nil
	}

	// Otherwise, treat it as a slug and resolve to project ID
	repo, err := c.GetRepositoryBySlug(ctx, input)
	if err != nil {
		return "", errors.Wrap(err, "failed to resolve project from slug")
	}

	return strconv.FormatInt(repo.ID, 10), nil
}

// ParseGitLabURL parses a GitLab URL and extracts the repository slug
// Supports various GitLab URL formats:
// - https://gitlab.com/owner/repo
// - https://gitlab.com/owner/repo.git
// - git@gitlab.com:owner/repo.git
// - https://custom-gitlab.example.com/owner/repo
func ParseGitLabURL(gitURL string) (baseURL, slug string, err error) {
	// Handle SSH URLs (git@host:owner/repo.git)
	if strings.HasPrefix(gitURL, "git@") {
		parts := strings.Split(gitURL, ":")
		if len(parts) != 2 {
			return "", "", errors.Errorf("invalid SSH URL format: %s", gitURL)
		}
		
		hostPart := parts[0][4:] // Remove "git@" prefix
		pathPart := strings.TrimSuffix(parts[1], ".git")
		
		baseURL = "https://" + hostPart
		slug = pathPart
		return baseURL, slug, nil
	}

	// Parse as regular URL
	u, err := url.Parse(gitURL)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to parse URL")
	}

	// Extract base URL (scheme + host)
	baseURL = u.Scheme + "://" + u.Host

	// Extract path and clean it up
	path := strings.TrimPrefix(u.Path, "/")
	path = strings.TrimSuffix(path, ".git")
	
	// Validate that we have owner/repo format
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return "", "", errors.Errorf("URL does not contain owner/repo path: %s", gitURL)
	}

	// Take the first two parts as owner/repo, ignoring any additional path components
	slug = parts[0] + "/" + parts[1]
	
	return baseURL, slug, nil
}

// IsGitLabURL checks if a URL is a GitLab URL
func IsGitLabURL(gitURL string) bool {
	// Handle SSH URLs
	if strings.HasPrefix(gitURL, "git@") {
		return strings.Contains(gitURL, "gitlab")
	}

	// Parse as regular URL
	u, err := url.Parse(gitURL)
	if err != nil {
		return false
	}

	// Check if hostname contains "gitlab"
	return strings.Contains(u.Host, "gitlab")
}