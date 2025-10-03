package main

import (
	"context"
	"fmt"
	"os"

	"emperror.dev/errors"
	"github.com/spf13/cobra"

	"github.com/aviator-co/av/internal/avgql"
	"github.com/aviator-co/av/internal/gh"
	"github.com/aviator-co/av/internal/gitlab"
	"github.com/aviator-co/av/internal/provider"
	"github.com/aviator-co/av/internal/utils/colors"
)

var authCmd = &cobra.Command{
	Use:          "auth",
	Short:        "Check user authentication status",
	SilenceUsage: true,
	Args:         cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		if err := checkAviatorAuthStatus(cmd.Context()); err != nil {
			fmt.Fprintln(os.Stderr, colors.Warning(err.Error()))
		}
		if err := checkProviderAuthStatus(cmd.Context()); err != nil {
			fmt.Fprintln(os.Stderr, colors.Failure(err.Error()))
		}
	},
}

func checkAviatorAuthStatus(ctx context.Context) error {
	avClient, err := avgql.NewClient(ctx)
	if err != nil {
		return err
	}

	var query struct{ avgql.ViewerSubquery }
	if err := avClient.Query(ctx, &query, nil); err != nil {
		if avgql.IsHTTPUnauthorized(err) {
			return errors.New(
				"You are not logged in to Aviator. Please verify that your API token is correct.",
			)
		}
		return errors.Wrap(err, "Failed to query Aviator")
	}

	fmt.Fprint(os.Stderr,
		"Logged in to Aviator as ", colors.UserInput(query.Viewer.FullName),
		" (", colors.UserInput(query.Viewer.Email), ").\n",
	)
	return nil
}

func checkGitHubAuthStatus(ctx context.Context) error {
	ghClient, err := getGitHubClient(ctx)
	if err != nil {
		return err
	}

	viewer, err := ghClient.Viewer(ctx)
	if err != nil {
		// GitHub API returns 401 Unauthorized if the token is invalid or
		// expired.
		if gh.IsHTTPUnauthorized(err) {
			return errors.New(
				"You are not logged in to GitHub. Please verify that your API token is correct.",
			)
		}
		return errors.Wrap(err, "Failed to query GitHub")
	}

	fmt.Fprint(os.Stderr,
		"Logged in to GitHub as ", colors.UserInput(viewer.Name),
		" (", colors.UserInput(viewer.Login), ").\n",
	)
	return nil
}

func checkGitLabAuthStatus(ctx context.Context) error {
	glClient, err := getGitLabClient(ctx)
	if err != nil {
		return err
	}

	viewer, err := glClient.Viewer(ctx)
	if err != nil {
		// GitLab API returns 401 Unauthorized if the token is invalid or
		// expired.
		if gitlab.IsHTTPUnauthorized(err) {
			return errors.New(
				"You are not logged in to GitLab. Please verify that your API token is correct.",
			)
		}
		return errors.Wrap(err, "Failed to query GitLab")
	}

	fmt.Fprint(os.Stderr,
		"Logged in to GitLab as ", colors.UserInput(viewer.Name),
		" (", colors.UserInput(viewer.Username), ").\n",
	)
	return nil
}

// checkProviderAuthStatus checks authentication status for the detected provider
func checkProviderAuthStatus(ctx context.Context) error {
	repo, err := getRepo(ctx)
	if err != nil {
		// If we can't detect the repo, fall back to checking both providers
		return checkBothProvidersAuthStatus(ctx)
	}
	
	detectedProvider, err := provider.DetectProvider(ctx, repo)
	if err != nil {
		// If we can't detect the provider, fall back to checking both
		return checkBothProvidersAuthStatus(ctx)
	}
	
	switch detectedProvider.Type {
	case provider.GitHub:
		return checkGitHubAuthStatus(ctx)
	case provider.GitLab:
		return checkGitLabAuthStatus(ctx)
	default:
		return checkBothProvidersAuthStatus(ctx)
	}
}

// checkBothProvidersAuthStatus checks both GitHub and GitLab when provider is ambiguous
func checkBothProvidersAuthStatus(ctx context.Context) error {
	githubErr := checkGitHubAuthStatus(ctx)
	gitlabErr := checkGitLabAuthStatus(ctx)
	
	// If both fail, return the more relevant error
	if githubErr != nil && gitlabErr != nil {
		// Check if either token is configured to provide better error message
		if discoverGitHubAPIToken(ctx) != "" {
			return githubErr
		}
		if discoverGitLabAPIToken(ctx) != "" {
			return gitlabErr
		}
		// If neither token is found, suggest configuring one
		return errors.New(
			"No GitHub or GitLab token found. Please configure authentication for your provider.",
		)
	}
	
	// At least one succeeded, so return nil
	return nil
}

func init() {
	// deprecated 'av auth status', hidden to avoid it showing up in 'av auth --help'
	// since that is the new command name
	deprecatedAuthStatus := deprecateCommand(*authCmd, "av auth", "status")
	deprecatedAuthStatus.Hidden = true

	authCmd.AddCommand(
		deprecatedAuthStatus,
	)
}
