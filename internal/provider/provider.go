// Package provider handles provider selection and detection for GitHub and GitLab.
//
// The provider system detects the appropriate provider (GitHub or GitLab) based on:
// 1. Manual configuration overrides (config file or command-line flags)
// 2. Automatic detection from git remote URLs
// 3. Environment variables and CLI tool authentication
//
// Example usage:
//
//	ctx := context.Background()
//	repo, err := git.GetRepo()
//	if err != nil {
//		return err
//	}
//	
//	provider, err := provider.DetectProvider(ctx, repo)
//	if err != nil {
//		return err
//	}
//	
//	switch provider.Type {
//	case provider.GitHub:
//		// Use GitHub client
//	case provider.GitLab:
//		// Use GitLab client
//	}
package provider

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"emperror.dev/errors"
	"github.com/aviator-co/av/internal/config"
	"github.com/aviator-co/av/internal/git"
	"github.com/aviator-co/av/internal/gitlab"
	"github.com/sirupsen/logrus"
)

// ProviderType represents the different provider types supported
type ProviderType string

const (
	// GitHub represents the GitHub provider
	GitHub ProviderType = "github"
	// GitLab represents the GitLab provider
	GitLab ProviderType = "gitlab"
	// Unknown represents an unknown or unsupported provider
	Unknown ProviderType = "unknown"
)

// String returns the string representation of the provider type
func (p ProviderType) String() string {
	return string(p)
}

// Provider contains information about the detected provider
type Provider struct {
	// Type is the detected provider type (GitHub or GitLab)
	Type ProviderType
	
	// BaseURL is the base URL for the provider (e.g., https://github.com, https://gitlab.com)
	BaseURL string
	
	// RepoSlug is the repository slug (e.g., "owner/repo" for GitHub, "namespace/project" for GitLab)
	RepoSlug string
	
	// IsEnterprise indicates if this is an enterprise/self-hosted instance
	IsEnterprise bool
	
	// Token is the authentication token for the provider (if available)
	Token string
}

// DetectionOptions configures provider detection behavior
type DetectionOptions struct {
	// ForceProvider can be used to force a specific provider type,
	// overriding automatic detection
	ForceProvider ProviderType
	
	// AllowFallback determines if detection should fall back to
	// configuration-based detection when remote detection fails
	AllowFallback bool
	
	// RequireAuthentication determines if detection should fail when
	// no valid authentication token is found
	RequireAuthentication bool
}

// DefaultDetectionOptions returns default options for provider detection
func DefaultDetectionOptions() DetectionOptions {
	return DetectionOptions{
		ForceProvider:         Unknown,
		AllowFallback:        true,
		RequireAuthentication: false,
	}
}

// DetectProvider automatically detects the provider type based on git remote
// URLs and configuration settings.
func DetectProvider(ctx context.Context, repo *git.Repo) (*Provider, error) {
	return DetectProviderWithOptions(ctx, repo, DefaultDetectionOptions())
}

// DetectProviderWithOptions detects the provider using the specified options
func DetectProviderWithOptions(
	ctx context.Context,
	repo *git.Repo,
	options DetectionOptions,
) (*Provider, error) {
	logrus.Debug("starting provider detection")
	
	// Check for manual override first
	if options.ForceProvider != Unknown {
		logrus.WithField("provider", options.ForceProvider).Debug("using forced provider")
		return createProviderFromType(ctx, repo, options.ForceProvider, options)
	}
	
	// Try detection from git remote URL
	if repo != nil {
		if provider, err := detectFromGitRemote(ctx, repo); err == nil {
			logrus.WithField("provider", provider.Type).Debug("detected provider from git remote")
			
			// Enhance with authentication info
			if err := enhanceWithAuthentication(ctx, provider); err != nil {
				if options.RequireAuthentication {
					return nil, errors.Wrap(err, "failed to find authentication for detected provider")
				}
				logrus.WithError(err).Debug("failed to find authentication, continuing without token")
			}
			
			return provider, nil
		}
	}
	
	// Fall back to configuration-based detection if allowed
	if options.AllowFallback {
		logrus.Debug("falling back to configuration-based provider detection")
		return detectFromConfiguration(ctx, repo, options)
	}
	
	return nil, errors.New("unable to detect provider: no git remote found and fallback disabled")
}

// detectFromGitRemote detects the provider from git remote URL
func detectFromGitRemote(ctx context.Context, repo *git.Repo) (*Provider, error) {
	// First try the enhanced multi-remote detection
	if provider, err := DetectFromAllRemotes(ctx, repo); err == nil {
		return provider, nil
	}
	
	// Fall back to origin-only detection for backward compatibility
	origin, err := repo.Origin(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get git origin")
	}
	
	return detectFromURL(origin.URL.String(), origin.RepoSlug)
}

