package github

import (
	"context"
	"net/http"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/aviator-co/av/internal/config"
	"github.com/aviator-co/av/internal/provider"
	"github.com/aviator-co/av/internal/utils/logutils"
	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// Client implements the GitProvider interface for GitHub
type Client struct {
	httpClient *http.Client
	gh         *githubv4.Client
}

// NewClient creates a new GitHub provider client
func NewClient(ctx context.Context, token string) (*Client, error) {
	if token == "" {
		return nil, errors.Errorf("no GitHub token provided (do you need to configure one?)")
	}
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(ctx, src)
	var gh *githubv4.Client
	if config.Av.GitHub.BaseURL == "" {
		gh = githubv4.NewClient(httpClient)
	} else {
		gh = githubv4.NewEnterpriseClient(config.Av.GitHub.BaseURL+"/api/graphql", httpClient)
	}
	return &Client{httpClient, gh}, nil
}

// Type returns the provider type
func (c *Client) Type() provider.ProviderType {
	return provider.GitHub
}

// CreatePullRequest creates a new pull request
func (c *Client) CreatePullRequest(ctx context.Context, input provider.CreatePullRequestInput) (*provider.PullRequest, error) {
	ghInput := githubv4.CreatePullRequestInput{
		RepositoryID: githubv4.ID(input.RepositoryID),
		Title:        githubv4.String(input.Title),
		Body:         nullable(githubv4.String(input.Body)),
		HeadRefName:  githubv4.String(input.HeadRefName),
		BaseRefName:  githubv4.String(input.BaseRefName),
		Draft:        nullable(githubv4.Boolean(input.IsDraft)),
	}

	var mutation struct {
		CreatePullRequest struct {
			PullRequest githubPullRequest
		} `graphql:"createPullRequest(input: $input)"`
	}
	if err := c.mutate(ctx, &mutation, ghInput, nil); err != nil {
		return nil, errors.Wrap(err, "failed to create pull request: github error")
	}
	
	return convertGithubPRToProvider(&mutation.CreatePullRequest.PullRequest), nil
}

// UpdatePullRequest updates an existing pull request
func (c *Client) UpdatePullRequest(ctx context.Context, input provider.UpdatePullRequestInput) (*provider.PullRequest, error) {
	ghInput := githubv4.UpdatePullRequestInput{
		PullRequestID: githubv4.ID(input.PullRequestID),
	}
	
	if input.Title != nil {
		ghInput.Title = nullable(githubv4.String(*input.Title))
	}
	if input.Body != nil {
		ghInput.Body = nullable(githubv4.String(*input.Body))
	}
	if input.IsDraft != nil {
		// GitHub API doesn't support updating draft status directly via UpdatePullRequest
		// This would need to be handled separately via ConvertPullRequestToDraft/MarkPullRequestReadyForReview
		// For now, we'll ignore this field and handle it in a wrapper function if needed
	}

	var mutation struct {
		UpdatePullRequest struct {
			PullRequest githubPullRequest
		} `graphql:"updatePullRequest(input: $input)"`
	}
	if err := c.mutate(ctx, &mutation, ghInput, nil); err != nil {
		return nil, errors.Wrap(err, "failed to update pull request: github error")
	}
	
	return convertGithubPRToProvider(&mutation.UpdatePullRequest.PullRequest), nil
}

// GetPullRequests retrieves pull requests matching the given criteria
func (c *Client) GetPullRequests(ctx context.Context, input provider.GetPullRequestsInput) (*provider.GetPullRequestsPage, error) {
	if input.First == 0 {
		input.First = 50
	}

	// Convert provider states to GitHub states
	var ghStates []githubv4.PullRequestState
	for _, state := range input.States {
		ghState, err := convertProviderStateToGithub(state)
		if err != nil {
			return nil, err
		}
		ghStates = append(ghStates, ghState)
	}

	var query struct {
		Repository struct {
			PullRequests struct {
				Nodes    []githubPullRequest
				PageInfo githubPageInfo
			} `graphql:"pullRequests(states: $states, headRefName: $headRefName, baseRefName: $baseRefName, first: $first, after: $after)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}
	
	if err := c.query(ctx, &query, map[string]interface{}{
		"owner":       githubv4.String(input.Owner),
		"repo":        githubv4.String(input.Repo),
		"headRefName": nullable(githubv4.String(input.HeadRefName)),
		"baseRefName": nullable(githubv4.String(input.BaseRefName)),
		"states":      &ghStates,
		"first":       githubv4.Int(input.First),
		"after":       nullable(githubv4.String(input.After)),
	}); err != nil {
		return nil, errors.Wrap(err, "failed to query pull requests")
	}

	// Convert response to provider types
	var providerPRs []provider.PullRequest
	for _, ghPR := range query.Repository.PullRequests.Nodes {
		providerPRs = append(providerPRs, *convertGithubPRToProvider(&ghPR))
	}

	return &provider.GetPullRequestsPage{
		PageInfo:     convertGithubPageInfoToProvider(&query.Repository.PullRequests.PageInfo),
		PullRequests: providerPRs,
	}, nil
}

// GetRepositoryBySlug retrieves a repository by its owner/name slug
func (c *Client) GetRepositoryBySlug(ctx context.Context, slug string) (*provider.Repository, error) {
	owner, name, ok := strings.Cut(slug, "/")
	if !ok {
		return nil, errors.Errorf(
			"unable to parse repository slug (expected <owner>/<repo>): %q",
			slug,
		)
	}

	var query struct {
		Repository githubRepository `graphql:"repository(owner: $owner, name: $name)"`
	}
	err := c.query(ctx, &query, map[string]interface{}{
		"owner": githubv4.String(owner),
		"name":  githubv4.String(name),
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch repository from GitHub")
	}

	return &provider.Repository{
		ID:    query.Repository.ID,
		Owner: query.Repository.Owner.Login,
		Name:  query.Repository.Name,
	}, nil
}

// User retrieves information about a user by login
func (c *Client) User(ctx context.Context, login string) (*provider.User, error) {
	var query struct {
		User githubUser `graphql:"user(login: $login)"`
	}
	if err := c.query(ctx, &query, map[string]any{
		"login": githubv4.String(login),
	}); err != nil {
		return nil, err
	}
	if query.User.ID == "" {
		return nil, errors.Errorf("GitHub user %q not found", login)
	}
	return &provider.User{
		ID:    string(query.User.ID),
		Login: query.User.Login,
	}, nil
}

// Viewer retrieves information about the authenticated user
func (c *Client) Viewer(ctx context.Context) (*provider.Viewer, error) {
	var query struct {
		Viewer githubViewer `graphql:"viewer"`
	}
	err := c.query(ctx, &query, nil)
	if err != nil {
		return nil, err
	}
	return &provider.Viewer{
		Name:  query.Viewer.Name,
		Login: query.Viewer.Login,
	}, nil
}

// Internal types that match the original GitHub GraphQL structure
type githubPullRequest struct {
	ID                  string
	Number              int64
	HeadRefName         string
	BaseRefName         string
	IsDraft             bool
	Permalink           string
	State               githubv4.PullRequestState
	Title               string
	Body                string
	PRIVATE_MergeCommit struct {
		Oid string
	} `graphql:"mergeCommit"`
	PRIVATE_TimelineItems struct {
		Nodes []struct {
			ClosedEvent struct {
				Closer struct {
					Commit struct {
						Oid string
					} `graphql:"... on Commit"`
				}
			} `graphql:"... on ClosedEvent"`
			MergedEvent struct {
				Commit struct {
					Oid string
				}
			} `graphql:"... on MergedEvent"`
		}
	} `graphql:"timelineItems(last: 10, itemTypes: [CLOSED_EVENT, MERGED_EVENT])"`
}

type githubRepository struct {
	ID    string
	Owner struct {
		Login string
	}
	Name string
}

type githubUser struct {
	ID    githubv4.ID `graphql:"id"`
	Login string      `graphql:"login"`
}

type githubViewer struct {
	Name  string `graphql:"name"`
	Login string `graphql:"login"`
}

type githubPageInfo struct {
	EndCursor       string
	HasNextPage     bool
	HasPreviousPage bool
	StartCursor     string
}

// Helper functions for conversion between GitHub and provider types
func convertGithubPRToProvider(ghPR *githubPullRequest) *provider.PullRequest {
	return &provider.PullRequest{
		ID:          ghPR.ID,
		Number:      ghPR.Number,
		HeadRefName: ghPR.HeadRefName,
		BaseRefName: ghPR.BaseRefName,
		IsDraft:     ghPR.IsDraft,
		Permalink:   ghPR.Permalink,
		State:       convertGithubStateToProvider(ghPR.State),
		Title:       ghPR.Title,
		Body:        ghPR.Body,
		MergeCommit: getMergeCommitFromGithubPR(ghPR),
	}
}

func convertGithubStateToProvider(ghState githubv4.PullRequestState) provider.PullRequestState {
	switch ghState {
	case githubv4.PullRequestStateOpen:
		return provider.PullRequestStateOpen
	case githubv4.PullRequestStateClosed:
		return provider.PullRequestStateClosed
	case githubv4.PullRequestStateMerged:
		return provider.PullRequestStateMerged
	default:
		return provider.PullRequestStateOpen // default fallback
	}
}

func convertProviderStateToGithub(providerState provider.PullRequestState) (githubv4.PullRequestState, error) {
	switch providerState {
	case provider.PullRequestStateOpen:
		return githubv4.PullRequestStateOpen, nil
	case provider.PullRequestStateClosed:
		return githubv4.PullRequestStateClosed, nil
	case provider.PullRequestStateMerged:
		return githubv4.PullRequestStateMerged, nil
	default:
		return "", errors.Errorf("unknown provider state: %s", providerState)
	}
}

func convertGithubPageInfoToProvider(ghPageInfo *githubPageInfo) provider.PageInfo {
	return provider.PageInfo{
		HasNextPage:     ghPageInfo.HasNextPage,
		HasPreviousPage: ghPageInfo.HasPreviousPage,
		StartCursor:     ghPageInfo.StartCursor,
		EndCursor:       ghPageInfo.EndCursor,
	}
}

func getMergeCommitFromGithubPR(p *githubPullRequest) string {
	if p.State == githubv4.PullRequestStateOpen {
		return ""
	} else if p.State == githubv4.PullRequestStateMerged && p.PRIVATE_MergeCommit.Oid != "" {
		return p.PRIVATE_MergeCommit.Oid
	}
	// The timeline is in chronological order, so we can iterate backwards to find the latest one.
	for i := len(p.PRIVATE_TimelineItems.Nodes) - 1; i >= 0; i-- {
		item := p.PRIVATE_TimelineItems.Nodes[i]
		if item.ClosedEvent.Closer.Commit.Oid != "" {
			return item.ClosedEvent.Closer.Commit.Oid
		}
		if item.MergedEvent.Commit.Oid != "" {
			return item.MergedEvent.Commit.Oid
		}
	}
	return ""
}

// Internal helper methods for GraphQL operations
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

// nullable returns a pointer to the argument if it's not the zero value,
// otherwise it returns nil.
// This is useful to translate between Golang-style "unset is zero" and GraphQL
// which distinguishes between unset (null) and zero values.
func nullable[T comparable](v T) *T {
	var zero T
	if v == zero {
		return nil
	}
	return &v
}