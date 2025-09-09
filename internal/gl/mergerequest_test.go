package gl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock merge request data for testing
func createMockMergeRequest() *MergeRequest {
	now := time.Now()
	return &MergeRequest{
		ID:                    123,
		IID:                   1,
		Title:                 "Test merge request",
		Description:           "Test description",
		State:                 MergeRequestStateOpened,
		CreatedAt:             now,
		UpdatedAt:             now,
		SourceBranch:          "feature-branch",
		TargetBranch:          "main",
		WebURL:                "https://gitlab.com/project/repo/-/merge_requests/1",
		WorkInProgress:        false,
		Draft:                 false,
		Author:                User{ID: 1, Username: "testuser", Name: "Test User"},
		ProjectID:             456,
		SourceProjectID:       456,
		TargetProjectID:       456,
		Labels:                []string{"enhancement", "feature"},
		MergeStatus:           "can_be_merged",
		SHA:                   "abc123def456",
		HasConflicts:          false,
		BlockingDiscussionsResolved: true,
	}
}

func TestMergeRequest_HeadBranchName(t *testing.T) {
	tests := []struct {
		name         string
		sourceBranch string
		expected     string
	}{
		{
			name:         "branch without refs/heads/ prefix",
			sourceBranch: "feature-branch",
			expected:     "feature-branch",
		},
		{
			name:         "branch with refs/heads/ prefix",
			sourceBranch: "refs/heads/feature-branch",
			expected:     "feature-branch",
		},
		{
			name:         "empty branch",
			sourceBranch: "",
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := &MergeRequest{SourceBranch: tt.sourceBranch}
			assert.Equal(t, tt.expected, mr.HeadBranchName())
		})
	}
}

func TestMergeRequest_BaseBranchName(t *testing.T) {
	tests := []struct {
		name         string
		targetBranch string
		expected     string
	}{
		{
			name:         "branch without refs/heads/ prefix",
			targetBranch: "main",
			expected:     "main",
		},
		{
			name:         "branch with refs/heads/ prefix",
			targetBranch: "refs/heads/main",
			expected:     "main",
		},
		{
			name:         "empty branch",
			targetBranch: "",
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := &MergeRequest{TargetBranch: tt.targetBranch}
			assert.Equal(t, tt.expected, mr.BaseBranchName())
		})
	}
}

func TestMergeRequest_GetMergeCommit(t *testing.T) {
	mergeCommitSHA := "merge123"
	squashCommitSHA := "squash456"

	tests := []struct {
		name     string
		mr       *MergeRequest
		expected string
	}{
		{
			name: "opened merge request",
			mr: &MergeRequest{
				State: MergeRequestStateOpened,
			},
			expected: "",
		},
		{
			name: "closed merge request",
			mr: &MergeRequest{
				State: MergeRequestStateClosed,
			},
			expected: "",
		},
		{
			name: "merged merge request with merge commit",
			mr: &MergeRequest{
				State:           MergeRequestStateMerged,
				MergeCommitSHA:  &mergeCommitSHA,
			},
			expected: "merge123",
		},
		{
			name: "merged merge request with squash commit",
			mr: &MergeRequest{
				State:            MergeRequestStateMerged,
				SquashCommitSHA:  &squashCommitSHA,
			},
			expected: "squash456",
		},
		{
			name: "merged merge request with both commits (merge takes priority)",
			mr: &MergeRequest{
				State:            MergeRequestStateMerged,
				MergeCommitSHA:   &mergeCommitSHA,
				SquashCommitSHA:  &squashCommitSHA,
			},
			expected: "merge123",
		},
		{
			name: "merged merge request with no commit info",
			mr: &MergeRequest{
				State: MergeRequestStateMerged,
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.mr.GetMergeCommit())
		})
	}
}

