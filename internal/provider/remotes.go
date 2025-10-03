package provider

import (
	"context"
	"strings"

	"emperror.dev/errors"
	"github.com/aviator-co/av/internal/config"
	"github.com/aviator-co/av/internal/git"
	"github.com/sirupsen/logrus"
)

// RemoteInfo contains information about a git remote
type RemoteInfo struct {
	Name string
	URL  string
	Type ProviderType
	RepoSlug string
}

// DetectFromAllRemotes detects the provider from all git remotes, not just origin.
// It checks remotes in order of preference: origin, upstream, then others.
func DetectFromAllRemotes(ctx context.Context, repo *git.Repo) (*Provider, error) {
	remotes, err := getAllRemotes(ctx, repo)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get git remotes")
	}
	
	if len(remotes) == 0 {
		return nil, errors.New("no git remotes found")
	}
	
	// Try remotes in order of preference
	preferredOrder := []string{"origin", "upstream"}

	// First, try preferred remotes
	for _, preferredName := range preferredOrder {
		for _, remote := range remotes {
			if remote.Name == preferredName && remote.Type != Unknown {
				logrus.WithFields(logrus.Fields{
					"remote":   remote.Name,
					"url":      remote.URL,
					"provider": remote.Type,
				}).Debug("detected provider from preferred remote")

				return &Provider{
					Type:         remote.Type,
					BaseURL:      determineBaseURLFromRemote(remote),
					RepoSlug:     remote.RepoSlug,
					IsEnterprise: isEnterpriseFromRemote(remote),
				}, nil
			}
		}
	}

	// If no preferred remotes found, try any remote with a recognized provider
	for _, remote := range remotes {
		if remote.Type != Unknown {
			logrus.WithFields(logrus.Fields{
				"remote":   remote.Name,
				"url":      remote.URL,
				"provider": remote.Type,
			}).Debug("detected provider from remote")

			return &Provider{
				Type:         remote.Type,
				BaseURL:      determineBaseURLFromRemote(remote),
				RepoSlug:     remote.RepoSlug,
				IsEnterprise: isEnterpriseFromRemote(remote),
			}, nil
		}
	}
	
	return nil, errors.New("no supported provider detected from git remotes")
}

// getAllRemotes gets all git remotes and their provider information
func getAllRemotes(ctx context.Context, repo *git.Repo) ([]*RemoteInfo, error) {
	// Get the configured remote name (defaults to "origin")
	remoteName := getConfiguredRemoteName()
	
	// First try the configured remote
	if remoteName != "" {
		if remote, err := getRemoteInfo(ctx, repo, remoteName); err == nil {
			return []*RemoteInfo{remote}, nil
		}
	}
	
	// Fall back to discovering all remotes via git command
	output, err := repo.Run(ctx, &git.RunOpts{
		Args: []string{"remote", "-v"},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list git remotes")
	}
	
	if output.ExitCode != 0 {
		return nil, errors.New("git remote command failed")
	}
	
	return parseRemoteOutput(string(output.Stdout))
}

// getConfiguredRemoteName returns the configured remote name from config
func getConfiguredRemoteName() string {
	if config.Av.Remote != "" {
		return config.Av.Remote
	}
	return git.DEFAULT_REMOTE_NAME
}

// getRemoteInfo gets information about a specific remote
func getRemoteInfo(ctx context.Context, repo *git.Repo, remoteName string) (*RemoteInfo, error) {
	output, err := repo.Run(ctx, &git.RunOpts{
		Args: []string{"remote", "get-url", remoteName},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get URL for remote %s", remoteName)
	}
	
	if output.ExitCode != 0 {
		return nil, errors.Errorf("remote %s not found", remoteName)
	}
	
	remoteURL := strings.TrimSpace(string(output.Stdout))
	if remoteURL == "" {
		return nil, errors.Errorf("empty URL for remote %s", remoteName)
	}
	
	// Detect provider and repo slug from URL
	provider, err := detectFromURL(remoteURL, "")
	if err != nil {
		return &RemoteInfo{
			Name: remoteName,
			URL:  remoteURL,
			Type: Unknown,
		}, nil
	}
	
	return &RemoteInfo{
		Name:     remoteName,
		URL:      remoteURL,
		Type:     provider.Type,
		RepoSlug: provider.RepoSlug,
	}, nil
}

// parseRemoteOutput parses the output of "git remote -v" command
func parseRemoteOutput(output string) ([]*RemoteInfo, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	remoteMap := make(map[string]string) // name -> URL

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		remoteName := parts[0]
		remoteURL := parts[1]

		// Only keep fetch URLs (ignore push URLs that might be different)
		if len(parts) >= 3 && parts[2] == "(push)" {
			continue
		}

		remoteMap[remoteName] = remoteURL
	}

	var remotes []*RemoteInfo
	for name, url := range remoteMap {
		// Detect provider for this remote
		provider, err := detectFromURL(url, "")
		remoteType := Unknown
		repoSlug := ""

		if err == nil {
			remoteType = provider.Type
			repoSlug = provider.RepoSlug
		}

		remotes = append(remotes, &RemoteInfo{
			Name:     name,
			URL:      url,
			Type:     remoteType,
			RepoSlug: repoSlug,
		})
	}
	
	return remotes, nil
}

// determineBaseURLFromRemote determines the base URL from remote info
func determineBaseURLFromRemote(remote *RemoteInfo) string {
	switch remote.Type {
	case GitHub:
		if hostname := extractHostnameFromURL(remote.URL); hostname != "" {
			return determineGitHubBaseURL(hostname)
		}
		return "https://github.com"

	case GitLab:
		if hostname := extractHostnameFromURL(remote.URL); hostname != "" {
			return determineGitLabBaseURL(hostname)
		}
		return "https://gitlab.com"

	default:
		return ""
	}
}

// isEnterpriseFromRemote determines if the remote is an enterprise instance
func isEnterpriseFromRemote(remote *RemoteInfo) bool {
	hostname := extractHostnameFromURL(remote.URL)
	if hostname == "" {
		return false
	}
	
	hostname = strings.ToLower(hostname)
	
	switch remote.Type {
	case GitHub:
		return hostname != "github.com"
	case GitLab:
		return hostname != "gitlab.com"
	default:
		return false
	}
}

// extractHostnameFromURL extracts hostname from a git URL (SSH or HTTPS)
func extractHostnameFromURL(gitURL string) string {
	// Handle SSH URLs (git@host:repo.git)
	if strings.HasPrefix(gitURL, "git@") && strings.Contains(gitURL, ":") {
		parts := strings.Split(gitURL, "@")
		if len(parts) >= 2 {
			hostRepo := parts[1]
			hostParts := strings.Split(hostRepo, ":")
			if len(hostParts) >= 1 {
				return hostParts[0]
			}
		}
	}

	// Handle HTTPS URLs
	if strings.HasPrefix(gitURL, "https://") || strings.HasPrefix(gitURL, "http://") {
		// Remove protocol
		gitURL = strings.TrimPrefix(gitURL, "https://")
		gitURL = strings.TrimPrefix(gitURL, "http://")

		// Extract hostname (before first slash)
		if slashIndex := strings.Index(gitURL, "/"); slashIndex != -1 {
			return gitURL[:slashIndex]
		}
		return gitURL
	}

	return ""
}
