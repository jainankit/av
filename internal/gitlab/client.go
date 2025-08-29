package gitlab

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/aviator-co/av/internal/utils/logutils"
	"github.com/sirupsen/logrus"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

// NewClient creates a new GitLab client.
// baseURL should be the GitLab instance base URL (e.g., "https://gitlab.com" or "https://gitlab.example.com")
func NewClient(ctx context.Context, token, baseURL string) (*Client, error) {
	if token == "" {
		return nil, errors.Errorf("no GitLab token provided (do you need to configure one?)")
	}
	if baseURL == "" {
		baseURL = "https://gitlab.com"
	}

	// Ensure baseURL doesn't have trailing slash and has proper format
	baseURL = strings.TrimSuffix(baseURL, "/")
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
		token:      token,
	}, nil
}

// makeRequest performs an HTTP request to the GitLab API with proper authentication and logging
func (c *Client) makeRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	var bodyBytes []byte

	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal request body")
		}
		reqBody = bytes.NewReader(bodyBytes)
	}

	// Construct full URL
	apiURL := c.baseURL + "/api/v4" + endpoint
	
	log := logrus.WithFields(logrus.Fields{
		"method":   method,
		"url":      apiURL,
		"body":     logutils.Format("%s", string(bodyBytes)),
	})
	log.Debug("executing GitLab API request...")
	startTime := time.Now()

	req, err := http.NewRequestWithContext(ctx, method, apiURL, reqBody)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create HTTP request")
	}

	// Set authentication header
	req.Header.Set("Authorization", "Bearer "+c.token)
	
	// Set content type for requests with body
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	elapsed := time.Since(startTime)

	log = log.WithFields(logrus.Fields{
		"elapsed": elapsed,
	})

	if err != nil {
		log.WithError(err).Debug("GitLab API request failed")
		return nil, errors.Wrap(err, "HTTP request failed")
	}

	log = log.WithField("status_code", resp.StatusCode)
	
	if resp.StatusCode >= 400 {
		// Read error response body for better error messages
		errorBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		
		log.WithField("error_body", string(errorBody)).Debug("GitLab API request failed with error status")
		return nil, NewHTTPError(resp.StatusCode, string(errorBody))
	}

	log.Debug("GitLab API request succeeded")
	return resp, nil
}

// get performs a GET request to the GitLab API
func (c *Client) get(ctx context.Context, endpoint string, result interface{}) error {
	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if result != nil {
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(result); err != nil {
			return errors.Wrap(err, "failed to decode response body")
		}
	}

	return nil
}

// post performs a POST request to the GitLab API
func (c *Client) post(ctx context.Context, endpoint string, body interface{}, result interface{}) error {
	resp, err := c.makeRequest(ctx, "POST", endpoint, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if result != nil {
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(result); err != nil {
			return errors.Wrap(err, "failed to decode response body")
		}
	}

	return nil
}

// put performs a PUT request to the GitLab API
func (c *Client) put(ctx context.Context, endpoint string, body interface{}, result interface{}) error {
	resp, err := c.makeRequest(ctx, "PUT", endpoint, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if result != nil {
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(result); err != nil {
			return errors.Wrap(err, "failed to decode response body")
		}
	}

	return nil
}

// HTTPError represents an HTTP error response from the GitLab API
type HTTPError struct {
	StatusCode int
	Body       string
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("GitLab API error (status %d): %s", e.StatusCode, e.Body)
}

// NewHTTPError creates a new HTTPError
func NewHTTPError(statusCode int, body string) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Body:       body,
	}
}

// IsHTTPUnauthorized returns true if the given error is an HTTP 401 Unauthorized error
func IsHTTPUnauthorized(err error) bool {
	if httpErr, ok := err.(*HTTPError); ok {
		return httpErr.StatusCode == 401
	}
	return false
}

// IsHTTPNotFound returns true if the given error is an HTTP 404 Not Found error
func IsHTTPNotFound(err error) bool {
	if httpErr, ok := err.(*HTTPError); ok {
		return httpErr.StatusCode == 404
	}
	return false
}

// IsHTTPForbidden returns true if the given error is an HTTP 403 Forbidden error
func IsHTTPForbidden(err error) bool {
	if httpErr, ok := err.(*HTTPError); ok {
		return httpErr.StatusCode == 403
	}
	return false
}

// IsHTTPConflict returns true if the given error is an HTTP 409 Conflict error
func IsHTTPConflict(err error) bool {
	if httpErr, ok := err.(*HTTPError); ok {
		return httpErr.StatusCode == 409
	}
	return false
}

// buildURL constructs a URL with query parameters
func (c *Client) buildURL(endpoint string, params map[string]string) string {
	if len(params) == 0 {
		return endpoint
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		// If we can't parse the URL, just return it as-is
		return endpoint
	}

	q := u.Query()
	for key, value := range params {
		q.Set(key, value)
	}
	u.RawQuery = q.Encode()

	return u.String()
}