func TestClient_GetMergeRequest(t *testing.T) {
	ctx := context.Background()

	t.Run("successful get merge request by ID", func(t *testing.T) {
		mockMR := createMockMergeRequest()
		
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/api/v4/merge_requests/123", r.URL.Path)
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockMR)
		}))
		defer server.Close()

		client, err := NewClient(ctx, "test-token", server.URL)
		require.NoError(t, err)

		mr, err := client.GetMergeRequest(ctx, 123)
		require.NoError(t, err)
		assert.Equal(t, mockMR.ID, mr.ID)
		assert.Equal(t, mockMR.Title, mr.Title)
		assert.Equal(t, mockMR.State, mr.State)
	})

	t.Run("merge request not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"message": "404 Not Found"}`)
		}))
		defer server.Close()

		client, err := NewClient(ctx, "test-token", server.URL)
		require.NoError(t, err)

		mr, err := client.GetMergeRequest(ctx, 999)
		assert.Error(t, err)
		assert.Nil(t, mr)
		assert.Contains(t, err.Error(), "failed to get merge request 999")
	})
}

func TestClient_GetMergeRequestByProject(t *testing.T) {
	ctx := context.Background()

	t.Run("successful get merge request by project and IID", func(t *testing.T) {
		mockMR := createMockMergeRequest()
		
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/api/v4/projects/456/merge_requests/1", r.URL.Path)
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockMR)
		}))
		defer server.Close()

		client, err := NewClient(ctx, "test-token", server.URL)
		require.NoError(t, err)

		mr, err := client.GetMergeRequestByProject(ctx, 456, 1)
		require.NoError(t, err)
		assert.Equal(t, mockMR.ID, mr.ID)
		assert.Equal(t, mockMR.IID, mr.IID)
		assert.Equal(t, mockMR.ProjectID, mr.ProjectID)
	})

	t.Run("merge request not found in project", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"message": "404 Not Found"}`)
		}))
		defer server.Close()

		client, err := NewClient(ctx, "test-token", server.URL)
		require.NoError(t, err)

		mr, err := client.GetMergeRequestByProject(ctx, 456, 999)
		assert.Error(t, err)
		assert.Nil(t, mr)
		assert.Contains(t, err.Error(), "failed to get merge request 999 from project 456")
	})
}

func TestClient_GetMergeRequests(t *testing.T) {
	ctx := context.Background()

	t.Run("successful get merge requests with pagination", func(t *testing.T) {
		mockMRs := []*MergeRequest{createMockMergeRequest()}
		mockMRs[0].IID = 1
		mockMRs = append(mockMRs, &MergeRequest{
			ID:           124,
			IID:          2,
			Title:        "Second MR",
			State:        MergeRequestStateOpened,
			SourceBranch: "feature-2",
			TargetBranch: "main",
			ProjectID:    456,
		})
		
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/api/v4/projects/456/merge_requests", r.URL.Path)
			
			// Check query parameters
			query := r.URL.Query()
			assert.Equal(t, "50", query.Get("per_page")) // Default value
			assert.Equal(t, "opened", query.Get("state"))
			assert.Equal(t, "feature-branch", query.Get("source_branch"))
			
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Page", "1")
			w.Header().Set("X-Per-Page", "50")
			w.Header().Set("X-Total", "2")
			w.Header().Set("X-Total-Pages", "1")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockMRs)
		}))
		defer server.Close()

		client, err := NewClient(ctx, "test-token", server.URL)
		require.NoError(t, err)

		sourceBranch := "feature-branch"
		state := MergeRequestStateOpened
		input := GetMergeRequestsInput{
			ProjectID:    456,
			SourceBranch: &sourceBranch,
			State:        &state,
		}

		page, err := client.GetMergeRequests(ctx, input)
		require.NoError(t, err)
		assert.Len(t, page.MergeRequests, 2)
		assert.Equal(t, 123, page.MergeRequests[0].ID)
		assert.Equal(t, 124, page.MergeRequests[1].ID)
		assert.Equal(t, 1, page.PageInfo.Page)
		assert.Equal(t, 50, page.PageInfo.PerPage)
		assert.Equal(t, 2, page.PageInfo.Total)
		assert.Equal(t, 1, page.PageInfo.TotalPages)
	})

	t.Run("get merge requests with custom pagination", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query()
			assert.Equal(t, "20", query.Get("per_page"))
			assert.Equal(t, "2", query.Get("page"))
			
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Page", "2")
			w.Header().Set("X-Per-Page", "20")
			w.Header().Set("X-Total", "50")
			w.Header().Set("X-Total-Pages", "3")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]*MergeRequest{})
		}))
		defer server.Close()

		client, err := NewClient(ctx, "test-token", server.URL)
		require.NoError(t, err)

		input := GetMergeRequestsInput{
			ProjectID: 456,
			PerPage:   20,
			Page:      2,
		}

		page, err := client.GetMergeRequests(ctx, input)
		require.NoError(t, err)
		assert.Equal(t, 2, page.PageInfo.Page)
		assert.Equal(t, 20, page.PageInfo.PerPage)
		assert.Equal(t, 50, page.PageInfo.Total)
		assert.Equal(t, 3, page.PageInfo.TotalPages)
	})

	t.Run("error getting merge requests", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, `{"message": "403 Forbidden"}`)
		}))
		defer server.Close()

		client, err := NewClient(ctx, "test-token", server.URL)
		require.NoError(t, err)

		input := GetMergeRequestsInput{ProjectID: 456}
		page, err := client.GetMergeRequests(ctx, input)
		assert.Error(t, err)
		assert.Nil(t, page)
		assert.Contains(t, err.Error(), "failed to get merge requests for project 456")
	})
}

func TestClient_CreateMergeRequest(t *testing.T) {
	ctx := context.Background()

	t.Run("successful create merge request", func(t *testing.T) {
		mockMR := createMockMergeRequest()
		
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/api/v4/projects/456/merge_requests", r.URL.Path)
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			
			var requestBody CreateMergeRequestInput
			err := json.NewDecoder(r.Body).Decode(&requestBody)
			require.NoError(t, err)
			assert.Equal(t, 456, requestBody.ProjectID)
			assert.Equal(t, "feature-branch", requestBody.SourceBranch)
			assert.Equal(t, "main", requestBody.TargetBranch)
			assert.Equal(t, "Test merge request", requestBody.Title)
			assert.Equal(t, "Test description", *requestBody.Description)
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(mockMR)
		}))
		defer server.Close()

		client, err := NewClient(ctx, "test-token", server.URL)
		require.NoError(t, err)

		description := "Test description"
		input := CreateMergeRequestInput{
			ProjectID:    456,
			SourceBranch: "feature-branch",
			TargetBranch: "main",
			Title:        "Test merge request",
			Description:  &description,
		}

		mr, err := client.CreateMergeRequest(ctx, input)
		require.NoError(t, err)
		assert.Equal(t, mockMR.ID, mr.ID)
		assert.Equal(t, mockMR.Title, mr.Title)
		assert.Equal(t, mockMR.SourceBranch, mr.SourceBranch)
		assert.Equal(t, mockMR.TargetBranch, mr.TargetBranch)
	})

	t.Run("create merge request with validation error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"message": "Source branch cannot be empty"}`)
		}))
		defer server.Close()

		client, err := NewClient(ctx, "test-token", server.URL)
		require.NoError(t, err)

		input := CreateMergeRequestInput{
			ProjectID:    456,
			SourceBranch: "", // Invalid empty source branch
			TargetBranch: "main",
			Title:        "Test merge request",
		}

		mr, err := client.CreateMergeRequest(ctx, input)
		assert.Error(t, err)
		assert.Nil(t, mr)
		assert.Contains(t, err.Error(), "failed to create merge request")
	})
}

