package gh

import (
	"context"
	"net/http"
	"time"

	"emperror.dev/errors"
	"github.com/aviator-co/av/internal/config"
	"github.com/aviator-co/av/internal/utils/logutils"
	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type Client struct {
	httpClient *http.Client
	gh         *githubv4.Client
}

// NewClient creates a new GitHub client.
// It takes configuration from the global config.Av.GitHub variable.
func NewClient(ctx context.Context, token string) (*Client, error) {
	return NewClientWithBaseURL(ctx, token, config.Av.GitHub.BaseURL)
}

// NewClientWithBaseURL creates a new GitHub client with a specific base URL.
// This function is compatible with provider factory instantiation.
// If baseURL is empty, it uses the public GitHub API.
func NewClientWithBaseURL(ctx context.Context, token, baseURL string) (*Client, error) {
	if token == "" {
		return nil, errors.Errorf("no GitHub token provided (do you need to configure one?)")
	}
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(ctx, src)
	var gh *githubv4.Client
	if baseURL == "" {
		gh = githubv4.NewClient(httpClient)
	} else {
		gh = githubv4.NewEnterpriseClient(baseURL+"/api/graphql", httpClient)
	}
	return &Client{httpClient, gh}, nil
}

func (c *Client) query(ctx context.Context, query any, variables map[string]any) (reterr error) {
	log := logrus.WithFields(logrus.Fields{
		"variables": logutils.Format("%#+v", variables),
	})
	log.Debug("executing GitHub API query...")
	startTime := time.Now()
	defer func() {
		log := log.WithFields(logrus.Fields{
			"elapsed": time.Since(startTime),
			"result":  logutils.Format("%#+v", query),
		})
		if reterr != nil {
			log.WithError(reterr).Debug("GitHub API query failed")
		} else {
			log.Debug("GitHub API query succeeded")
		}
	}()
	return c.gh.Query(ctx, query, variables)
}

func (c *Client) mutate(
	ctx context.Context,
	mutation any,
	input githubv4.Input,
	variables map[string]any,
) (reterr error) {
	log := logrus.WithFields(logrus.Fields{
		"input": logutils.Format("%#+v", input),
	})
	log.Debug("executing GitHub API mutation...")
	startTime := time.Now()
	defer func() {
		log := log.WithFields(logrus.Fields{
			"elapsed": time.Since(startTime),
			"result":  logutils.Format("%#+v", mutation),
		})
		if reterr != nil {
			log.WithError(reterr).Debug("GitHub API mutation failed")
		} else {
			log.Debug("GitHub API mutation succeeded")
		}
	}()
	return c.gh.Mutate(ctx, mutation, input, variables)
}

// HTTPClient returns the underlying HTTP client for advanced usage
func (c *Client) HTTPClient() *http.Client {
	return c.httpClient
}

// GraphQLClient returns the underlying GitHub GraphQL client for direct access
func (c *Client) GraphQLClient() *githubv4.Client {
	return c.gh
}
