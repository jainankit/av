package gitlab

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/aviator-co/av/internal/provider"
	"github.com/aviator-co/av/internal/utils/logutils"
	"github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

// Client implements the GitProvider interface for GitLab
type Client struct {
	gitlab   *gitlab.Client
	baseURL  string
	token    string
}

// NewClient creates a new GitLab client
func NewClient(ctx context.Context, token, baseURL string) (*Client, error) {
	if token == "" {
		return nil, errors.New("no GitLab token provided (do you need to configure one?)")
	}

	var gitlabClient *gitlab.Client
	var err error

	if baseURL == "" {
		// Use GitLab.com
		gitlabClient, err = gitlab.NewClient(token)
	} else {
		// Use self-hosted GitLab instance
		gitlabClient, err = gitlab.NewClient(token, gitlab.WithBaseURL(baseURL))
	}

	if err != nil {
		return nil, errors.Wrap(err, "failed to create GitLab client")
	}

	// Configure rate limiting and retries
	gitlabClient.SetUserAgent("aviator-cli")

	return &Client{
		gitlab:  gitlabClient,
		baseURL: baseURL,
		token:   token,
	}, nil
}

// Pull request operations are implemented in mergerequest.go
// Repository operations are implemented in repository.go  
// User operations are implemented in user.go