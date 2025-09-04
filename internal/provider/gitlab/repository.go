package gitlab

import (
	"context"
	"strconv"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/aviator-co/av/internal/provider"
	"github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

// GetRepositoryBySlug gets repository information by owner/name slug
func (c *Client) GetRepositoryBySlug(ctx context.Context, slug string) (*provider.Repository, error) {
	log := logrus.WithField("repository", slug)
	log.Debug("getting GitLab project...")

	startTime := time.Now()
	defer func() {
		log.WithField("elapsed", time.Since(startTime)).Debug("GitLab project retrieval completed")
	}()

	// GitLab uses URL-encoded project paths (owner/name -> owner%2Fname)
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

// GetRepositoryByID gets repository information by numeric ID
func (c *Client) GetRepositoryByID(ctx context.Context, id string) (*provider.Repository, error) {
	log := logrus.WithField("projectID", id)
	log.Debug("getting GitLab project by ID...")

	startTime := time.Now()
	defer func() {
		log.WithField("elapsed", time.Since(startTime)).Debug("GitLab project retrieval by ID completed")
	}()

	project, response, err := c.gitlab.Projects.GetProject(id, &gitlab.GetProjectOptions{}, gitlab.WithContext(ctx))
	if err != nil {
		return nil, c.handleAPIError(err, response, "get project by ID")
	}

	return &provider.Repository{
		ID:    strconv.Itoa(project.ID),
		Owner: project.Namespace.Name,
		Name:  project.Name,
	}, nil
}

// GetRepositoryMembers gets project members with their access levels
func (c *Client) GetRepositoryMembers(ctx context.Context, projectID string) ([]ProjectMember, error) {
	log := logrus.WithField("projectID", projectID)
	log.Debug("getting GitLab project members...")

	startTime := time.Now()
	defer func() {
		log.WithField("elapsed", time.Since(startTime)).Debug("GitLab project members retrieval completed")
	}()

	members, response, err := c.gitlab.ProjectMembers.ListProjectMembers(projectID, &gitlab.ListProjectMembersOptions{}, gitlab.WithContext(ctx))
	if err != nil {
		return nil, c.handleAPIError(err, response, "get project members")
	}

	result := make([]ProjectMember, len(members))
	for i, member := range members {
		result[i] = ProjectMember{
			ID:          strconv.Itoa(member.ID),
			Username:    member.Username,
			Name:        member.Name,
			AccessLevel: mapGitLabAccessLevel(member.AccessLevel),
		}
	}

	return result, nil
}

// GetRepositorySettings gets detailed project settings and metadata
func (c *Client) GetRepositorySettings(ctx context.Context, projectID string) (*RepositorySettings, error) {
	log := logrus.WithField("projectID", projectID)
	log.Debug("getting GitLab project settings...")

	startTime := time.Now()
	defer func() {
		log.WithField("elapsed", time.Since(startTime)).Debug("GitLab project settings retrieval completed")
	}()

	project, response, err := c.gitlab.Projects.GetProject(projectID, &gitlab.GetProjectOptions{}, gitlab.WithContext(ctx))
	if err != nil {
		return nil, c.handleAPIError(err, response, "get project settings")
	}

	return &RepositorySettings{
		// Basic project information
		ID:                   strconv.Itoa(project.ID),
		Name:                 project.Name,
		Path:                 project.Path,
		PathWithNamespace:    project.PathWithNamespace,
		Description:          project.Description,
		DefaultBranch:        project.DefaultBranch,
		
		// Visibility and access
		Visibility:           string(project.Visibility),
		
		// URLs for different access methods
		HTTPURLToRepo:        project.HTTPURLToRepo,
		SSHURLToRepo:         project.SSHURLToRepo,
		WebURL:               project.WebURL,
		
		// Feature flags
		MergeRequestsEnabled: project.MergeRequestsEnabled,
		IssuesEnabled:        project.IssuesEnabled,
		WikiEnabled:          project.WikiEnabled,
		SnippetsEnabled:      project.SnippetsEnabled,
		JobsEnabled:          project.JobsEnabled,
		PagesEnabled:         project.PagesAccessLevel != gitlab.DisabledAccessControl,
		ContainerRegistryEnabled: project.ContainerRegistryEnabled,
		
		// Repository settings
		ArchiveOnDestroy:     project.ArchiveOnDestroy,
		RequestAccessEnabled: project.RequestAccessEnabled,
		OnlyAllowMergeIfPipelineSucceeds: project.OnlyAllowMergeIfPipelineSucceeds,
		OnlyAllowMergeIfAllDiscussionsAreResolved: project.OnlyAllowMergeIfAllDiscussionsAreResolved,
		
		// Additional metadata
		ForksCount:           project.ForksCount,
		StarsCount:           project.StarCount,
		OpenIssuesCount:      project.OpenIssuesCount,
		CreatedAt:            project.CreatedAt,
		LastActivityAt:       project.LastActivityAt,
	}, nil
}

// GenerateRepositoryPermalink creates a permalink URL for the repository
func (c *Client) GenerateRepositoryPermalink(ctx context.Context, owner, name string) (string, error) {
	urlBuilder := NewURLBuilder(c.baseURL)
	return urlBuilder.ProjectURL(owner, name), nil
}

// GenerateMergeRequestPermalink creates a permalink URL for a merge request
func (c *Client) GenerateMergeRequestPermalink(ctx context.Context, owner, name string, mrNumber int64) (string, error) {
	urlBuilder := NewURLBuilder(c.baseURL)
	return urlBuilder.MergeRequestURL(owner, name, mrNumber), nil
}

// GenerateBranchPermalink creates a permalink URL for a branch
func (c *Client) GenerateBranchPermalink(ctx context.Context, owner, name, branch string) (string, error) {
	urlBuilder := NewURLBuilder(c.baseURL)
	return urlBuilder.BranchURL(owner, name, branch), nil
}

// GenerateCommitPermalink creates a permalink URL for a commit
func (c *Client) GenerateCommitPermalink(ctx context.Context, owner, name, sha string) (string, error) {
	urlBuilder := NewURLBuilder(c.baseURL)
	return urlBuilder.CommitURL(owner, name, sha), nil
}

// ProjectMember represents a GitLab project member
type ProjectMember struct {
	ID          string
	Username    string
	Name        string
	AccessLevel AccessLevel
}

// RepositorySettings represents GitLab project settings mapped to common format
type RepositorySettings struct {
	// Basic project information
	ID                   string
	Name                 string
	Path                 string
	PathWithNamespace    string
	Description          string
	DefaultBranch        string
	
	// Visibility and access
	Visibility           string
	
	// URLs for different access methods
	HTTPURLToRepo        string
	SSHURLToRepo         string
	WebURL               string
	
	// Feature flags
	MergeRequestsEnabled bool
	IssuesEnabled        bool
	WikiEnabled          bool
	SnippetsEnabled      bool
	JobsEnabled          bool
	PagesEnabled         bool
	ContainerRegistryEnabled bool
	
	// Repository settings
	ArchiveOnDestroy     bool
	RequestAccessEnabled bool
	OnlyAllowMergeIfPipelineSucceeds bool
	OnlyAllowMergeIfAllDiscussionsAreResolved bool
	
	// Additional metadata
	ForksCount           int
	StarsCount           int
	OpenIssuesCount      int
	CreatedAt            *time.Time
	LastActivityAt       *time.Time
}

// AccessLevel represents GitLab access levels
type AccessLevel string

const (
	AccessLevelNoAccess     AccessLevel = "NO_ACCESS"
	AccessLevelMinimal      AccessLevel = "MINIMAL"
	AccessLevelGuest        AccessLevel = "GUEST"
	AccessLevelReporter     AccessLevel = "REPORTER"
	AccessLevelDeveloper    AccessLevel = "DEVELOPER"
	AccessLevelMaintainer   AccessLevel = "MAINTAINER"
	AccessLevelOwner        AccessLevel = "OWNER"
)

// mapGitLabAccessLevel maps GitLab numeric access levels to our enum
func mapGitLabAccessLevel(level gitlab.AccessLevelValue) AccessLevel {
	switch level {
	case gitlab.NoPermissions:
		return AccessLevelNoAccess
	case gitlab.MinimalAccessPermissions:
		return AccessLevelMinimal
	case gitlab.GuestPermissions:
		return AccessLevelGuest
	case gitlab.ReporterPermissions:
		return AccessLevelReporter
	case gitlab.DeveloperPermissions:
		return AccessLevelDeveloper
	case gitlab.MaintainerPermissions:
		return AccessLevelMaintainer
	case gitlab.OwnerPermissions:
		return AccessLevelOwner
	default:
		return AccessLevelNoAccess
	}
}