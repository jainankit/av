package gitlab

import (
	"context"
	"strconv"
	"time"

	"emperror.dev/errors"
	"github.com/aviator-co/av/internal/provider"
	"github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

// User gets user information by login/username
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

// GetUserByID gets user information by numeric ID
func (c *Client) GetUserByID(ctx context.Context, id string) (*GitLabUser, error) {
	log := logrus.WithField("userID", id)
	log.Debug("getting GitLab user by ID...")

	startTime := time.Now()
	defer func() {
		log.WithField("elapsed", time.Since(startTime)).Debug("GitLab user by ID retrieval completed")
	}()

	userID, err := strconv.Atoi(id)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid user ID: %s", id)
	}

	user, response, err := c.gitlab.Users.GetUser(userID, gitlab.WithContext(ctx))
	if err != nil {
		return nil, c.handleAPIError(err, response, "get user by ID")
	}

	return &GitLabUser{
		ID:        strconv.Itoa(user.ID),
		Username:  user.Username,
		Name:      user.Name,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
		WebURL:    user.WebURL,
		State:     user.State,
	}, nil
}

// GetGroups gets groups for the authenticated user
func (c *Client) GetGroups(ctx context.Context) ([]GitLabGroup, error) {
	log := logrus.WithField("operation", "get_groups")
	log.Debug("getting GitLab groups...")

	startTime := time.Now()
	defer func() {
		log.WithField("elapsed", time.Since(startTime)).Debug("GitLab groups retrieval completed")
	}()

	groups, response, err := c.gitlab.Groups.ListGroups(&gitlab.ListGroupsOptions{
		AllAvailable: gitlab.Ptr(true),
		MinAccessLevel: gitlab.Ptr(gitlab.ReporterPermissions),
	}, gitlab.WithContext(ctx))
	if err != nil {
		return nil, c.handleAPIError(err, response, "get groups")
	}

	result := make([]GitLabGroup, len(groups))
	for i, group := range groups {
		result[i] = GitLabGroup{
			ID:                strconv.Itoa(group.ID),
			Name:              group.Name,
			Path:              group.Path,
			Description:       group.Description,
			Visibility:        string(group.Visibility),
			FullName:          group.FullName,
			FullPath:          group.FullPath,
			WebURL:            group.WebURL,
		}
	}

	return result, nil
}

// GetGroupByPath gets a specific group by its path
func (c *Client) GetGroupByPath(ctx context.Context, path string) (*GitLabGroup, error) {
	log := logrus.WithField("groupPath", path)
	log.Debug("getting GitLab group by path...")

	startTime := time.Now()
	defer func() {
		log.WithField("elapsed", time.Since(startTime)).Debug("GitLab group by path retrieval completed")
	}()

	group, response, err := c.gitlab.Groups.GetGroup(path, &gitlab.GetGroupOptions{}, gitlab.WithContext(ctx))
	if err != nil {
		return nil, c.handleAPIError(err, response, "get group by path")
	}

	return &GitLabGroup{
		ID:                strconv.Itoa(group.ID),
		Name:              group.Name,
		Path:              group.Path,
		Description:       group.Description,
		Visibility:        string(group.Visibility),
		FullName:          group.FullName,
		FullPath:          group.FullPath,
		WebURL:            group.WebURL,
	}, nil
}

// GetGroupMembers gets members of a specific group
func (c *Client) GetGroupMembers(ctx context.Context, groupID string) ([]GroupMember, error) {
	log := logrus.WithField("groupID", groupID)
	log.Debug("getting GitLab group members...")

	startTime := time.Now()
	defer func() {
		log.WithField("elapsed", time.Since(startTime)).Debug("GitLab group members retrieval completed")
	}()

	members, response, err := c.gitlab.GroupMembers.ListGroupMembers(groupID, &gitlab.ListGroupMembersOptions{}, gitlab.WithContext(ctx))
	if err != nil {
		return nil, c.handleAPIError(err, response, "get group members")
	}

	result := make([]GroupMember, len(members))
	for i, member := range members {
		result[i] = GroupMember{
			ID:          strconv.Itoa(member.ID),
			Username:    member.Username,
			Name:        member.Name,
			AccessLevel: mapGitLabAccessLevel(member.AccessLevel),
		}
	}

	return result, nil
}

// SearchUsers searches for users by query string
func (c *Client) SearchUsers(ctx context.Context, query string) ([]GitLabUser, error) {
	log := logrus.WithField("query", query)
	log.Debug("searching GitLab users...")

	startTime := time.Now()
	defer func() {
		log.WithField("elapsed", time.Since(startTime)).Debug("GitLab user search completed")
	}()

	users, response, err := c.gitlab.Users.ListUsers(&gitlab.ListUsersOptions{
		Search: gitlab.Ptr(query),
	}, gitlab.WithContext(ctx))
	if err != nil {
		return nil, c.handleAPIError(err, response, "search users")
	}

	result := make([]GitLabUser, len(users))
	for i, user := range users {
		result[i] = GitLabUser{
			ID:        strconv.Itoa(user.ID),
			Username:  user.Username,
			Name:      user.Name,
			Email:     user.Email,
			AvatarURL: user.AvatarURL,
			WebURL:    user.WebURL,
			State:     user.State,
		}
	}

	return result, nil
}

// GitLabUser represents a GitLab user with extended information
type GitLabUser struct {
	ID        string
	Username  string
	Name      string
	Email     string
	AvatarURL string
	WebURL    string
	State     string
}

// GitLabGroup represents a GitLab group/namespace
type GitLabGroup struct {
	ID                string
	Name              string
	Path              string
	Description       string
	Visibility        string
	FullName          string
	FullPath          string
	WebURL            string
}

// GroupMember represents a member of a GitLab group
type GroupMember struct {
	ID          string
	Username    string
	Name        string
	AccessLevel AccessLevel
}