// detectFromURL detects the provider from a git URL
func detectFromURL(gitURL, repoSlug string) (*Provider, error) {
	// Parse the URL to get hostname
	parsedURL, err := url.Parse(gitURL)
	if err != nil {
		// Try parsing SSH URLs (git@host:repo.git)
		if strings.HasPrefix(gitURL, "git@") && strings.Contains(gitURL, ":") {
			parts := strings.Split(gitURL, "@")
			if len(parts) == 2 {
				hostRepo := parts[1]
				hostParts := strings.Split(hostRepo, ":")
				if len(hostParts) >= 2 {
					hostname := hostParts[0]
					return detectFromHostname(hostname, gitURL, repoSlug)
				}
			}
		}
		return nil, errors.Wrapf(err, "failed to parse git URL: %s", gitURL)
	}
	
	hostname := parsedURL.Hostname()
	return detectFromHostname(hostname, gitURL, repoSlug)
}

// detectFromHostname detects provider from hostname
func detectFromHostname(hostname, gitURL, repoSlug string) (*Provider, error) {
	hostname = strings.ToLower(hostname)
	
	// GitHub detection
	if hostname == "github.com" || strings.HasSuffix(hostname, "github.com") {
		return &Provider{
			Type:         GitHub,
			BaseURL:      determineGitHubBaseURL(hostname),
			RepoSlug:     repoSlug,
			IsEnterprise: hostname != "github.com",
		}, nil
	}
	
	// GitLab detection
	if hostname == "gitlab.com" || strings.Contains(hostname, "gitlab") {
		return &Provider{
			Type:         GitLab,
			BaseURL:      determineGitLabBaseURL(hostname),
			RepoSlug:     repoSlug,
			IsEnterprise: hostname != "gitlab.com",
		}, nil
	}
	
	return nil, errors.Errorf("unsupported provider for hostname: %s", hostname)
}

// determineGitHubBaseURL determines the GitHub base URL from hostname
func determineGitHubBaseURL(hostname string) string {
	if hostname == "github.com" {
		return "https://github.com"
	}
	return fmt.Sprintf("https://%s", hostname)
}

// determineGitLabBaseURL determines the GitLab base URL from hostname
func determineGitLabBaseURL(hostname string) string {
	if hostname == "gitlab.com" {
		return "https://gitlab.com"
	}
	return fmt.Sprintf("https://%s", hostname)
}

// detectFromConfiguration detects provider from configuration settings
func detectFromConfiguration(
	ctx context.Context,
	repo *git.Repo,
	options DetectionOptions,
) (*Provider, error) {
	// Check if GitHub is configured
	if config.Av.GitHub.Token != "" || config.Av.GitHub.BaseURL != "" {
		logrus.Debug("detected GitHub from configuration")
		return createProviderFromType(ctx, repo, GitHub, options)
	}
	
	// Check if GitLab is configured
	if config.Av.GitLab.Token != "" || config.Av.GitLab.BaseURL != "" {
		logrus.Debug("detected GitLab from configuration")
		return createProviderFromType(ctx, repo, GitLab, options)
	}
	
	return nil, errors.New("no provider configuration found")
}

// createProviderFromType creates a Provider struct from a provider type
func createProviderFromType(
	ctx context.Context,
	repo *git.Repo,
	providerType ProviderType,
	options DetectionOptions,
) (*Provider, error) {
	provider := &Provider{
		Type: providerType,
	}
	
	switch providerType {
	case GitHub:
		provider.BaseURL = config.Av.GitHub.BaseURL
		if provider.BaseURL == "" {
			provider.BaseURL = "https://github.com"
		}
		provider.IsEnterprise = provider.BaseURL != "https://github.com"
		
	case GitLab:
		provider.BaseURL = config.Av.GitLab.BaseURL
		if provider.BaseURL == "" {
			provider.BaseURL = "https://gitlab.com"
		}
		provider.IsEnterprise = provider.BaseURL != "https://gitlab.com"
		
	default:
		return nil, errors.Errorf("unsupported provider type: %s", providerType)
	}
	
	// Try to get repo slug from git remote if repo is available
	if repo != nil {
		if origin, err := repo.Origin(ctx); err == nil {
			provider.RepoSlug = origin.RepoSlug
		}
	}
	
	// Add authentication
	if err := enhanceWithAuthentication(ctx, provider); err != nil {
		if options.RequireAuthentication {
			return nil, errors.Wrap(err, "failed to find authentication for provider")
		}
		logrus.WithError(err).Debug("failed to find authentication, continuing without token")
	}
	
	return provider, nil
}

// enhanceWithAuthentication adds authentication information to the provider
func enhanceWithAuthentication(ctx context.Context, provider *Provider) error {
	switch provider.Type {
	case GitHub:
		token := discoverGitHubToken(ctx)
		if token == "" {
			return errors.New("no GitHub authentication token found")
		}
		provider.Token = token
		
	case GitLab:
		token := discoverGitLabToken(ctx)
		if token == "" {
			return errors.New("no GitLab authentication token found")
		}
		provider.Token = token
		
	default:
		return errors.Errorf("authentication discovery not implemented for provider: %s", provider.Type)
	}
	
	return nil
}

