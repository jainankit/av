package provider

import (
	"context"

	"emperror.dev/errors"
	"github.com/aviator-co/av/internal/config"
	"github.com/aviator-co/av/internal/git"
	"github.com/sirupsen/logrus"
)

// ConfigOverrides allows overriding provider configuration
type ConfigOverrides struct {
	// Provider forces a specific provider type
	Provider ProviderType
	
	// GitHubToken overrides the GitHub token
	GitHubToken string
	
	// GitLabToken overrides the GitLab token  
	GitLabToken string
	
	// BaseURL overrides the base URL
	BaseURL string
}

// DetectProviderWithConfig detects provider using configuration overrides
func DetectProviderWithConfig(
	ctx context.Context,
	repo *git.Repo,
	overrides *ConfigOverrides,
) (*Provider, error) {
	options := DefaultDetectionOptions()
	
	// Apply overrides to detection options
	if overrides != nil {
		if overrides.Provider != Unknown {
			options.ForceProvider = overrides.Provider
		}
	}
	
	// Detect the provider
	provider, err := DetectProviderWithOptions(ctx, repo, options)
	if err != nil {
		return nil, err
	}
	
	// Apply configuration overrides
	if overrides != nil {
		applyConfigOverrides(provider, overrides)
	}
	
	return provider, nil
}

// applyConfigOverrides applies configuration overrides to the provider
func applyConfigOverrides(provider *Provider, overrides *ConfigOverrides) {
	if provider == nil || overrides == nil {
		return
	}
	
	// Apply token overrides
	switch provider.Type {
	case GitHub:
		if overrides.GitHubToken != "" {
			provider.Token = overrides.GitHubToken
			logrus.Debug("applied GitHub token override from configuration")
		}
	case GitLab:
		if overrides.GitLabToken != "" {
			provider.Token = overrides.GitLabToken
			logrus.Debug("applied GitLab token override from configuration")
		}
	}
	
	// Apply base URL override
	if overrides.BaseURL != "" {
		provider.BaseURL = overrides.BaseURL
		provider.IsEnterprise = isEnterpriseBaseURL(provider.Type, overrides.BaseURL)
		logrus.WithField("base_url", overrides.BaseURL).Debug(
			"applied base URL override from configuration",
		)
	}
}

// GetProviderFromConfig attempts to determine the provider from configuration alone
func GetProviderFromConfig() (*Provider, error) {
	// Check mutual exclusivity (this should already be validated during config load)
	githubConfigured := config.Av.GitHub.Token != "" || config.Av.GitHub.BaseURL != ""
	gitlabConfigured := config.Av.GitLab.Token != "" || config.Av.GitLab.BaseURL != ""
	
	if githubConfigured && gitlabConfigured {
		return nil, errors.New("both GitHub and GitLab are configured - they are mutually exclusive")
	}
	
	// Create provider from GitHub config
	if githubConfigured {
		provider := &Provider{
			Type:    GitHub,
			BaseURL: config.Av.GitHub.BaseURL,
			Token:   config.Av.GitHub.Token,
		}
		
		if provider.BaseURL == "" {
			provider.BaseURL = "https://github.com"
		}
		provider.IsEnterprise = provider.BaseURL != "https://github.com"
		
		return provider, nil
	}
	
	// Create provider from GitLab config
	if gitlabConfigured {
		provider := &Provider{
			Type:    GitLab,
			BaseURL: config.Av.GitLab.BaseURL,
			Token:   config.Av.GitLab.Token,
		}
		
		if provider.BaseURL == "" {
			provider.BaseURL = "https://gitlab.com"
		}
		provider.IsEnterprise = provider.BaseURL != "https://gitlab.com"
		
		return provider, nil
	}
	
	return nil, errors.New("no provider configuration found")
}

// ConfigFromFlags creates config overrides from command-line flags
func ConfigFromFlags(flags *ProviderFlags) (*ConfigOverrides, error) {
	if flags == nil {
		return nil, nil
	}
	
	overrides := &ConfigOverrides{
		GitHubToken: flags.GitHubToken,
		GitLabToken: flags.GitLabToken,
		BaseURL:     flags.BaseURL,
	}
	
	if flags.Provider != "" {
		providerType, err := ParseProviderType(flags.Provider)
		if err != nil {
			return nil, err
		}
		overrides.Provider = providerType
	}
	
	return overrides, nil
}

// ValidateConfigOverrides validates configuration overrides
func ValidateConfigOverrides(overrides *ConfigOverrides) error {
	if overrides == nil {
		return nil
	}
	
	// Check for conflicting tokens
	if overrides.GitHubToken != "" && overrides.GitLabToken != "" {
		return errors.New("cannot specify both GitHub and GitLab tokens in overrides")
	}
	
	// Validate forced provider type
	if overrides.Provider != Unknown && 
		overrides.Provider != GitHub && 
		overrides.Provider != GitLab {
		return errors.Errorf("invalid provider type in overrides: %s", overrides.Provider)
	}
	
	return nil
}

// MergeConfigOverrides merges multiple config override sources
func MergeConfigOverrides(base, override *ConfigOverrides) *ConfigOverrides {
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}
	
	result := *base
	
	// Override provider type
	if override.Provider != Unknown {
		result.Provider = override.Provider
	}
	
	// Override tokens
	if override.GitHubToken != "" {
		result.GitHubToken = override.GitHubToken
	}
	if override.GitLabToken != "" {
		result.GitLabToken = override.GitLabToken
	}
	
	// Override base URL
	if override.BaseURL != "" {
		result.BaseURL = override.BaseURL
	}
	
	return &result
}
