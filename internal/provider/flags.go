package provider

import (
	"strings"

	"emperror.dev/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ProviderFlags contains command-line flags for provider selection
type ProviderFlags struct {
	// Provider forces a specific provider type
	Provider string
	
	// GitHubToken overrides the GitHub token
	GitHubToken string
	
	// GitLabToken overrides the GitLab token
	GitLabToken string
	
	// BaseURL overrides the base URL for the provider
	BaseURL string
}

// AddProviderFlags adds provider-related flags to a command
func AddProviderFlags(cmd *cobra.Command, flags *ProviderFlags) {
	cmd.Flags().StringVar(
		&flags.Provider, "provider", "",
		"force a specific provider (github or gitlab)")
		
	cmd.Flags().StringVar(
		&flags.GitHubToken, "github-token", "",
		"GitHub API token (overrides config)")
		
	cmd.Flags().StringVar(
		&flags.GitLabToken, "gitlab-token", "",
		"GitLab API token (overrides config)")
		
	cmd.Flags().StringVar(
		&flags.BaseURL, "base-url", "",
		"base URL for the provider API (for enterprise instances)")
}

// AddProviderFlagsToFlagSet adds provider-related flags to a flag set
func AddProviderFlagsToFlagSet(fs *pflag.FlagSet, flags *ProviderFlags) {
	fs.StringVar(
		&flags.Provider, "provider", "",
		"force a specific provider (github or gitlab)")
		
	fs.StringVar(
		&flags.GitHubToken, "github-token", "",
		"GitHub API token (overrides config)")
		
	fs.StringVar(
		&flags.GitLabToken, "gitlab-token", "",
		"GitLab API token (overrides config)")
		
	fs.StringVar(
		&flags.BaseURL, "base-url", "",
		"base URL for the provider API (for enterprise instances)")
}

// ParseProviderType parses a provider type string
func ParseProviderType(providerStr string) (ProviderType, error) {
	if providerStr == "" {
		return Unknown, nil
	}
	
	switch strings.ToLower(providerStr) {
	case "github", "gh":
		return GitHub, nil
	case "gitlab", "gl":
		return GitLab, nil
	default:
		return Unknown, errors.Errorf("unsupported provider type: %s", providerStr)
	}
}

// ToDetectionOptions converts provider flags to detection options
func (f *ProviderFlags) ToDetectionOptions() (DetectionOptions, error) {
	options := DefaultDetectionOptions()
	
	if f.Provider != "" {
		providerType, err := ParseProviderType(f.Provider)
		if err != nil {
			return options, err
		}
		options.ForceProvider = providerType
	}
	
	return options, nil
}

// ApplyOverrides applies command-line flag overrides to a provider
func (f *ProviderFlags) ApplyOverrides(provider *Provider) {
	if provider == nil {
		return
	}
	
	// Apply token overrides
	switch provider.Type {
	case GitHub:
		if f.GitHubToken != "" {
			provider.Token = f.GitHubToken
		}
	case GitLab:
		if f.GitLabToken != "" {
			provider.Token = f.GitLabToken
		}
	}
	
	// Apply base URL override
	if f.BaseURL != "" {
		provider.BaseURL = f.BaseURL
		// Recalculate enterprise status
		provider.IsEnterprise = isEnterpriseBaseURL(provider.Type, f.BaseURL)
	}
}

// isEnterpriseBaseURL determines if a base URL represents an enterprise instance
func isEnterpriseBaseURL(providerType ProviderType, baseURL string) bool {
	baseURL = strings.ToLower(baseURL)
	
	switch providerType {
	case GitHub:
		return !strings.Contains(baseURL, "github.com")
	case GitLab:
		return !strings.Contains(baseURL, "gitlab.com")
	default:
		return false
	}
}

// ValidateFlags validates the provider flags
func (f *ProviderFlags) ValidateFlags() error {
	// Validate provider type if specified
	if f.Provider != "" {
		if _, err := ParseProviderType(f.Provider); err != nil {
			return err
		}
	}
	
	// Check for conflicting tokens
	if f.GitHubToken != "" && f.GitLabToken != "" {
		return errors.New("cannot specify both --github-token and --gitlab-token")
	}
	
	return nil
}

// HasOverrides returns true if any override flags are set
func (f *ProviderFlags) HasOverrides() bool {
	return f.Provider != "" ||
		f.GitHubToken != "" ||
		f.GitLabToken != "" ||
		f.BaseURL != ""
}
