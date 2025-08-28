package provider

import (
	"context"
	"fmt"
	"strings"
)

// ProviderConfig contains configuration for a specific provider
type ProviderConfig struct {
	Name     ProviderName `json:"name"`
	BaseURL  string       `json:"base_url,omitempty"`
	Token    string       `json:"token"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// NewProvider creates a new provider instance based on the configuration
func NewProvider(ctx context.Context, config ProviderConfig) (Provider, error) {
	switch config.Name {
	case ProviderGitHub:
		// Will be implemented in Step 1.2 when we create the GitHub provider
		return nil, fmt.Errorf("GitHub provider implementation not yet available")
	case ProviderGitLab:
		// Will be implemented in Step 2 when we create the GitLab provider
		return nil, fmt.Errorf("GitLab provider implementation not yet available")
	default:
		return nil, fmt.Errorf("unsupported provider: %s", config.Name)
	}
}

// DetectProviderFromURL attempts to detect the provider from a git remote URL
func DetectProviderFromURL(remoteURL string) (ProviderName, error) {
	if remoteURL == "" {
		return "", fmt.Errorf("empty remote URL")
	}

	// Normalize the URL for detection
	url := strings.ToLower(remoteURL)
	
	// Remove common prefixes
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "git@")
	url = strings.TrimPrefix(url, "ssh://")
	
	// Convert SSH format to HTTP-like format for easier parsing
	if strings.Contains(url, ":") && !strings.Contains(url, "//") {
		parts := strings.SplitN(url, ":", 2)
		if len(parts) == 2 {
			url = parts[0] + "/" + parts[1]
		}
	}
	
	// Remove .git suffix
	url = strings.TrimSuffix(url, ".git")
	
	// Check for known provider domains
	if strings.HasPrefix(url, "github.com") || strings.Contains(url, "github.com") {
		return ProviderGitHub, nil
	}
	
	if strings.HasPrefix(url, "gitlab.com") || strings.Contains(url, "gitlab.com") {
		return ProviderGitLab, nil
	}
	
	// Check for self-hosted GitLab instances (common patterns)
	// This is a heuristic and may need adjustment based on actual usage
	if strings.Contains(url, "gitlab") {
		return ProviderGitLab, nil
	}
	
	return "", fmt.Errorf("unable to detect provider from URL: %s", remoteURL)
}

// ValidateProviderConfig validates that the provider configuration is complete and valid
func ValidateProviderConfig(config ProviderConfig) error {
	if config.Name == "" {
		return fmt.Errorf("provider name is required")
	}
	
	if config.Token == "" {
		return fmt.Errorf("provider token is required")
	}
	
	switch config.Name {
	case ProviderGitHub:
		// GitHub-specific validation can be added here
		return nil
	case ProviderGitLab:
		// GitLab-specific validation can be added here
		// GitLab typically requires a base URL for self-hosted instances
		return nil
	default:
		return fmt.Errorf("unsupported provider: %s", config.Name)
	}
}