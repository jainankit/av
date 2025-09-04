package gitlab

import (
	"fmt"
	"net/url"
	"strings"
)

// URLBuilder handles GitLab URL construction for various resources
type URLBuilder struct {
	baseURL string
}

// NewURLBuilder creates a new URL builder for GitLab
func NewURLBuilder(baseURL string) *URLBuilder {
	if baseURL == "" {
		baseURL = "https://gitlab.com"
	}
	
	// Normalize base URL by removing trailing slash and /api/v4 suffix
	baseURL = strings.TrimSuffix(baseURL, "/")
	baseURL = strings.TrimSuffix(baseURL, "/api/v4")
	
	return &URLBuilder{baseURL: baseURL}
}

// ProjectURL generates a URL for a GitLab project
func (u *URLBuilder) ProjectURL(owner, name string) string {
	return fmt.Sprintf("%s/%s/%s", u.baseURL, owner, name)
}

// MergeRequestURL generates a URL for a GitLab merge request
func (u *URLBuilder) MergeRequestURL(owner, name string, mrNumber int64) string {
	return fmt.Sprintf("%s/%s/%s/-/merge_requests/%d", u.baseURL, owner, name, mrNumber)
}

// IssueURL generates a URL for a GitLab issue
func (u *URLBuilder) IssueURL(owner, name string, issueNumber int64) string {
	return fmt.Sprintf("%s/%s/%s/-/issues/%d", u.baseURL, owner, name, issueNumber)
}

// BranchURL generates a URL for a specific branch
func (u *URLBuilder) BranchURL(owner, name, branch string) string {
	escapedBranch := url.PathEscape(branch)
	return fmt.Sprintf("%s/%s/%s/-/tree/%s", u.baseURL, owner, name, escapedBranch)
}

// CommitURL generates a URL for a specific commit
func (u *URLBuilder) CommitURL(owner, name, sha string) string {
	return fmt.Sprintf("%s/%s/%s/-/commit/%s", u.baseURL, owner, name, sha)
}

// CompareURL generates a URL for comparing two branches or commits
func (u *URLBuilder) CompareURL(owner, name, from, to string) string {
	escapedFrom := url.PathEscape(from)
	escapedTo := url.PathEscape(to)
	return fmt.Sprintf("%s/%s/%s/-/compare/%s...%s", u.baseURL, owner, name, escapedFrom, escapedTo)
}

// FileURL generates a URL for a specific file in the repository
func (u *URLBuilder) FileURL(owner, name, branch, filePath string) string {
	escapedBranch := url.PathEscape(branch)
	escapedPath := strings.TrimPrefix(filePath, "/")
	return fmt.Sprintf("%s/%s/%s/-/blob/%s/%s", u.baseURL, owner, name, escapedBranch, escapedPath)
}

// UserURL generates a URL for a GitLab user profile
func (u *URLBuilder) UserURL(username string) string {
	return fmt.Sprintf("%s/%s", u.baseURL, username)
}

// GroupURL generates a URL for a GitLab group
func (u *URLBuilder) GroupURL(groupPath string) string {
	return fmt.Sprintf("%s/%s", u.baseURL, groupPath)
}

// ParseGitLabURL parses a GitLab URL to extract components
func ParseGitLabURL(gitlabURL string) (*GitLabURLComponents, error) {
	parsed, err := url.Parse(gitlabURL)
	if err != nil {
		return nil, err
	}

	baseURL := fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host)
	
	// Remove leading slash from path
	path := strings.TrimPrefix(parsed.Path, "/")
	pathParts := strings.Split(path, "/")
	
	if len(pathParts) < 2 {
		return &GitLabURLComponents{
			BaseURL: baseURL,
			Type:    "unknown",
		}, nil
	}

	components := &GitLabURLComponents{
		BaseURL: baseURL,
		Owner:   pathParts[0],
		Name:    pathParts[1],
	}

	// Determine URL type and extract additional information
	if len(pathParts) >= 4 && pathParts[2] == "-" {
		switch pathParts[3] {
		case "merge_requests":
			components.Type = "merge_request"
			if len(pathParts) >= 5 {
				components.Number = pathParts[4]
			}
		case "issues":
			components.Type = "issue"
			if len(pathParts) >= 5 {
				components.Number = pathParts[4]
			}
		case "tree":
			components.Type = "branch"
			if len(pathParts) >= 5 {
				components.Branch = strings.Join(pathParts[4:], "/")
			}
		case "blob":
			components.Type = "file"
			if len(pathParts) >= 5 {
				components.Branch = pathParts[4]
				if len(pathParts) >= 6 {
					components.FilePath = strings.Join(pathParts[5:], "/")
				}
			}
		case "commit":
			components.Type = "commit"
			if len(pathParts) >= 5 {
				components.SHA = pathParts[4]
			}
		case "compare":
			components.Type = "compare"
			if len(pathParts) >= 5 {
				compareSpec := pathParts[4]
				if strings.Contains(compareSpec, "...") {
					compareParts := strings.Split(compareSpec, "...")
					if len(compareParts) == 2 {
						components.FromRef = compareParts[0]
						components.ToRef = compareParts[1]
					}
				}
			}
		default:
			components.Type = "project"
		}
	} else if len(pathParts) == 2 {
		components.Type = "project"
	} else if len(pathParts) == 1 {
		// Could be a user or group
		components.Type = "user_or_group"
		components.Owner = ""
		components.Name = pathParts[0]
	}

	return components, nil
}

// GitLabURLComponents represents parsed components of a GitLab URL
type GitLabURLComponents struct {
	BaseURL  string
	Owner    string
	Name     string
	Type     string // project, merge_request, issue, branch, file, commit, compare, user_or_group
	Number   string // for MRs and issues
	Branch   string
	FilePath string
	SHA      string
	FromRef  string // for compare URLs
	ToRef    string // for compare URLs
}

// IsGitLabURL determines if a URL is a GitLab URL
func IsGitLabURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	
	host := strings.ToLower(parsed.Host)
	return host == "gitlab.com" || 
		   host == "www.gitlab.com" ||
		   strings.Contains(host, "gitlab") // basic heuristic for self-hosted
}

// NormalizeGitLabBaseURL normalizes a GitLab base URL for API usage
func NormalizeGitLabBaseURL(baseURL string) string {
	if baseURL == "" {
		return ""
	}
	
	baseURL = strings.TrimSuffix(baseURL, "/")
	
	// If it doesn't end with /api/v4, add it
	if !strings.HasSuffix(baseURL, "/api/v4") {
		baseURL += "/api/v4"
	}
	
	return baseURL
}