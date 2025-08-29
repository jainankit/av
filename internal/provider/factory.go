package provider

import (
	"context"
	"net/url"
	"strings"

	"emperror.dev/errors"
	
	"github.com/aviator-co/av/internal/config"
)

// ProviderConfig represents the configuration needed to create a provider.
type ProviderConfig struct {
	// Type specifies the provider type (GitHub or GitLab).
	Type ProviderType
	// Token is the authentication token for the provider.
	Token string
	// BaseURL is the base URL for the provider API (optional for public providers).
	BaseURL string
}

// NewProvider creates a new provider instance based on the provided configuration.
func NewProvider(ctx context.Context, providerType ProviderType, config ProviderConfig) (Provider, error) {
	if err := validateProviderConfig(providerType, config); err != nil {
		return nil, err
	}

	switch providerType {
	case GitHub:
		return newGitHubProvider(ctx, config)
	case GitLab:
		return newGitLabProvider(ctx, config)
	default:
		return nil, errors.Errorf("unsupported provider type: %s", providerType)
	}
}

// DetectProviderFromURL detects the provider type from a git remote URL.
func DetectProviderFromURL(remoteURL string) (ProviderType, error) {
	if remoteURL == "" {
		return "", errors.New("remote URL is empty")
	}

	// Parse the URL to extract the hostname
	parsedURL, err := parseGitURL(remoteURL)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse remote URL: %s", remoteURL)
	}

	hostname := strings.ToLower(parsedURL.Host)

	// Check for GitLab patterns first (more specific)
	if isGitLabURL(hostname) {
		return GitLab, nil
	}

	// Check for GitHub patterns
	if isGitHubURL(hostname) {
		return GitHub, nil
	}

	// Default to GitHub for unknown patterns (backward compatibility)
	return GitHub, nil
}

// parseGitURL parses various git URL formats into a standard URL.
func parseGitURL(gitURL string) (*url.URL, error) {
	// Handle SSH URLs like git@github.com:owner/repo.git
	if strings.HasPrefix(gitURL, "git@") || strings.Contains(gitURL, "@") && !strings.HasPrefix(gitURL, "http") {
		// Convert SSH format to HTTPS for parsing
		parts := strings.Split(gitURL, "@")
		if len(parts) != 2 {
			return nil, errors.New("invalid SSH URL format")
		}
		
		hostPath := parts[1]
		colonIndex := strings.Index(hostPath, ":")
		if colonIndex == -1 {
			return nil, errors.New("invalid SSH URL format: missing colon")
		}
		
		host := hostPath[:colonIndex]
		path := hostPath[colonIndex+1:]
		
		return &url.URL{
			Scheme: "https",
			Host:   host,
			Path:   "/" + path,
		}, nil
	}

	// Handle HTTPS URLs
	if strings.HasPrefix(gitURL, "https://") || strings.HasPrefix(gitURL, "http://") {
		return url.Parse(gitURL)
	}

	// Handle other formats by trying to parse directly
	return url.Parse(gitURL)
}

// isGitHubURL checks if the hostname indicates a GitHub instance.
func isGitHubURL(hostname string) bool {
	return hostname == "github.com" || strings.HasSuffix(hostname, ".github.com")
}

// isGitLabURL checks if the hostname indicates a GitLab instance.
func isGitLabURL(hostname string) bool {
	// Common GitLab patterns
	gitlabPatterns := []string{
		"gitlab.com",
		".gitlab.com",
		"gitlab.",
		".gitlab.",
	}

	for _, pattern := range gitlabPatterns {
		if hostname == strings.TrimPrefix(pattern, ".") || strings.Contains(hostname, pattern) {
			return true
		}
	}

	return false
}

// validateProviderConfig validates the provider configuration.
func validateProviderConfig(providerType ProviderType, config ProviderConfig) error {
	if providerType == "" {
		return errors.New("provider type is required")
	}

	if config.Token == "" {
		return errors.Errorf("authentication token is required for %s provider", providerType)
	}

	// Validate provider-specific requirements
	switch providerType {
	case GitHub:
		// GitHub token validation - basic check for format
		if !isValidGitHubToken(config.Token) {
			return errors.New("invalid GitHub token format")
		}
	case GitLab:
		// GitLab token validation - basic check for format
		if !isValidGitLabToken(config.Token) {
			return errors.New("invalid GitLab token format")
		}
	default:
		return errors.Errorf("unsupported provider type: %s", providerType)
	}

	return nil
}

