package gitlab

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitLabUserType(t *testing.T) {
	user := GitLabUser{
		ID:        "123",
		Username:  "testuser",
		Name:      "Test User",
		Email:     "test@example.com",
		AvatarURL: "https://gitlab.com/uploads/-/avatar/123.jpg",
		WebURL:    "https://gitlab.com/testuser",
		State:     "active",
	}

	assert.Equal(t, "123", user.ID)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "Test User", user.Name)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "https://gitlab.com/uploads/-/avatar/123.jpg", user.AvatarURL)
	assert.Equal(t, "https://gitlab.com/testuser", user.WebURL)
	assert.Equal(t, "active", user.State)
}

func TestGitLabGroupType(t *testing.T) {
	group := GitLabGroup{
		ID:          "456", 
		Name:        "Test Group",
		Path:        "test-group",
		Description: "A test group for testing",
		Visibility:  "public",
		FullName:    "Test Group",
		FullPath:    "test-group",
		WebURL:      "https://gitlab.com/groups/test-group",
	}

	assert.Equal(t, "456", group.ID)
	assert.Equal(t, "Test Group", group.Name)
	assert.Equal(t, "test-group", group.Path)
	assert.Equal(t, "A test group for testing", group.Description)
	assert.Equal(t, "public", group.Visibility)
	assert.Equal(t, "Test Group", group.FullName)
	assert.Equal(t, "test-group", group.FullPath)
	assert.Equal(t, "https://gitlab.com/groups/test-group", group.WebURL)
}

func TestGroupMemberType(t *testing.T) {
	member := GroupMember{
		ID:          "789",
		Username:    "groupuser",
		Name:        "Group User",
		AccessLevel: AccessLevelMaintainer,
	}

	assert.Equal(t, "789", member.ID)
	assert.Equal(t, "groupuser", member.Username) 
	assert.Equal(t, "Group User", member.Name)
	assert.Equal(t, AccessLevelMaintainer, member.AccessLevel)
}

func TestUserSearchResults(t *testing.T) {
	// Test that user search results can be handled properly
	users := []GitLabUser{
		{
			ID:        "1",
			Username:  "alice",
			Name:      "Alice Smith",
			Email:     "alice@example.com",
			State:     "active",
		},
		{
			ID:        "2", 
			Username:  "bob",
			Name:      "Bob Jones",
			Email:     "bob@example.com",
			State:     "active",
		},
	}

	assert.Len(t, users, 2)
	assert.Equal(t, "alice", users[0].Username)
	assert.Equal(t, "bob", users[1].Username)
}

func TestGroupMembersCollection(t *testing.T) {
	// Test that group members can be collected and managed
	members := []GroupMember{
		{
			ID:          "1",
			Username:    "owner",
			Name:        "Owner User",
			AccessLevel: AccessLevelOwner,
		},
		{
			ID:          "2",
			Username:    "maintainer",
			Name:        "Maintainer User", 
			AccessLevel: AccessLevelMaintainer,
		},
		{
			ID:          "3",
			Username:    "developer",
			Name:        "Developer User",
			AccessLevel: AccessLevelDeveloper,
		},
	}

	assert.Len(t, members, 3)
	
	// Verify different access levels
	assert.Equal(t, AccessLevelOwner, members[0].AccessLevel)
	assert.Equal(t, AccessLevelMaintainer, members[1].AccessLevel)
	assert.Equal(t, AccessLevelDeveloper, members[2].AccessLevel)
	
	// Verify usernames
	assert.Equal(t, "owner", members[0].Username)
	assert.Equal(t, "maintainer", members[1].Username)
	assert.Equal(t, "developer", members[2].Username)
}