// discoverGitHubToken discovers GitHub authentication token from various sources
func discoverGitHubToken(ctx context.Context) string {
	// Check configuration first
	if config.Av.GitHub.Token != "" {
		return config.Av.GitHub.Token
	}
	
	// Try GitHub CLI
	if token := tryGitHubCLI(ctx); token != "" {
		return token
	}
	
	return ""
}

// discoverGitLabToken discovers GitLab authentication token from various sources
func discoverGitLabToken(ctx context.Context) string {
	// Check configuration first
	if config.Av.GitLab.Token != "" {
		return config.Av.GitLab.Token
	}
	
	// Try GitLab CLI (glab)
	if token := tryGitLabCLI(ctx); token != "" {
		return token
	}
	
	return ""
}

// tryGitHubCLI attempts to get GitHub token from the GitHub CLI
func tryGitHubCLI(ctx context.Context) string {
	token, err := tryGitHubCLIEnhanced(ctx)
	if err != nil {
		logrus.WithError(err).Debug("failed to get GitHub token from CLI")
		return ""
	}
	return token
}

// tryGitLabCLI attempts to get GitLab token from the GitLab CLI (glab)
func tryGitLabCLI(ctx context.Context) string {
	token, err := tryGitLabCLIEnhanced(ctx)
	if err != nil {
		logrus.WithError(err).Debug("failed to get GitLab token from CLI")
		return ""
	}
	return token
}

// ValidateProvider validates that the provider has proper authentication
// and connectivity
func ValidateProvider(ctx context.Context, provider *Provider) error {
	if provider == nil {
		return errors.New("provider is nil")
	}
	
	switch provider.Type {
	case GitHub:
		return validateGitHubProvider(ctx, provider)
	case GitLab:
		return validateGitLabProvider(ctx, provider)
	default:
		return errors.Errorf("validation not supported for provider: %s", provider.Type)
	}
}

// validateGitHubProvider validates GitHub provider authentication and connectivity
func validateGitHubProvider(ctx context.Context, provider *Provider) error {
	if provider.Token == "" {
		return errors.New("GitHub token is required")
	}
	
	// TODO: Implement GitHub API connectivity check
	// This would create a GitHub client and test authentication
	logrus.WithFields(logrus.Fields{
		"provider": "github",
		"base_url": provider.BaseURL,
		"enterprise": provider.IsEnterprise,
	}).Debug("GitHub provider validation not yet implemented")
	
	return nil
}

// validateGitLabProvider validates GitLab provider authentication and connectivity
func validateGitLabProvider(ctx context.Context, provider *Provider) error {
	if provider.Token == "" {
		return errors.New("GitLab token is required")
	}
	
	// TODO: Implement GitLab API connectivity check
	// This would create a GitLab client and test authentication
	logrus.WithFields(logrus.Fields{
		"provider": "gitlab",
		"base_url": provider.BaseURL,
		"enterprise": provider.IsEnterprise,
	}).Debug("GitLab provider validation not yet implemented")
	
	return nil
}

// IsGitHubURL checks if the given URL is a GitHub URL
func IsGitHubURL(gitURL string) bool {
	provider, err := detectFromURL(gitURL, "")
	if err != nil {
		return false
	}
	return provider.Type == GitHub
}

// IsGitLabURL checks if the given URL is a GitLab URL  
func IsGitLabURL(gitURL string) bool {
	// Use the existing implementation from gitlab package
	return gitlab.IsGitLabURL(gitURL)
}

// ParseGitHubURL extracts repository information from a GitHub URL
func ParseGitHubURL(gitURL string) (repoSlug string, err error) {
	provider, err := detectFromURL(gitURL, "")
	if err != nil {
		return "", err
	}
	
	if provider.Type != GitHub {
		return "", errors.New("URL is not a GitHub URL")
	}
	
	// Extract repo slug from URL path
	parsedURL, err := url.Parse(gitURL)
	if err != nil {
		// Handle SSH URLs
		if strings.HasPrefix(gitURL, "git@") && strings.Contains(gitURL, ":") {
			parts := strings.Split(gitURL, ":")
			if len(parts) >= 2 {
				repoPath := parts[len(parts)-1]
				repoPath = strings.TrimSuffix(repoPath, ".git")
				return repoPath, nil
			}
		}
		return "", err
	}
	
	path := strings.TrimPrefix(parsedURL.Path, "/")
	path = strings.TrimSuffix(path, ".git")
	
	if path == "" {
		return "", errors.New("no repository path found in GitHub URL")
	}
	
	return path, nil
}

// ParseGitLabURL extracts repository information from a GitLab URL
func ParseGitLabURL(gitURL string) (gitlab.ProjectSlug, error) {
	// Use the existing implementation from gitlab package
	return gitlab.ParseGitLabURL(gitURL)
}
