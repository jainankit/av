package provider

import (
	"context"
	"regexp"
	"strings"

	"emperror.dev/errors"
	giturls "github.com/chainguard-dev/git-urls"
	"github.com/aviator-co/av/internal/git"
)

// DetectProvider determines the Git provider type based on the remote URL
func DetectProvider(ctx context.Context, repo *git.Repo) (ProviderType, error) {
	origin, err := repo.Origin(ctx)
	if err != nil {
		return "", errors.Wrap(err, "failed to determine repository origin")
	}

	return DetectProviderFromURL(origin.URL), nil
}

// DetectProviderFromURL determines the Git provider type from a remote URL
func DetectProviderFromURL(remoteURL string) ProviderType {
	u, err := giturls.Parse(remoteURL)
	if err != nil {
		return ""
	}

	host := strings.ToLower(u.Host)

	// Handle GitHub patterns
	if isGitHubHost(host) {
		return GitHub
	}

	// Handle GitLab patterns
	if isGitLabHost(host) {
		return GitLab
	}

	// Default to GitHub if we can't determine (for backward compatibility)
	return GitHub
}

// isGitHubHost checks if the host is a GitHub instance
func isGitHubHost(host string) bool {
	// GitHub.com
	if host == "github.com" {
		return true
	}

	// GitHub Enterprise Server patterns
	// Common patterns include:
	// - github.enterprise.com
	// - git.company.com
	// - github.company.internal
	githubPatterns := []string{
		`^github\..*`,
		`^git\..*`,
		`.*github.*`,
	}

	for _, pattern := range githubPatterns {
		if matched, _ := regexp.MatchString(pattern, host); matched {
			return true
		}
	}

	return false
}

// isGitLabHost checks if the host is a GitLab instance
func isGitLabHost(host string) bool {
	// GitLab.com
	if host == "gitlab.com" {
		return true
	}

	// Self-hosted GitLab patterns
	// Common patterns include:
	// - gitlab.company.com
	// - git.company.com (when not GitHub)
	// - code.company.com
	gitlabPatterns := []string{
		`^gitlab\..*`,
		`^git\..*gitlab.*`,
		`^code\..*`,
		`.*gitlab.*`,
	}

	for _, pattern := range gitlabPatterns {
		if matched, _ := regexp.MatchString(pattern, host); matched {
			return true
		}
	}

	return false
}

// GetRepoSlugFromURL extracts the repository slug (owner/repo) from a remote URL
func GetRepoSlugFromURL(remoteURL string) (string, error) {
	u, err := giturls.Parse(remoteURL)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse remote URL %q", remoteURL)
	}

	repoSlug := strings.TrimSuffix(u.Path, ".git")
	repoSlug = strings.TrimPrefix(repoSlug, "/")
	
	if repoSlug == "" {
		return "", errors.Errorf("unable to determine repository slug from URL %q", remoteURL)
	}

	return repoSlug, nil
}