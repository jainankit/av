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

// Pull request operations are now implemented in mergerequest.go

// GetRepositoryBySlug gets repository information by owner/name slug
func (c *Client) GetRepositoryBySlug(ctx context.Context, slug string) (*provider.Repository, error) {
	log := logrus.WithField("repository", slug)
	log.Debug("getting GitLab project...")

	startTime := time.Now()
	defer func() {
		log.WithField("elapsed", time.Since(startTime)).Debug("GitLab project retrieval completed")
	}()

	projectPath := strings.ReplaceAll(slug, "/", "%2F")
	project, response, err := c.gitlab.Projects.GetProject(projectPath, &gitlab.GetProjectOptions{}, gitlab.WithContext(ctx))
	if err != nil {
		return nil, c.handleAPIError(err, response, "get project")
	}

	return &provider.Repository{
		ID:    strconv.Itoa(project.ID),
		Owner: project.Namespace.Name,
		Name:  project.Name,
	}, nil
}

// User gets user information by login
func (c *Client) User(ctx context.Context, login string) (*provider.User, error) {
	log := logrus.WithField("login", login)
	log.Debug("getting GitLab user...")

	startTime := time.Now()
	defer func() {
		log.WithField("elapsed", time.Since(startTime)).Debug("GitLab user retrieval completed")
	}()

	users, response, err := c.gitlab.Users.ListUsers(&gitlab.ListUsersOptions{
		Username: gitlab.Ptr(login),
	}, gitlab.WithContext(ctx))
	if err != nil {
		return nil, c.handleAPIError(err, response, "get user")
	}

	if len(users) == 0 {
		return nil, errors.Errorf("GitLab user %q not found", login)
	}

	user := users[0]
	return &provider.User{
		ID:    strconv.Itoa(user.ID),
		Login: user.Username,
	}, nil
}

// Viewer gets the authenticated user information
func (c *Client) Viewer(ctx context.Context) (*provider.Viewer, error) {
	log := logrus.WithField("operation", "viewer")
	log.Debug("getting GitLab current user...")

	startTime := time.Now()
	defer func() {
		log.WithField("elapsed", time.Since(startTime)).Debug("GitLab current user retrieval completed")
	}()

	user, response, err := c.gitlab.Users.CurrentUser(gitlab.WithContext(ctx))
	if err != nil {
		return nil, c.handleAPIError(err, response, "get current user")
	}

	return &provider.Viewer{
		Name:  user.Name,
		Login: user.Username,
	}, nil
}

// Helper functions for user and repository operations are in their respective files