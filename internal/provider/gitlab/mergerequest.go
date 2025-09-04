package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/aviator-co/av/internal/provider"
	"github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

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

	// Handle draft merge requests using GitLab's draft mechanism
	if opts.Draft {
		// Use the WIP: prefix for older GitLab versions compatibility, or set draft flag for newer versions
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
	
	// Handle draft status changes
	if opts.Draft != nil {
		if *opts.Draft {
			// Convert to draft by adding "Draft: " prefix if not already present
			if opts.Title != nil && !strings.HasPrefix(*opts.Title, "Draft: ") {
				updateOpts.Title = gitlab.Ptr("Draft: " + *opts.Title)
			}
		} else {
			// Convert from draft by removing "Draft: " or "WIP: " prefix
			if opts.Title != nil {
				title := *opts.Title
				title = strings.TrimPrefix(title, "Draft: ")
				title = strings.TrimPrefix(title, "WIP: ")
				updateOpts.Title = gitlab.Ptr(title)
			}
		}
	}

	mr, response, err := c.gitlab.MergeRequests.UpdateMergeRequest(projectID, mrIID, updateOpts, gitlab.WithContext(ctx))
	if err != nil {
		return nil, c.handleAPIError(err, response, "update merge request")
	}

	return c.convertMergeRequestToPullRequest(mr, projectID), nil
}