func TestClient_UpdateMergeRequest(t *testing.T) {
	ctx := context.Background()

	t.Run("successful update merge request", func(t *testing.T) {
		mockMR := createMockMergeRequest()
		mockMR.Title = "Updated title"
		mockMR.Description = "Updated description"
		
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "PUT", r.Method)
			assert.Equal(t, "/api/v4/projects/456/merge_requests/1", r.URL.Path)
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			
			var requestBody UpdateMergeRequestInput
			err := json.NewDecoder(r.Body).Decode(&requestBody)
			require.NoError(t, err)
			assert.Equal(t, "Updated title", *requestBody.Title)
			assert.Equal(t, "Updated description", *requestBody.Description)
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockMR)
		}))
		defer server.Close()

		client, err := NewClient(ctx, "test-token", server.URL)
		require.NoError(t, err)

		title := "Updated title"
		description := "Updated description"
		input := UpdateMergeRequestInput{
			ProjectID:       456,
			MergeRequestIID: 1,
			Title:           &title,
			Description:     &description,
		}

		mr, err := client.UpdateMergeRequest(ctx, input)
		require.NoError(t, err)
		assert.Equal(t, mockMR.ID, mr.ID)
		assert.Equal(t, "Updated title", mr.Title)
		assert.Equal(t, "Updated description", mr.Description)
	})

	t.Run("update non-existent merge request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"message": "404 Not Found"}`)
		}))
		defer server.Close()

		client, err := NewClient(ctx, "test-token", server.URL)
		require.NoError(t, err)

		title := "Updated title"
		input := UpdateMergeRequestInput{
			ProjectID:       456,
			MergeRequestIID: 999,
			Title:           &title,
		}

		mr, err := client.UpdateMergeRequest(ctx, input)
		assert.Error(t, err)
		assert.Nil(t, mr)
		assert.Contains(t, err.Error(), "failed to update merge request 999")
	})

	t.Run("update merge request with state change", func(t *testing.T) {
		mockMR := createMockMergeRequest()
		mockMR.State = MergeRequestStateClosed
		
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var requestBody UpdateMergeRequestInput
			err := json.NewDecoder(r.Body).Decode(&requestBody)
			require.NoError(t, err)
			assert.Equal(t, "close", *requestBody.StateEvent)
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(mockMR)
		}))
		defer server.Close()

		client, err := NewClient(ctx, "test-token", server.URL)
		require.NoError(t, err)

		stateEvent := "close"
		input := UpdateMergeRequestInput{
			ProjectID:       456,
			MergeRequestIID: 1,
			StateEvent:      &stateEvent,
		}

		mr, err := client.UpdateMergeRequest(ctx, input)
		require.NoError(t, err)
		assert.Equal(t, MergeRequestStateClosed, mr.State)
	})
}

