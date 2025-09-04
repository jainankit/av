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

// CreatePullRequest creates a new merge request in GitLab
func (c *Client) CreatePullRequest(ctx context.Context, opts provider.CreatePullRequestOpts) (*provider.PullRequest, error) {
	log := logrus.WithFields(logrus.Fields{
		"repository_id": opts.RepositoryID,
		"head_ref":      opts.HeadRefName,
		"base_ref":      opts.BaseRefName,
		"title":         opts.Title,
		"draft":         opts.Draft,
	})
	log.Debug("creating GitLab merge request...")

	startTime := time.Now()
	defer func() {
		log.WithField("elapsed", time.Since(startTime)).Debug("GitLab merge request creation completed")
	}()

	// Parse project ID from repository ID
	projectID, err := parseProjectID(opts.RepositoryID)
	if err != nil {
		return nil, errors.Wrap(err, "invalid repository ID")
	}

	createOpts := &gitlab.CreateMergeRequestOptions{
		Title:              gitlab.Ptr(opts.Title),
		Description:        gitlab.Ptr(opts.Body),
		SourceBranch:       gitlab.Ptr(opts.HeadRefName),
		TargetBranch:       gitlab.Ptr(opts.BaseRefName),
		RemoveSourceBranch: gitlab.Ptr(false),
	}

	// Handle draft merge requests
	if opts.Draft {
		createOpts.Title = gitlab.Ptr("Draft: " + opts.Title)
	}

	mr, response, err := c.gitlab.MergeRequests.CreateMergeRequest(projectID, createOpts, gitlab.WithContext(ctx))
	if err != nil {
		return nil, c.handleAPIError(err, response, "create merge request")
	}

	return c.convertMergeRequestToPullRequest(mr, projectID), nil
}

// UpdatePullRequest updates an existing merge request
func (c *Client) UpdatePullRequest(ctx context.Context, opts provider.UpdatePullRequestOpts) (*provider.PullRequest, error) {
	log := logrus.WithFields(logrus.Fields{
		"pull_request_id": opts.PullRequestID,
	})
	log.Debug("updating GitLab merge request...")

	startTime := time.Now()
	defer func() {
		log.WithField("elapsed", time.Since(startTime)).Debug("GitLab merge request update completed")
	}()

	projectID, mrIID, err := parseMergeRequestID(opts.PullRequestID)
	if err != nil {
		return nil, errors.Wrap(err, "invalid pull request ID")
	}

	updateOpts := &gitlab.UpdateMergeRequestOptions{}
	if opts.Title != nil {
		updateOpts.Title = opts.Title
	}
	if opts.Body != nil {
		updateOpts.Description = opts.Body
	}

	mr, response, err := c.gitlab.MergeRequests.UpdateMergeRequest(projectID, mrIID, updateOpts, gitlab.WithContext(ctx))
	if err != nil {
		return nil, c.handleAPIError(err, response, "update merge request")
	}

	return c.convertMergeRequestToPullRequest(mr, projectID), nil
}

// GetPullRequests gets merge requests for a repository
func (c *Client) GetPullRequests(ctx context.Context, repoSlug string, opts provider.GetPullRequestsOpts) ([]*provider.PullRequest, error) {
	log := logrus.WithFields(logrus.Fields{
		"repository": repoSlug,
		"head":       opts.Head,
		"base":       opts.Base,
		"state":      opts.State,
	})
	log.Debug("getting GitLab merge requests...")

	startTime := time.Now()
	defer func() {
		log.WithField("elapsed", time.Since(startTime)).Debug("GitLab merge requests retrieval completed")
	}()

	projectPath := strings.ReplaceAll(repoSlug, "/", "%2F")

	listOpts := &gitlab.ListProjectMergeRequestsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
		},
	}

	// Map states from provider to GitLab
	if opts.State != "" {
		switch opts.State {
		case provider.PullRequestStateOpen:
			listOpts.State = gitlab.Ptr("opened")
		case provider.PullRequestStateClosed:
			listOpts.State = gitlab.Ptr("closed")
		case provider.PullRequestStateMerged:
			listOpts.State = gitlab.Ptr("merged")
		}
	}

	if opts.Head != "" {
		listOpts.SourceBranch = gitlab.Ptr(opts.Head)
	}
	if opts.Base != "" {
		listOpts.TargetBranch = gitlab.Ptr(opts.Base)
	}

	mrs, response, err := c.gitlab.MergeRequests.ListProjectMergeRequests(projectPath, listOpts, gitlab.WithContext(ctx))
	if err != nil {
		return nil, c.handleAPIError(err, response, "list merge requests")
	}

	// Get project ID for URL generation
	project, _, err := c.gitlab.Projects.GetProject(projectPath, &gitlab.GetProjectOptions{}, gitlab.WithContext(ctx))
	if err != nil {
		return nil, c.handleAPIError(err, response, "get project details")
	}

	result := make([]*provider.PullRequest, len(mrs))
	for i, mr := range mrs {
		result[i] = c.convertMergeRequestToPullRequest(mr, project.ID)
	}

	return result, nil
}