// GetPullRequests gets merge requests for a repository with pagination support
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

	// Handle GitLab pagination patterns
	listOpts := &gitlab.ListProjectMergeRequestsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 100, // GitLab's maximum per page
		},
	}

	// Map provider states to GitLab states
	if opts.State != "" {
		switch opts.State {
		case provider.PullRequestStateOpen:
			listOpts.State = gitlab.Ptr("opened")
		case provider.PullRequestStateClosed:
			listOpts.State = gitlab.Ptr("closed")
		case provider.PullRequestStateMerged:
			listOpts.State = gitlab.Ptr("merged")
		default:
			return nil, errors.Errorf("unsupported pull request state: %s", opts.State)
		}
	}

	if opts.Head != "" {
		listOpts.SourceBranch = gitlab.Ptr(opts.Head)
	}
	if opts.Base != "" {
		listOpts.TargetBranch = gitlab.Ptr(opts.Base)
	}

	var allMRs []*gitlab.MergeRequest
	page := 1

	// Handle GitLab's pagination
	for {
		listOpts.Page = page
		mrs, response, err := c.gitlab.MergeRequests.ListProjectMergeRequests(projectPath, listOpts, gitlab.WithContext(ctx))
		if err != nil {
			return nil, c.handleAPIError(err, response, "list merge requests")
		}

		allMRs = append(allMRs, mrs...)

		// Check if we have more pages
		if response.NextPage == 0 {
			break
		}
		page = response.NextPage
	}

	// Get project ID for URL generation
	project, _, err := c.gitlab.Projects.GetProject(projectPath, &gitlab.GetProjectOptions{}, gitlab.WithContext(ctx))
	if err != nil {
		return nil, c.handleAPIError(err, nil, "get project details")
	}

	result := make([]*provider.PullRequest, len(allMRs))
	for i, mr := range allMRs {
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

// RequestReviews requests reviews from the given users on the merge request
// This implements GitLab's reviewer management functionality
func (c *Client) RequestReviews(ctx context.Context, pullRequestID string, reviewerLogins []string) (*provider.PullRequest, error) {
	log := logrus.WithFields(logrus.Fields{
		"pull_request_id": pullRequestID,
		"reviewers":       reviewerLogins,
	})
	log.Debug("requesting reviews for GitLab merge request...")

	projectID, mrIID, err := parseMergeRequestID(pullRequestID)
	if err != nil {
		return nil, errors.Wrap(err, "invalid pull request ID")
	}

	// Convert usernames to user IDs
	var reviewerIDs []int
	for _, login := range reviewerLogins {
		users, response, err := c.gitlab.Users.ListUsers(&gitlab.ListUsersOptions{
			Username: gitlab.Ptr(login),
		}, gitlab.WithContext(ctx))
		if err != nil {
			return nil, c.handleAPIError(err, response, fmt.Sprintf("find user %s", login))
		}
		if len(users) == 0 {
			return nil, errors.Errorf("GitLab user %q not found", login)
		}
		reviewerIDs = append(reviewerIDs, users[0].ID)
	}

	// Update the merge request with reviewer IDs
	updateOpts := &gitlab.UpdateMergeRequestOptions{
		ReviewerIDs: &reviewerIDs,
	}

	mr, response, err := c.gitlab.MergeRequests.UpdateMergeRequest(projectID, mrIID, updateOpts, gitlab.WithContext(ctx))
	if err != nil {
		return nil, c.handleAPIError(err, response, "add reviewers to merge request")
	}

	return c.convertMergeRequestToPullRequest(mr, projectID), nil
}

// convertMergeRequestToPullRequest converts GitLab merge request to common provider format
func (c *Client) convertMergeRequestToPullRequest(mr *gitlab.MergeRequest, projectID interface{}) *provider.PullRequest {
	// Map GitLab merge request states to common provider states
	state := provider.PullRequestStateOpen
	switch mr.State {
	case "closed":
		state = provider.PullRequestStateClosed
	case "merged":
		state = provider.PullRequestStateMerged
	case "opened":
		state = provider.PullRequestStateOpen
	default:
		// Default to open for unknown states
		state = provider.PullRequestStateOpen
	}

	// Generate permalink - use GitLab's web URL if available
	permalink := mr.WebURL
	if permalink == "" && c.baseURL != "" {
		// Fallback URL generation for self-hosted instances
		permalink = fmt.Sprintf("%s/merge_requests/%d", c.baseURL, mr.IID)
	}

	// Detect if this is a draft merge request
	isDraft := mr.Draft || 
		strings.HasPrefix(strings.ToLower(mr.Title), "draft:") ||
		strings.HasPrefix(strings.ToLower(mr.Title), "wip:")

	return &provider.PullRequest{
		ID:             fmt.Sprintf("%v:%d", projectID, mr.IID),
		Number:         int64(mr.IID),
		HeadRefName:    mr.SourceBranch,
		BaseRefName:    mr.TargetBranch,
		IsDraft:        isDraft,
		Permalink:      permalink,
		State:          state,
		Title:          mr.Title,
		Body:           mr.Description,
		MergeCommitSHA: mr.MergeCommitSHA,
	}
}

// handleAPIError provides consistent error handling for GitLab API responses
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
			return errors.Wrapf(err, "GitLab API rate limit exceeded - please wait before retrying")
		case http.StatusUnprocessableEntity:
			return errors.Wrapf(err, "GitLab API request validation failed")
		case http.StatusConflict:
			return errors.Wrapf(err, "GitLab API conflict - resource may already exist")
		}
	}
	return errors.Wrapf(err, "GitLab API error during %s", operation)
}

// Utility functions for parsing GitLab-specific IDs

func parseProjectID(repositoryID string) (interface{}, error) {
	// Try parsing as integer ID first
	if id, err := strconv.Atoi(repositoryID); err == nil {
		return id, nil
	}
	// Otherwise assume it's a project path (owner/repo format)
	return repositoryID, nil
}

