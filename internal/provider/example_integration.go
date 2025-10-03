package provider

import (
	"context"
	"fmt"

	"github.com/aviator-co/av/internal/git"
	"github.com/spf13/cobra"
)

// ExampleCommand demonstrates how to integrate the provider package
// into an av command. This is a reference implementation showing
// best practices.
//
// NOTE: This file is for documentation purposes and should not be
// included in the final build.
func ExampleCommand() *cobra.Command {
	var flags ProviderFlags
	
	cmd := &cobra.Command{
		Use:   "example",
		Short: "Example command showing provider integration",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			
			// Validate command-line flags
			if err := flags.ValidateFlags(); err != nil {
				return fmt.Errorf("invalid flags: %w", err)
			}
			
			// Get the git repository
			repo, err := git.OpenRepo(".")
			if err != nil {
				return fmt.Errorf("failed to open git repository: %w", err)
			}
			
			// Convert flags to config overrides
			overrides, err := ConfigFromFlags(&flags)
			if err != nil {
				return fmt.Errorf("failed to process flags: %w", err)
			}
			
			// Detect the provider
			provider, err := DetectProviderWithConfig(ctx, repo, overrides)
			if err != nil {
				return fmt.Errorf("failed to detect provider: %w", err)
			}
			
			// Validate the provider
			result := ValidateProviderComprehensive(ctx, provider)
			if !result.IsValid {
				fmt.Printf("Provider validation warnings:\n")
				for _, warning := range result.Warnings {
					fmt.Printf("  - %s\n", warning)
				}
				
				if len(result.Errors) > 0 {
					fmt.Printf("Provider validation errors:\n")
					for _, err := range result.Errors {
						fmt.Printf("  - %v\n", err)
					}
					return fmt.Errorf("provider validation failed")
				}
			}
			
			// Use the provider
			return useProvider(ctx, provider)
		},
	}
	
	// Add provider flags to the command
	AddProviderFlags(cmd, &flags)
	
	return cmd
}

// useProvider demonstrates how to use the detected provider
func useProvider(ctx context.Context, provider *Provider) error {
	fmt.Printf("Using %s provider\n", provider.Type)
	fmt.Printf("Base URL: %s\n", provider.BaseURL)
	fmt.Printf("Repository: %s\n", provider.RepoSlug)
	fmt.Printf("Enterprise: %v\n", provider.IsEnterprise)
	
	switch provider.Type {
	case GitHub:
		return useGitHubProvider(ctx, provider)
	case GitLab:
		return useGitLabProvider(ctx, provider)
	default:
		return fmt.Errorf("unsupported provider: %s", provider.Type)
	}
}

// useGitHubProvider shows how to use the GitHub provider
func useGitHubProvider(ctx context.Context, provider *Provider) error {
	fmt.Printf("Creating GitHub client...\n")
	
	// TODO: Create GitHub client using provider.Token and provider.BaseURL
	// client, err := gh.NewClient(ctx, provider.Token)
	// if err != nil {
	//     return fmt.Errorf("failed to create GitHub client: %w", err)
	// }
	
	// TODO: Use the client for GitHub operations
	// user, err := client.GetAuthenticatedUser(ctx)
	
	fmt.Printf("GitHub provider ready\n")
	return nil
}

// useGitLabProvider shows how to use the GitLab provider
func useGitLabProvider(ctx context.Context, provider *Provider) error {
	fmt.Printf("Creating GitLab client...\n")
	
	// TODO: Create GitLab client using provider.Token and provider.BaseURL
	// client, err := gitlab.NewClient(ctx, provider.Token, provider.BaseURL)
	// if err != nil {
	//     return fmt.Errorf("failed to create GitLab client: %w", err)
	// }
	
	// TODO: Use the client for GitLab operations
	// user, err := client.GetCurrentUser(ctx)
	
	fmt.Printf("GitLab provider ready\n")
	return nil
}

// ExampleDetectionInExistingCommand shows how to add provider detection
// to an existing command that currently only supports GitHub
func ExampleDetectionInExistingCommand(existingCmd *cobra.Command) {
	var providerFlags ProviderFlags
	
	// Add provider flags to existing command
	AddProviderFlags(existingCmd, &providerFlags)
	
	// Store the original RunE function
	originalRunE := existingCmd.RunE
	
	// Wrap the original RunE with provider detection
	existingCmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		
		// Detect provider before running original logic
		repo, err := git.OpenRepo(".")
		if err != nil {
			// If not in a git repo, fall back to original behavior
			return originalRunE(cmd, args)
		}
		
		overrides, err := ConfigFromFlags(&providerFlags)
		if err != nil {
			return err
		}
		
		provider, err := DetectProviderWithConfig(ctx, repo, overrides)
		if err != nil {
			// Fall back to original behavior if detection fails
			return originalRunE(cmd, args)
		}
		
		// Store provider in context for use by the original command
		ctx = context.WithValue(ctx, "provider", provider)
		cmd.SetContext(ctx)
		
		return originalRunE(cmd, args)
	}
}

// GetProviderFromContext retrieves the provider from context
// This is a helper function for commands that use the wrapped approach
func GetProviderFromContext(ctx context.Context) (*Provider, bool) {
	provider, ok := ctx.Value("provider").(*Provider)
	return provider, ok
}

// ExampleGlobalProviderSetup shows how to set up provider detection
// at the root command level for use by all subcommands
func ExampleGlobalProviderSetup() *cobra.Command {
	var globalFlags ProviderFlags
	
	rootCmd := &cobra.Command{
		Use: "av",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			
			// Only detect provider for commands that need it
			// (check if command has provider-related flags or is in a git repo)
			if !needsProvider(cmd) {
				return nil
			}
			
			repo, err := git.OpenRepo(".")
			if err != nil {
				// Not in a git repo, skip provider detection
				return nil
			}
			
			overrides, err := ConfigFromFlags(&globalFlags)
			if err != nil {
				return err
			}
			
			provider, err := DetectProviderWithConfig(ctx, repo, overrides)
			if err != nil {
				// Provider detection failed, continue without provider
				// Commands can handle this case individually
				return nil
			}
			
			// Store provider in context for subcommands
			ctx = context.WithValue(ctx, "provider", provider)
			cmd.SetContext(ctx)
			
			return nil
		},
	}
	
	// Add global provider flags
	AddProviderFlagsToFlagSet(rootCmd.PersistentFlags(), &globalFlags)
	
	return rootCmd
}

// needsProvider determines if a command needs provider detection
func needsProvider(cmd *cobra.Command) bool {
	// Commands that work with remote repositories need provider detection
	needsProviderCommands := map[string]bool{
		"pr":     true,
		"sync":   true,
		"fetch":  true,
		"init":   true,
		"auth":   true,
	}
	
	return needsProviderCommands[cmd.Name()]
}

// ExampleConfigurationUsage shows how to use configuration-based provider selection
func ExampleConfigurationUsage() error {
	ctx := context.Background()
	
	// Get provider from configuration alone (no git repo needed)
	provider, err := GetProviderFromConfig()
	if err != nil {
		return fmt.Errorf("no provider configured: %w", err)
	}
	
	// Validate the configured provider
	result := ValidateProviderComprehensive(ctx, provider)
	if !result.IsValid {
		return fmt.Errorf("configured provider is invalid")
	}
	
	fmt.Printf("Using configured %s provider\n", provider.Type)
	return nil
}
