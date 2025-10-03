package provider

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/sirupsen/logrus"
)

// ValidationResult contains the result of provider validation
type ValidationResult struct {
	// IsValid indicates if the provider configuration is valid
	IsValid bool

	// Errors contains any validation errors found
	Errors []error

	// Warnings contains any validation warnings
	Warnings []string

	// AuthenticationStatus indicates the authentication status
	AuthenticationStatus AuthStatus

	// ConnectivityStatus indicates the connectivity status
	ConnectivityStatus ConnectivityStatus

	// UserInfo contains authenticated user information (if available)
	UserInfo *ValidationUserInfo

	// RepositoryAccess indicates if the repository is accessible
	RepositoryAccess *RepositoryAccessInfo
}

// AuthStatus represents authentication status
type AuthStatus string

const (
	AuthStatusValid   AuthStatus = "valid"
	AuthStatusInvalid AuthStatus = "invalid"
	AuthStatusMissing AuthStatus = "missing"
	AuthStatusUnknown AuthStatus = "unknown"
)

// ConnectivityStatus represents connectivity status
type ConnectivityStatus string

const (
	ConnectivityStatusConnected ConnectivityStatus = "connected"
	ConnectivityStatusFailed    ConnectivityStatus = "failed"
	ConnectivityStatusUnknown   ConnectivityStatus = "unknown"
)

// ValidationUserInfo contains user information from validation
type ValidationUserInfo struct {
	Username string
	Name     string
	Email    string
}

// RepositoryAccessInfo contains repository access information
type RepositoryAccessInfo struct {
	HasReadAccess  bool
	HasWriteAccess bool
	HasAdminAccess bool
	RepositoryExists bool
}

// ValidateProviderComprehensive performs comprehensive validation of a provider
func ValidateProviderComprehensive(ctx context.Context, provider *Provider) *ValidationResult {
	result := &ValidationResult{
		AuthenticationStatus: AuthStatusUnknown,
		ConnectivityStatus:   ConnectivityStatusUnknown,
		Errors:               []error{},
		Warnings:             []string{},
	}
	
	if provider == nil {
		result.Errors = append(result.Errors, errors.New("provider is nil"))
		return result
	}
	
	logrus.WithFields(logrus.Fields{
		"provider":   provider.Type,
		"base_url":   provider.BaseURL,
		"enterprise": provider.IsEnterprise,
	}).Debug("starting comprehensive provider validation")

	// Basic validation
	if err := validateBasicProvider(provider); err != nil {
		result.Errors = append(result.Errors, err)
	}

	// Authentication validation
	authStatus, authErr, userInfo := validateAuthentication(ctx, provider)
	result.AuthenticationStatus = authStatus
	result.UserInfo = userInfo
	if authErr != nil {
		result.Errors = append(result.Errors, authErr)
	}

	// Connectivity validation (only if authentication is available)
	if result.AuthenticationStatus == AuthStatusValid ||
		result.AuthenticationStatus == AuthStatusUnknown {
		connectivityStatus, connectivityErr := validateConnectivity(ctx, provider)
		result.ConnectivityStatus = connectivityStatus
		if connectivityErr != nil {
			result.Errors = append(result.Errors, connectivityErr)
		}
	}

	// Repository access validation (if repo slug is available)
	if provider.RepoSlug != "" && result.AuthenticationStatus == AuthStatusValid {
		repoAccess, repoErr := validateRepositoryAccess(ctx, provider)
		result.RepositoryAccess = repoAccess
		if repoErr != nil {
			// Repository access errors are warnings, not hard errors
			result.Warnings = append(result.Warnings, repoErr.Error())
		}
	}

	// Add warnings for enterprise instances
	if provider.IsEnterprise {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Using enterprise instance at %s - some features may behave differently",
				provider.BaseURL))
	}

	// Determine overall validity
	result.IsValid = len(result.Errors) == 0 &&
		result.AuthenticationStatus != AuthStatusInvalid &&
		result.ConnectivityStatus != ConnectivityStatusFailed
	
	return result
}

// validateBasicProvider validates basic provider configuration
func validateBasicProvider(provider *Provider) error {
	if provider.Type == Unknown {
		return errors.New("provider type is unknown")
	}
	
	if provider.Type != GitHub && provider.Type != GitLab {
		return errors.Errorf("unsupported provider type: %s", provider.Type)
	}
	
	if provider.BaseURL == "" {
		return errors.New("provider base URL is empty")
	}
	
	// Validate base URL format
	if !strings.HasPrefix(provider.BaseURL, "https://") &&
		!strings.HasPrefix(provider.BaseURL, "http://") {
		return errors.Errorf(
			"provider base URL must start with https:// or http://: %s",
			provider.BaseURL,
		)
	}

	return nil
}

// validateAuthentication validates provider authentication
func validateAuthentication(
	ctx context.Context,
	provider *Provider,
) (AuthStatus, error, *ValidationUserInfo) {
	if provider.Token == "" {
		// Try to discover token from external sources
		token := discoverTokenForProvider(ctx, provider)
		if token == "" {
			return AuthStatusMissing, errors.New("no authentication token found"), nil
		}
		provider.Token = token
	}

	// Validate token format
	if err := validateTokenFormat(provider); err != nil {
		return AuthStatusInvalid, err, nil
	}

	// Try to get user information to validate token
	userInfo, err := getUserInfo(ctx, provider)
	if err != nil {
		return AuthStatusInvalid, errors.Wrap(err, "failed to validate authentication token"), nil
	}

	return AuthStatusValid, nil, userInfo
}