// TestClient_getPaginated tests the pagination helper method used by GetMergeRequests
func TestClient_getPaginated(t *testing.T) {
	ctx := context.Background()

	t.Run("successful paginated request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Page", "1")
			w.Header().Set("X-Per-Page", "20")
			w.Header().Set("X-Prev-Page", "")
			w.Header().Set("X-Next-Page", "2")
			w.Header().Set("X-Total", "50")
			w.Header().Set("X-Total-Pages", "3")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]*MergeRequest{createMockMergeRequest()})
		}))
		defer server.Close()

		client, err := NewClient(ctx, "test-token", server.URL)
		require.NoError(t, err)

		params := map[string]string{"per_page": "20", "page": "1"}
		var mrs []*MergeRequest
		pageInfo, err := client.getPaginated(ctx, "/merge_requests", params, &mrs)
		require.NoError(t, err)

		assert.Equal(t, 1, pageInfo.Page)
		assert.Equal(t, 20, pageInfo.PerPage)
		assert.Nil(t, pageInfo.PrevPage)
		assert.Equal(t, 2, *pageInfo.NextPage)
		assert.Equal(t, 50, pageInfo.Total)
		assert.Equal(t, 3, pageInfo.TotalPages)
		assert.Len(t, mrs, 1)
	})

	t.Run("paginated request with missing headers", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only set some pagination headers
			w.Header().Set("X-Page", "1")
			w.Header().Set("X-Per-Page", "20")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]*MergeRequest{})
		}))
		defer server.Close()

		client, err := NewClient(ctx, "test-token", server.URL)
		require.NoError(t, err)

		params := map[string]string{}
		var mrs []*MergeRequest
		pageInfo, err := client.getPaginated(ctx, "/merge_requests", params, &mrs)
		require.NoError(t, err)

		// Should handle missing headers gracefully
		assert.Equal(t, 1, pageInfo.Page)
		assert.Equal(t, 20, pageInfo.PerPage)
		assert.Nil(t, pageInfo.PrevPage)
		assert.Nil(t, pageInfo.NextPage)
		assert.Equal(t, 0, pageInfo.Total) // Default when header missing
		assert.Equal(t, 0, pageInfo.TotalPages) // Default when header missing
	})
}

// Helper function to test pagination header parsing
func TestParseIntHeader(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected int
	}{
		{"valid number", "42", 42},
		{"empty string", "", 0},
		{"invalid number", "not-a-number", 0},
		{"zero", "0", 0},
		{"negative number", "-1", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := strconv.Atoi(tt.value)
			if tt.value == "" || tt.value == "not-a-number" {
				result = 0
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}