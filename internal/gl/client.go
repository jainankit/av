package gl

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
	gl         *githubv4.Client
}

// NewClient creates a new GitLab client.
// It takes configuration from the global config.Av.GitLab variable.
func NewClient(ctx context.Context, token string) (*Client, error) {
	if token == "" {
		return nil, errors.Errorf("no GitLab token provided (do you need to configure one?)")
	}
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(ctx, src)
	var gl *githubv4.Client
	if config.Av.GitLab.BaseURL == "" {
		// Default to GitLab.com GraphQL API endpoint
		gl = githubv4.NewEnterpriseClient("https://gitlab.com/api/graphql", httpClient)
	} else {
		// Self-hosted GitLab instance
		gl = githubv4.NewEnterpriseClient(config.Av.GitLab.BaseURL+"/api/graphql", httpClient)
	}
	return &Client{httpClient, gl}, nil
}

func (c *Client) query(ctx context.Context, query any, variables map[string]any) (reterr error) {
	log := logrus.WithFields(logrus.Fields{
		"variables": logutils.Format("%#+v", variables),
	})
	log.Debug("executing GitLab API query...")
	startTime := time.Now()
	defer func() {
		log := log.WithFields(logrus.Fields{
			"elapsed": time.Since(startTime),
			"result":  logutils.Format("%#+v", query),
		})
		if reterr != nil {
			log.WithError(reterr).Debug("GitLab API query failed")
		} else {
			log.Debug("GitLab API query succeeded")
		}
	}()
	return c.gl.Query(ctx, query, variables)
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
	log.Debug("executing GitLab API mutation...")
	startTime := time.Now()
	defer func() {
		log := log.WithFields(logrus.Fields{
			"elapsed": time.Since(startTime),
			"result":  logutils.Format("%#+v", mutation),
		})
		if reterr != nil {
			log.WithError(reterr).Debug("GitLab API mutation failed")
		} else {
			log.Debug("GitLab API mutation succeeded")
		}
	}()
	return c.gl.Mutate(ctx, mutation, input, variables)
}