// PullRequest gets a specific merge request by ID
func (c *Client) PullRequest(ctx context.Context, id string) (*provider.PullRequest, error) {
	log := logrus.WithField("pull_request_id", id)
	log.Debug("getting GitLab merge request...")

	startTime := time.Now()
	defer func() {
		log.WithField("elapsed", time.Since(startTime)).Debug("GitLab merge request retrieval completed")
	}()

	projectID, mrIID, err := parseMergeRequestID(id)
	if err != nil {
		return nil, errors.Wrap(err, "invalid pull request ID")
	}

	mr, response, err := c.gitlab.MergeRequests.GetMergeRequest(projectID, mrIID, &gitlab.GetMergeRequestsOptions{}, gitlab.WithContext(ctx))
	if err != nil {
		return nil, c.handleAPIError(err, response, "get merge request")
	}

	return c.convertMergeRequestToPullRequest(mr, projectID), nil
}

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

// Helper functions

func (c *Client) convertMergeRequestToPullRequest(mr *gitlab.MergeRequest, projectID interface{}) *provider.PullRequest {
	state := provider.PullRequestStateOpen
	switch mr.State {
	case "closed":
		state = provider.PullRequestStateClosed
	case "merged":
		state = provider.PullRequestStateMerged
	}

	// Generate permalink
	permalink := mr.WebURL
	if permalink == "" && c.baseURL != "" {
		// Fallback URL generation for self-hosted instances
		permalink = fmt.Sprintf("%s/merge_requests/%d", c.baseURL, mr.IID)
	}

	return &provider.PullRequest{
		ID:             fmt.Sprintf("%v:%d", projectID, mr.IID),
		Number:         int64(mr.IID),
		HeadRefName:    mr.SourceBranch,
		BaseRefName:    mr.TargetBranch,
		IsDraft:        mr.Draft,
		Permalink:      permalink,
		State:          state,
		Title:          mr.Title,
		Body:           mr.Description,
		MergeCommitSHA: mr.MergeCommitSHA,
	}
}

func (c *Client) handleAPIError(err error, response *gitlab.Response, operation string) error {
	if response != nil && response.Response != nil {
		switch response.StatusCode {
		case http.StatusUnauthorized:
			return errors.Wrapf(err, "GitLab authentication failed - check your token")
		case http.StatusForbidden:
			return errors.Wrapf(err, "GitLab API access forbidden - check your permissions")
		case http.StatusNotFound:
			return errors.Wrapf(err, "GitLab resource not found")
		case http.StatusTooManyRequests:
			return errors.Wrapf(err, "GitLab API rate limit exceeded")
		}
	}
	return errors.Wrapf(err, "GitLab API error during %s", operation)
}

func parseProjectID(repositoryID string) (interface{}, error) {
	// Try parsing as integer ID first
	if id, err := strconv.Atoi(repositoryID); err == nil {
		return id, nil
	}
	// Otherwise assume it's a project path
	return repositoryID, nil
}

func parseMergeRequestID(pullRequestID string) (interface{}, int, error) {
	parts := strings.SplitN(pullRequestID, ":", 2)
	if len(parts) != 2 {
		return nil, 0, errors.Errorf("invalid merge request ID format: %s", pullRequestID)
	}

	projectID, err := parseProjectID(parts[0])
	if err != nil {
		return nil, 0, errors.Wrap(err, "invalid project ID in merge request ID")
	}

	mrIID, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, 0, errors.Wrap(err, "invalid merge request IID")
	}

	return projectID, mrIID, nil
}