// validateTokenFormat validates the format of authentication tokens
func validateTokenFormat(provider *Provider) error {
	token := provider.Token
	
	switch provider.Type {
	case GitHub:
		// GitHub personal access tokens start with "ghp_", "gho_", "ghu_", or "ghs_" for newer tokens
		// Classic tokens are 40 character hex strings
		if len(token) == 40 {
			// Classic token format (hex string)
			for _, char := range token {
				if !((char >= '0' && char <= '9') ||
					(char >= 'a' && char <= 'f') ||
					(char >= 'A' && char <= 'F')) {
					return errors.New(
						"invalid GitHub token format: classic tokens must be 40-character hex strings",
					)
				}
			}
		} else if strings.HasPrefix(token, "ghp_") ||
			strings.HasPrefix(token, "gho_") ||
			strings.HasPrefix(token, "ghu_") ||
			strings.HasPrefix(token, "ghs_") {
			// New token format is valid
		} else {
			return errors.New(
				"invalid GitHub token format: must be a 40-character hex string or start with ghp_, gho_, ghu_, or ghs_",
			)
		}
		
	case GitLab:
		// GitLab personal access tokens start with "glpat_" for newer tokens
		// Project access tokens start with "glpat_"
		// Deploy tokens start with different prefixes
		if strings.HasPrefix(token, "glpat_") {
			// New token format is valid
		} else if len(token) >= 20 && len(token) <= 64 {
			// Legacy token format (alphanumeric string)
			for _, char := range token {
				if !((char >= '0' && char <= '9') ||
					(char >= 'a' && char <= 'z') ||
					(char >= 'A' && char <= 'Z') ||
					char == '_' || char == '-') {
					return errors.New(
						"invalid GitLab token format: contains invalid characters",
					)
				}
			}
		} else {
			return errors.New(
				"invalid GitLab token format: must start with glpat_ or be a 20-64 character alphanumeric string",
			)
		}
	}
	
	return nil
}

// discoverTokenForProvider discovers authentication token for a provider
func discoverTokenForProvider(ctx context.Context, provider *Provider) string {
	switch provider.Type {
	case GitHub:
		return discoverGitHubToken(ctx)
	case GitLab:
		return discoverGitLabToken(ctx)
	default:
		return ""
	}
}

// validateConnectivity validates connectivity to the provider API
func validateConnectivity(ctx context.Context, provider *Provider) (ConnectivityStatus, error) {
	// Create a context with timeout for connectivity check
	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	
	// TODO: Implement actual API connectivity test
	// This would make a simple API call to test connectivity
	// For now, we'll just validate that the base URL is reachable
	
	logrus.WithField("base_url", provider.BaseURL).Debug(
		"connectivity validation not fully implemented",
	)
	
	// Basic URL validation was already done in validateBasicProvider
	return ConnectivityStatusUnknown, nil
}

// getUserInfo attempts to get user information from the provider
func getUserInfo(ctx context.Context, provider *Provider) (*ValidationUserInfo, error) {
	// TODO: Implement actual user info retrieval using provider clients
	// This would create the appropriate client and make an API call
	
	logrus.WithField("provider", provider.Type).Debug("user info validation not fully implemented")
	
	// For now, return a placeholder result
	return &ValidationUserInfo{
		Username: "unknown",
		Name:     "Unknown User",
		Email:    "",
	}, nil
}

// validateRepositoryAccess validates access to the specified repository
func validateRepositoryAccess(
	ctx context.Context,
	provider *Provider,
) (*RepositoryAccessInfo, error) {
	// TODO: Implement actual repository access validation
	// This would check if the repository exists and what permissions the user has
	
	logrus.WithFields(logrus.Fields{
		"provider": provider.Type,
		"repo_slug": provider.RepoSlug,
	}).Debug("repository access validation not fully implemented")
	
	// For now, return a placeholder result
	return &RepositoryAccessInfo{
		HasReadAccess:    true,
		HasWriteAccess:   false,
		HasAdminAccess:   false,
		RepositoryExists: true,
	}, nil
}

// Enhanced CLI token discovery with better error handling
func tryGitHubCLIEnhanced(ctx context.Context) (string, error) {
	ghCli, err := exec.LookPath("gh")
	if err != nil {
		return "", errors.New("GitHub CLI (gh) not found in PATH")
	}
	
	// Create a context with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(cmdCtx, ghCli, "auth", "status", "--show-token")
	output, err := cmd.Output()
	if err != nil {
		// Try the alternative command
		cmd = exec.CommandContext(cmdCtx, ghCli, "auth", "token")
		output, err = cmd.Output()
		if err != nil {
			return "", errors.Wrap(err, "failed to get GitHub token from CLI")
		}
	}
	
	token := strings.TrimSpace(string(output))
	if token == "" {
		return "", errors.New("GitHub CLI returned empty token")
	}
	
	return token, nil
}

// tryGitLabCLIEnhanced attempts to get GitLab token from the GitLab CLI with better error handling
func tryGitLabCLIEnhanced(ctx context.Context) (string, error) {
	glabCli, err := exec.LookPath("glab")
	if err != nil {
		return "", errors.New("GitLab CLI (glab) not found in PATH")
	}
	
	// Create a context with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(cmdCtx, glabCli, "auth", "print-access-token")
	output, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, "failed to get GitLab token from CLI")
	}
	
	token := strings.TrimSpace(string(output))
	if token == "" {
		return "", errors.New("GitLab CLI returned empty token")
	}
	
	return token, nil
}