// isValidGitHubToken performs basic validation on GitHub token format.
func isValidGitHubToken(token string) bool {
	if token == "" {
		return false
	}
	
	// GitHub personal access tokens typically start with ghp_, gho_, ghu_, ghs_, or ghr_
	// Classic tokens are 40-character hex strings
	if strings.HasPrefix(token, "ghp_") || strings.HasPrefix(token, "gho_") ||
		strings.HasPrefix(token, "ghu_") || strings.HasPrefix(token, "ghs_") ||
		strings.HasPrefix(token, "ghr_") {
		return len(token) > 4 // Basic length check
	}
	
	// Classic token format (40 hex characters)
	if len(token) == 40 {
		for _, char := range token {
			if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
				return false
			}
		}
		return true
	}
	
	return false
}

// isValidGitLabToken performs basic validation on GitLab token format.
func isValidGitLabToken(token string) bool {
	if token == "" {
		return false
	}
	
	// GitLab personal access tokens are typically alphanumeric with hyphens/underscores
	// They're usually at least 20 characters long
	if len(token) < 20 {
		return false
	}
	
	// Check for valid characters (letters, numbers, hyphens, underscores)
	for _, char := range token {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'z') || 
			(char >= 'A' && char <= 'Z') || char == '-' || char == '_') {
			return false
		}
	}
	
	return true
}

// newGitHubProvider creates a new GitHub provider instance.
// This is a placeholder implementation that will be completed in Step 3.1.
func newGitHubProvider(ctx context.Context, config ProviderConfig) (Provider, error) {
	// TODO: Implement GitHub provider creation in Step 3.1
	return nil, errors.New("GitHub provider implementation not yet available")
}

// LoadProviderConfig loads provider configuration from the global config.
func LoadProviderConfig(providerType ProviderType) (ProviderConfig, error) {
	cfg := ProviderConfig{Type: providerType}
	
	switch providerType {
	case GitHub:
		cfg.Token = config.Av.GitHub.Token
		cfg.BaseURL = config.Av.GitHub.BaseURL
	case GitLab:
		cfg.Token = config.Av.GitLab.Token
		cfg.BaseURL = config.Av.GitLab.BaseURL
		// Default to gitlab.com if no base URL is provided
		if cfg.BaseURL == "" {
			cfg.BaseURL = "https://gitlab.com/"
		}
	default:
		return cfg, errors.Errorf("unsupported provider type: %s", providerType)
	}
	
	return cfg, nil
}

// GetProviderTypeFromConfig gets the provider type from configuration or auto-detects it.
func GetProviderTypeFromConfig(remoteURL string) (ProviderType, error) {
	// First check if provider type is explicitly configured
	if config.Av.Provider.Type != "" {
		switch strings.ToLower(config.Av.Provider.Type) {
		case "github":
			return GitHub, nil
		case "gitlab":
			return GitLab, nil
		default:
			return "", errors.Errorf("invalid provider type in configuration: %s", config.Av.Provider.Type)
		}
	}
	
	// Fall back to URL detection
	return DetectProviderFromURL(remoteURL)
}

// NewProviderFromConfig creates a provider instance using the global configuration.
// This is a convenience function that combines LoadProviderConfig and NewProvider.
func NewProviderFromConfig(ctx context.Context, remoteURL string) (Provider, error) {
	providerType, err := GetProviderTypeFromConfig(remoteURL)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to determine provider type")
	}
	
	config, err := LoadProviderConfig(providerType)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load provider configuration")
	}
	
	return NewProvider(ctx, providerType, config)
}

// newGitHubProvider creates a new GitHub provider instance.
// This is a placeholder implementation that will be completed in Step 3.1.
func newGitHubProvider(ctx context.Context, config ProviderConfig) (Provider, error) {
	// TODO: Implement GitHub provider creation in Step 3.1
	return nil, errors.New("GitHub provider implementation not yet available")
}

// newGitLabProvider creates a new GitLab provider instance.
// This is a placeholder implementation that will be completed in Step 2.4.
func newGitLabProvider(ctx context.Context, config ProviderConfig) (Provider, error) {
	// TODO: Implement GitLab provider creation in Step 2.4
	return nil, errors.New("GitLab provider implementation not yet available")
}