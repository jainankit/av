package gl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	ctx := context.Background()

	t.Run("success with default base URL", func(t *testing.T) {
		client, err := NewClient(ctx, "test-token", "")
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "https://gitlab.com", client.BaseURL())
		assert.Equal(t, "test-token", client.token)
		assert.NotNil(t, client.httpClient)
		assert.Equal(t, 30*time.Second, client.httpClient.Timeout)
	})

	t.Run("success with custom base URL", func(t *testing.T) {
		client, err := NewClient(ctx, "test-token", "https://gitlab.mycompany.com")
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "https://gitlab.mycompany.com", client.BaseURL())
	})

	t.Run("success with base URL without protocol", func(t *testing.T) {
		client, err := NewClient(ctx, "test-token", "gitlab.mycompany.com")
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "https://gitlab.mycompany.com", client.BaseURL())
	})

	t.Run("success with trailing slash removed", func(t *testing.T) {
		client, err := NewClient(ctx, "test-token", "https://gitlab.mycompany.com/")
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "https://gitlab.mycompany.com", client.BaseURL())
	})

	t.Run("error with empty token", func(t *testing.T) {
		client, err := NewClient(ctx, "", "https://gitlab.com")
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "no GitLab token provided")
	})

	t.Run("error with invalid base URL", func(t *testing.T) {
		client, err := NewClient(ctx, "test-token", "://invalid-url")
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "invalid GitLab base URL")
	})
}

func TestClient_request(t *testing.T) {
	ctx := context.Background()

	t.Run("successful GET request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/api/v4/test", r.URL.Path)
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			assert.Equal(t, "av-cli", r.Header.Get("User-Agent"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"success": true}`)
		}))
		defer server.Close()

		client, err := NewClient(ctx, "test-token", server.URL)
		require.NoError(t, err)

		resp, err := client.request(ctx, "GET", "/test", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("successful POST request with body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/api/v4/test", r.URL.Path)
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			var body map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&body)
			require.NoError(t, err)
			assert.Equal(t, "test-value", body["test_field"])

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"created": true}`)
		}))
		defer server.Close()

		client, err := NewClient(ctx, "test-token", server.URL)
		require.NoError(t, err)

		requestBody := map[string]interface{}{
			"test_field": "test-value",
		}

		resp, err := client.request(ctx, "POST", "/test", requestBody)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("request with invalid body", func(t *testing.T) {
		client, err := NewClient(ctx, "test-token", "https://gitlab.com")
		require.NoError(t, err)

		// Channels cannot be marshaled to JSON
		invalidBody := make(chan int)

		_, err = client.request(ctx, "POST", "/test", invalidBody)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal request body")
	})
}

func TestClient_get(t *testing.T) {
	ctx := context.Background()

	t.Run("successful GET with response unmarshaling", func(t *testing.T) {
		expectedResponse := map[string]interface{}{
			"id":   123,
			"name": "test-project",
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/api/v4/projects/123", r.URL.Path)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedResponse)
		}))
		defer server.Close()

		client, err := NewClient(ctx, "test-token", server.URL)
		require.NoError(t, err)

		var result map[string]interface{}
		err = client.get(ctx, "/projects/123", &result)
		require.NoError(t, err)

		assert.Equal(t, float64(123), result["id"]) // JSON numbers are float64
		assert.Equal(t, "test-project", result["name"])
	})

	t.Run("GET with nil result", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client, err := NewClient(ctx, "test-token", server.URL)
		require.NoError(t, err)

		err = client.get(ctx, "/projects/123", nil)
		require.NoError(t, err)
	})

	t.Run("GET with invalid JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `invalid json`)
		}))
		defer server.Close()

		client, err := NewClient(ctx, "test-token", server.URL)
		require.NoError(t, err)

		var result map[string]interface{}
		err = client.get(ctx, "/projects/123", &result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode response")
	})
}

func TestClient_post(t *testing.T) {
	ctx := context.Background()

	t.Run("successful POST with request and response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/api/v4/projects", r.URL.Path)

			var requestBody map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&requestBody)
			require.NoError(t, err)
			assert.Equal(t, "new-project", requestBody["name"])

			response := map[string]interface{}{
				"id":   456,
				"name": "new-project",
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client, err := NewClient(ctx, "test-token", server.URL)
		require.NoError(t, err)

		requestBody := map[string]interface{}{
			"name": "new-project",
		}

		var result map[string]interface{}
		err = client.post(ctx, "/projects", requestBody, &result)
		require.NoError(t, err)

		assert.Equal(t, float64(456), result["id"])
		assert.Equal(t, "new-project", result["name"])
	})
}

func TestClient_put(t *testing.T) {
	ctx := context.Background()

	t.Run("successful PUT with request and response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "PUT", r.Method)
			assert.Equal(t, "/api/v4/projects/123", r.URL.Path)

			var requestBody map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&requestBody)
			require.NoError(t, err)
			assert.Equal(t, "updated-project", requestBody["name"])

			response := map[string]interface{}{
				"id":   123,
				"name": "updated-project",
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client, err := NewClient(ctx, "test-token", server.URL)
		require.NoError(t, err)

		requestBody := map[string]interface{}{
			"name": "updated-project",
		}

		var result map[string]interface{}
		err = client.put(ctx, "/projects/123", requestBody, &result)
		require.NoError(t, err)

		assert.Equal(t, float64(123), result["id"])
		assert.Equal(t, "updated-project", result["name"])
	})
}

func TestClient_checkResponse(t *testing.T) {
	client := &Client{}

	t.Run("successful response codes", func(t *testing.T) {
		successCodes := []int{200, 201, 202, 204, 299}

		for _, code := range successCodes {
			resp := &http.Response{
				StatusCode: code,
				Body:       http.NoBody,
			}
			err := client.checkResponse(resp)
			assert.NoError(t, err, "Status code %d should not return error", code)
		}
	})

	t.Run("error response with GitLab message", func(t *testing.T) {
		errorBody := `{"message": "Project not found"}`
		resp := &http.Response{
			StatusCode: 404,
			Body:       http.NoBody,
		}
		resp.Body = stringReadCloser(errorBody)

		err := client.checkResponse(resp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "GitLab API error (status 404): Project not found")
	})

	t.Run("error response with GitLab error field", func(t *testing.T) {
		errorBody := `{"error": "Unauthorized"}`
		resp := &http.Response{
			StatusCode: 401,
			Body:       http.NoBody,
		}
		resp.Body = stringReadCloser(errorBody)

		err := client.checkResponse(resp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "GitLab API error (status 401): Unauthorized")
	})

	t.Run("error response with generic message", func(t *testing.T) {
		errorBody := `Server Error`
		resp := &http.Response{
			StatusCode: 500,
			Body:       http.NoBody,
		}
		resp.Body = stringReadCloser(errorBody)

		err := client.checkResponse(resp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "GitLab API request failed (status 500): Server Error")
	})

	t.Run("error response with unreadable body", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: 500,
			Body:       errorReadCloser{},
		}

		err := client.checkResponse(resp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "GitLab API request failed with status 500")
	})
}

// Helper functions for testing

func stringReadCloser(s string) *strings.Reader {
	return strings.NewReader(s)
}

type errorReadCloser struct{}

func (e errorReadCloser) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("read error")
}

func (e errorReadCloser) Close() error {
	return nil
}