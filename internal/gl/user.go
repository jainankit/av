package gl

import (
	"context"
	"fmt"
	"time"

	"emperror.dev/errors"
)

type State string

const (
	StateActive   State = "active"
	StateBlocked  State = "blocked"
	StateDeactivated State = "deactivated"
)

type User struct {
	ID               int       `json:"id"`
	Username         string    `json:"username"`
	Name             string    `json:"name"`
	Email            string    `json:"email"`
	State            State     `json:"state"`
	Locked           bool      `json:"locked"`
	AvatarURL        string    `json:"avatar_url"`
	WebURL           string    `json:"web_url"`
	CreatedAt        time.Time `json:"created_at"`
	Bio              string    `json:"bio"`
	Location         string    `json:"location"`
	PublicEmail      string    `json:"public_email"`
	Skype            string    `json:"skype"`
	LinkedIn         string    `json:"linkedin"`
	Twitter          string    `json:"twitter"`
	Discord          string    `json:"discord"`
	WebsiteURL       string    `json:"website_url"`
	Organization     string    `json:"organization"`
	JobTitle         string    `json:"job_title"`
	Pronouns         string    `json:"pronouns"`
	Bot              bool      `json:"bot"`
	WorkInformation  string    `json:"work_information"`
	Followers        int       `json:"followers"`
	Following        int       `json:"following"`
	LocalTime        string    `json:"local_time"`
	LastSignInAt     *time.Time `json:"last_sign_in_at"`
	ConfirmedAt      *time.Time `json:"confirmed_at"`
	LastActivityOn   string    `json:"last_activity_on"`
	ColorSchemeID    int       `json:"color_scheme_id"`
	ProjectsLimit    int       `json:"projects_limit"`
	CurrentSignInAt  *time.Time `json:"current_sign_in_at"`
	Identities       []Identity `json:"identities"`
	CanCreateGroup   bool      `json:"can_create_group"`
	CanCreateProject bool      `json:"can_create_project"`
	TwoFactorEnabled bool      `json:"two_factor_enabled"`
	External         bool      `json:"external"`
	PrivateProfile   bool      `json:"private_profile"`
	CommitEmail      string    `json:"commit_email"`
}

type Identity struct {
	Provider  string `json:"provider"`
	ExternUID string `json:"extern_uid"`
}

func (c *Client) GetUser(ctx context.Context, userID int) (*User, error) {
	path := fmt.Sprintf("/users/%d", userID)
	
	var user User
	if err := c.get(ctx, path, nil, &user); err != nil {
		return nil, errors.Wrapf(err, "failed to get user %d", userID)
	}
	
	return &user, nil
}

func (c *Client) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	path := "/users"
	params := map[string]string{
		"username": username,
	}
	
	var users []User
	if err := c.get(ctx, path, params, &users); err != nil {
		return nil, errors.Wrapf(err, "failed to search for user %s", username)
	}
	
	if len(users) == 0 {
		return nil, errors.Errorf("GitLab user %q not found", username)
	}
	
	// Return the first user found
	return &users[0], nil
}

func (c *Client) GetCurrentUser(ctx context.Context) (*User, error) {
	path := "/user"
	
	var user User
	if err := c.get(ctx, path, nil, &user); err != nil {
		return nil, errors.Wrap(err, "failed to get current user")
	}
	
	return &user, nil
}

// User returns information about the given user by username.
// This method maintains compatibility with the GitHub client interface.
func (c *Client) User(ctx context.Context, username string) (*User, error) {
	return c.GetUserByUsername(ctx, username)
}