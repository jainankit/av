package gl

import (
	"context"
	"net/http"
	"time"

	"emperror.dev/errors"
	"github.com/aviator-co/av/internal/config"
	"github.com/aviator-co/av/internal/utils/logutils"
	"github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

type Client struct {
	httpClient *http.Client
	gl         *gitlab.Client
}

// NewClient creates a new GitLab client.
// It takes configuration from the global config.Av.GitLab variable.
func NewClient(ctx context.Context, token string) (*Client, error) {
	if token == "" {
		return nil, errors.Errorf("no GitLab token provided (do you need to configure one?)")
	}
	
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	var gl *gitlab.Client
	var err error
	if config.Av.GitLab.BaseURL == "" {
		// Use GitLab.com
		gl, err = gitlab.NewClient(token, gitlab.WithHTTPClient(httpClient))
	} else {
		// Use self-hosted GitLab instance
		gl, err = gitlab.NewClient(token, gitlab.WithBaseURL(config.Av.GitLab.BaseURL), gitlab.WithHTTPClient(httpClient))
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to create GitLab client")
	}
	
	return &Client{httpClient, gl}, nil
}

// request executes a GitLab API request with logging, similar to GitHub's query method.
// This provides a consistent interface for making GitLab API calls with proper error handling and logging.
func (c *Client) request(ctx context.Context, operation string, fn func() error) (reterr error) {
	log := logrus.WithField("operation", operation)
	log.Debug("executing GitLab API operation...")
	startTime := time.Now()
	defer func() {
		log := log.WithField("elapsed", time.Since(startTime))
		if reterr != nil {
			log.WithError(reterr).Debug("GitLab API operation failed")
		} else {
			log.Debug("GitLab API operation succeeded")
		}
	}()
	
	return fn()
}

// GetMergeRequest fetches a merge request by project ID and merge request IID.
// This follows the pattern of GitHub's query helper methods.
func (c *Client) GetMergeRequest(ctx context.Context, projectID interface{}, mergeRequestIID int) (*gitlab.MergeRequest, error) {
	var mr *gitlab.MergeRequest
	var err error
	
	err = c.request(ctx, "GetMergeRequest", func() error {
		mr, _, err = c.gl.MergeRequests.GetMergeRequest(projectID, mergeRequestIID, nil)
		return err
	})
	
	return mr, err
}

// ListMergeRequests lists merge requests for a project with optional filtering.
// This provides a consistent interface similar to GitHub's pull request queries.
func (c *Client) ListMergeRequests(ctx context.Context, projectID interface{}, opts *gitlab.ListProjectMergeRequestsOptions) ([]*gitlab.MergeRequest, error) {
	var mrs []*gitlab.MergeRequest
	var err error
	
	err = c.request(ctx, "ListMergeRequests", func() error {
		mrs, _, err = c.gl.MergeRequests.ListProjectMergeRequests(projectID, opts)
		return err
	})
	
	return mrs, err
}

// CreateMergeRequest creates a new merge request.
// This follows the mutation pattern similar to GitHub's client.
func (c *Client) CreateMergeRequest(ctx context.Context, projectID interface{}, opts *gitlab.CreateMergeRequestOptions) (*gitlab.MergeRequest, error) {
	var mr *gitlab.MergeRequest
	var err error
	
	err = c.request(ctx, "CreateMergeRequest", func() error {
		mr, _, err = c.gl.MergeRequests.CreateMergeRequest(projectID, opts)
		return err
	})
	
	return mr, err
}

// UpdateMergeRequest updates an existing merge request.
// This provides mutation functionality similar to GitHub's mutate method.
func (c *Client) UpdateMergeRequest(ctx context.Context, projectID interface{}, mergeRequestIID int, opts *gitlab.UpdateMergeRequestOptions) (*gitlab.MergeRequest, error) {
	var mr *gitlab.MergeRequest
	var err error
	
	err = c.request(ctx, "UpdateMergeRequest", func() error {
		mr, _, err = c.gl.MergeRequests.UpdateMergeRequest(projectID, mergeRequestIID, opts)
		return err
	})
	
	return mr, err
}

// GetProject fetches a project by ID or path.
// This provides repository-like functionality similar to GitHub's repository operations.
func (c *Client) GetProject(ctx context.Context, projectID interface{}) (*gitlab.Project, error) {
	var project *gitlab.Project
	var err error
	
	err = c.request(ctx, "GetProject", func() error {
		project, _, err = c.gl.Projects.GetProject(projectID, nil)
		return err
	})
	
	return project, err
}

// GetCurrentUser fetches the current authenticated user.
// This follows the user pattern similar to GitHub's user operations.
func (c *Client) GetCurrentUser(ctx context.Context) (*gitlab.User, error) {
	var user *gitlab.User
	var err error
	
	err = c.request(ctx, "GetCurrentUser", func() error {
		user, _, err = c.gl.Users.CurrentUser()
		return err
	})
	
	return user, err
}