func parseMergeRequestID(pullRequestID string) (interface{}, int, error) {
	parts := strings.SplitN(pullRequestID, ":", 2)
	if len(parts) != 2 {
		return nil, 0, errors.Errorf("invalid merge request ID format: %s (expected format: project:iid)", pullRequestID)
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

// GitLab-specific metadata handling using similar comment patterns as GitHub

// GitLab merge request metadata constants - using HTML comments like GitHub
const (
	GitLabMRMetadataCommentStart = "<!-- av mr metadata"
	GitLabMRMetadataCommentEnd   = "-->"
	GitLabMRStackCommentStart    = "<!-- av mr stack begin -->"
	GitLabMRStackCommentEnd      = "<!-- av mr stack end -->"
	GitLabMRMetadataHelpText     = "This information is embedded by the av CLI when creating merge requests to track the status of stacks when using Aviator. Please do not delete or edit this section of the MR.\n"
)

// MRMetadata represents metadata for GitLab merge requests
type MRMetadata struct {
	Parent     string `json:"parent"`
	ParentHead string `json:"parentHead"`
	ParentMR   int64  `json:"parentMR,omitempty"`
	Trunk      string `json:"trunk"`
}

// extractContent parses the given input and looks for the start and end strings
// Same logic as GitHub implementation for consistency
func extractContent(input string, start string, end string) (content string, output string) {
	startIndex := strings.Index(input, start)
	if startIndex == -1 {
		return "", input
	}
	contentIndex := startIndex + len(start)
	endIndex := strings.Index(input[contentIndex:], end)
	if endIndex == -1 {
		return "", input
	}

	content = strings.TrimSpace(input[contentIndex : contentIndex+endIndex])
	preContent := strings.TrimSpace(input[:startIndex])
	postContent := strings.TrimSpace(input[contentIndex+endIndex+len(end):])
	output = preContent
	if postContent != "" {
		output += "\n" + postContent
	}
	return
}

// ParseMRBody parses GitLab merge request body to extract metadata
func ParseMRBody(input string) (body string, mrMeta MRMetadata, retErr error) {
	metadata, body := extractContent(input, GitLabMRMetadataCommentStart, GitLabMRMetadataCommentEnd)
	metadataContent, _ := extractContent(metadata, "```", "```")
	if metadataContent != "" {
		if err := json.Unmarshal([]byte(metadataContent), &mrMeta); err != nil {
			retErr = errors.Wrapf(err, "decoding MR metadata")
			return
		}
	}

	// Remove stack information for clean body
	_, body = extractContent(body, GitLabMRStackCommentStart, GitLabMRStackCommentEnd)

	return
}

// ReadMRMetadata extracts metadata from GitLab merge request description
func ReadMRMetadata(body string) (MRMetadata, error) {
	_, mrMeta, err := ParseMRBody(body)
	return mrMeta, err
}

// AddMRMetadataAndStack adds Aviator metadata and stack information to GitLab merge request description
// This function provides equivalent functionality to GitHub's PR metadata handling
func AddMRMetadataAndStack(body string, mrMeta MRMetadata, branchName string, hasStack bool) string {
	// Parse existing body to remove old metadata
	body, _, err := ParseMRBody(body)
	if err != nil {
		// No existing metadata comment, so continue with the full body
		logrus.WithError(err).Debug("could not parse MR metadata (assuming it doesn't exist)")
	}

	sb := strings.Builder{}

	// Add stack information if applicable (similar to GitHub implementation)
	if hasStack {
		sb.WriteString(GitLabMRStackCommentStart)
		sb.WriteString("\n")
		
		// Use GitLab-specific table formatting for consistency
		sb.WriteString("**🔗 This merge request is part of a stack created with [Aviator](https://github.com/aviator-co/av).**\n\n")
		if mrMeta.Parent != "" && mrMeta.ParentMR > 0 {
			sb.WriteString(fmt.Sprintf("⬆️ **Depends on !%d**\n\n", mrMeta.ParentMR))
		}
		sb.WriteString("_Stack will be shown here once all merge requests are created._\n\n")
		
		sb.WriteString(GitLabMRStackCommentEnd)
		sb.WriteString("\n\n")
	}

	sb.WriteString(body)

	// Add metadata section
	sb.WriteString("\n\n")
	sb.WriteString(GitLabMRMetadataCommentStart)
	sb.WriteString("\n")
	sb.WriteString(GitLabMRMetadataHelpText)
	sb.WriteString("```json\n")

	// Encode metadata as JSON
	if err := json.NewEncoder(&sb).Encode(mrMeta); err != nil {
		// shouldn't ever happen since we're encoding a simple struct to a buffer
		panic(errors.Wrapf(err, "encoding MR metadata"))
	}
	sb.WriteString("```\n")
	sb.WriteString(GitLabMRMetadataCommentEnd)
	sb.WriteString("\n")

	return sb.String()
}