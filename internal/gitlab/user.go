package gitlab

import (
	"context"
	"net/url"
	"strconv"

	"emperror.dev/errors"
)

// User represents a GitLab user
type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	State    string `json:"state"`
	WebURL   string `json:"web_url"`
}

// GetLogin returns the username (equivalent to GitHub's login)
func (u *User) GetLogin() string {
	return u.Username
}

// GetID returns the user ID as a string
func (u *User) GetID() string {
	return strconv.FormatInt(u.ID, 10)
}

// GetUser retrieves a user by username
func (c *Client) GetUser(ctx context.Context, username string) (*User, error) {
	// First try to get user by username
	users, err := c.SearchUsers(ctx, username, 1)
	if err != nil {
		return nil, errors.Wrap(err, "failed to search for user")
	}

	// Find exact username match
	for _, user := range users {
		if user.Username == username {
			return &user, nil
		}
	}

	return nil, errors.Errorf("GitLab user %q not found", username)
}

// GetUserByID retrieves a user by their numeric ID
func (c *Client) GetUserByID(ctx context.Context, userID int64) (*User, error) {
	endpoint := "/users/" + strconv.FormatInt(userID, 10)
	
	var result User
	if err := c.get(ctx, endpoint, &result); err != nil {
		if IsHTTPNotFound(err) {
			return nil, errors.Errorf("user with ID %d not found", userID)
		}
		return nil, errors.Wrap(err, "failed to get user")
	}

	return &result, nil
}

// SearchUsers searches for users by username or name
func (c *Client) SearchUsers(ctx context.Context, search string, perPage int) ([]User, error) {
	endpoint := "/users"
	
	// Build query parameters
	params := make(map[string]string)
	if search != "" {
		params["search"] = search
	}
	if perPage > 0 {
		params["per_page"] = strconv.Itoa(perPage)
	} else {
		params["per_page"] = "20" // Default
	}

	// Build the URL with query parameters
	fullEndpoint := c.buildURL(endpoint, params)

	var users []User
	if err := c.get(ctx, fullEndpoint, &users); err != nil {
		return nil, errors.Wrap(err, "failed to search users")
	}

	return users, nil
}

// GetProjectMembers retrieves members of a project (for reviewer assignment)
func (c *Client) GetProjectMembers(ctx context.Context, projectID string) ([]User, error) {
	endpoint := "/projects/" + url.PathEscape(projectID) + "/members"
	
	// Build query parameters for pagination
	params := map[string]string{
		"per_page": "100", // Get up to 100 members
	}

	fullEndpoint := c.buildURL(endpoint, params)

	var members []struct {
		User `json:",inline"`
		// Additional member fields if needed
		AccessLevel int `json:"access_level"`
	}
	
	if err := c.get(ctx, fullEndpoint, &members); err != nil {
		return nil, errors.Wrap(err, "failed to get project members")
	}

	// Convert to User slice
	users := make([]User, len(members))
	for i, member := range members {
		users[i] = member.User
	}

	return users, nil
}

// ValidateUserExists checks if a user exists and is active
func (c *Client) ValidateUserExists(ctx context.Context, username string) error {
	user, err := c.GetUser(ctx, username)
	if err != nil {
		return err
	}

	// Check if user is active (not blocked/deactivated)
	if user.State != "active" {
		return errors.Errorf("user %q is not active (state: %s)", username, user.State)
	}

	return nil
}

// GetCurrentUser retrieves the currently authenticated user
func (c *Client) GetCurrentUser(ctx context.Context) (*User, error) {
	endpoint := "/user"
	
	var result User
	if err := c.get(ctx, endpoint, &result); err != nil {
		return nil, errors.Wrap(err, "failed to get current user")
	}

	return &result, nil
}