package gitlab

import (
	"context"
	"net/http"

	"emperror.dev/errors"
	"github.com/shurcooL/graphql"
	"golang.org/x/oauth2"
)

type Client struct {
	httpClient *http.Client
	gql        *graphql.Client
}

// NewClient creates a new GitLab client.
// It takes configuration from the global config.Av.GitLab variable.
func NewClient(ctx context.Context, token string, baseURL string) (*Client, error) {
	if token == "" {
		return nil, errors.Errorf("no GitLab token provided (do you need to configure one?)")
	}
	
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(ctx, src)
	
	// Default to GitLab.com if no base URL provided
	graphqlEndpoint := "https://gitlab.com/api/graphql"
	if baseURL != "" {
		// For self-hosted GitLab instances, construct the GraphQL endpoint
		graphqlEndpoint = baseURL + "/api/graphql"
	}
	
	gql := graphql.NewClient(graphqlEndpoint, httpClient)
	
	return &Client{httpClient, gql}, nil
}

func (c *Client) query(ctx context.Context, query any, variables map[string]any) error {
	return c.gql.Query(ctx, query, variables)
}

func (c *Client) mutate(
	ctx context.Context,
	mutation any,
	variables map[string]any,
) error {
	return c.gql.Mutate(ctx, mutation, variables)
}