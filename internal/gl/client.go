package gl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
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

// NewClient creates a new GitLab client using personal access token authentication.
// baseURL should be the GitLab instance base URL (e.g., "https://gitlab.com" or "https://gitlab.mycompany.com")
func NewClient(ctx context.Context, token string, baseURL string) (*Client, error) {
	if token == "" {
		return nil, errors.Errorf("no GitLab token provided (do you need to configure one?)")
	}
	
	if baseURL == "" {
		baseURL = "https://gitlab.com"
	}
	
	// Ensure baseURL doesn't have trailing slash and has proper format
	baseURL = strings.TrimRight(baseURL, "/")
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}
	
	// Validate the base URL
	if _, err := url.Parse(baseURL); err != nil {
		return nil, errors.Wrapf(err, "invalid GitLab base URL: %s", baseURL)
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

// request executes an HTTP request to the GitLab API with proper authentication and error handling
func (c *Client) request(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal request body")
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}
	
	// Construct the full URL for the API endpoint
	apiURL := c.baseURL + "/api/v4" + path
	
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
	
	// Set user agent
	req.Header.Set("User-Agent", "av-cli")
	
	log := logrus.WithFields(logrus.Fields{
		"method": method,
		"url":    apiURL,
	})
	log.Debug("executing GitLab API request...")
	startTime := time.Now()
	
	resp, err := c.httpClient.Do(req)
	
	elapsed := time.Since(startTime)
	log = log.WithField("elapsed", elapsed)
	
	if err != nil {
		log.WithError(err).Debug("GitLab API request failed")
		return nil, errors.Wrap(err, "HTTP request failed")
	}
	
	log.WithField("status_code", resp.StatusCode).Debug("GitLab API request completed")
	
	return resp, nil
}

// get executes a GET request and unmarshals the response into the provided result struct
func (c *Client) get(ctx context.Context, path string, params map[string]string, result interface{}) (reterr error) {
	log := logrus.WithField("path", path)
	log.Debug("executing GitLab API GET request...")
	startTime := time.Now()
	
	defer func() {
		log := log.WithFields(logrus.Fields{
			"elapsed": time.Since(startTime),
			"result":  logutils.Format("%#+v", result),
		})
		if reterr != nil {
			log.WithError(reterr).Debug("GitLab API GET request failed")
		} else {
			log.Debug("GitLab API GET request succeeded")
		}
	}()
	
	// Add query parameters if provided
	if len(params) > 0 {
		u, err := url.Parse(c.baseURL + "/api/v4" + path)
		if err != nil {
			return errors.Wrap(err, "failed to parse URL")
		}
		q := u.Query()
		for key, value := range params {
			q.Set(key, value)
		}
		u.RawQuery = q.Encode()
		path = strings.TrimPrefix(u.RequestURI(), "/api/v4")
	}
	
	resp, err := c.request(ctx, "GET", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if err := c.checkResponse(resp); err != nil {
		return err
	}
	
	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return errors.Wrap(err, "failed to decode response")
		}
	}
	
	return nil
}

// post executes a POST request and unmarshals the response into the provided result struct
func (c *Client) post(ctx context.Context, path string, body interface{}, result interface{}) (reterr error) {
	log := logrus.WithFields(logrus.Fields{
		"path": path,
		"body": logutils.Format("%#+v", body),
	})
	log.Debug("executing GitLab API POST request...")
	startTime := time.Now()
	
	defer func() {
		log := log.WithFields(logrus.Fields{
			"elapsed": time.Since(startTime),
			"result":  logutils.Format("%#+v", result),
		})
		if reterr != nil {
			log.WithError(reterr).Debug("GitLab API POST request failed")
		} else {
			log.Debug("GitLab API POST request succeeded")
		}
	}()
	
	resp, err := c.request(ctx, "POST", path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if err := c.checkResponse(resp); err != nil {
		return err
	}
	
	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return errors.Wrap(err, "failed to decode response")
		}
	}
	
	return nil
}

// put executes a PUT request and unmarshals the response into the provided result struct
func (c *Client) put(ctx context.Context, path string, body interface{}, result interface{}) (reterr error) {
	log := logrus.WithFields(logrus.Fields{
		"path": path,
		"body": logutils.Format("%#+v", body),
	})
	log.Debug("executing GitLab API PUT request...")
	startTime := time.Now()
	
	defer func() {
		log := log.WithFields(logrus.Fields{
			"elapsed": time.Since(startTime),
			"result":  logutils.Format("%#+v", result),
		})
		if reterr != nil {
			log.WithError(reterr).Debug("GitLab API PUT request failed")
		} else {
			log.Debug("GitLab API PUT request succeeded")
		}
	}()
	
	resp, err := c.request(ctx, "PUT", path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if err := c.checkResponse(resp); err != nil {
		return err
	}
	
	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return errors.Wrap(err, "failed to decode response")
		}
	}
	
	return nil
}

// checkResponse checks the HTTP response for errors and returns appropriate error messages
func (c *Client) checkResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	
	// Try to read the error response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Errorf("GitLab API request failed with status %d", resp.StatusCode)
	}
	
	// Try to parse GitLab error response
	var errorResp struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}
	
	if err := json.Unmarshal(bodyBytes, &errorResp); err == nil && errorResp.Message != "" {
		return errors.Errorf("GitLab API error (status %d): %s", resp.StatusCode, errorResp.Message)
	}
	
	if err := json.Unmarshal(bodyBytes, &errorResp); err == nil && errorResp.Error != "" {
		return errors.Errorf("GitLab API error (status %d): %s", resp.StatusCode, errorResp.Error)
	}
	
	// Fall back to generic error with response body
	return errors.Errorf("GitLab API request failed (status %d): %s", resp.StatusCode, string(bodyBytes))
}

// getPaginated executes a GET request and returns pagination info along with the response
func (c *Client) getPaginated(ctx context.Context, path string, params map[string]string, result interface{}) (*PageInfo, error) {
	// Add query parameters if provided
	if len(params) > 0 {
		u, err := url.Parse(c.baseURL + "/api/v4" + path)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse URL")
		}
		q := u.Query()
		for key, value := range params {
			q.Set(key, value)
		}
		u.RawQuery = q.Encode()
		path = strings.TrimPrefix(u.RequestURI(), "/api/v4")
	}
	
	resp, err := c.request(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}
	
	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return nil, errors.Wrap(err, "failed to decode response")
		}
	}
	
	// Extract pagination info from response headers
	pageInfo := &PageInfo{}
	
	if page := resp.Header.Get("X-Page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			pageInfo.Page = p
		}
	}
	
	if perPage := resp.Header.Get("X-Per-Page"); perPage != "" {
		if pp, err := strconv.Atoi(perPage); err == nil {
			pageInfo.PerPage = pp
		}
	}
	
	if prevPage := resp.Header.Get("X-Prev-Page"); prevPage != "" && prevPage != "" {
		if pp, err := strconv.Atoi(prevPage); err == nil {
			pageInfo.PrevPage = &pp
		}
	}
	
	if nextPage := resp.Header.Get("X-Next-Page"); nextPage != "" && nextPage != "" {
		if np, err := strconv.Atoi(nextPage); err == nil {
			pageInfo.NextPage = &np
		}
	}
	
	if totalPages := resp.Header.Get("X-Total-Pages"); totalPages != "" {
		if tp, err := strconv.Atoi(totalPages); err == nil {
			pageInfo.TotalPages = tp
		}
	}
	
	if total := resp.Header.Get("X-Total"); total != "" {
		if t, err := strconv.Atoi(total); err == nil {
			pageInfo.Total = t
		}
	}
	
	return pageInfo, nil
}

// BaseURL returns the configured base URL for this GitLab client
func (c *Client) BaseURL() string {
	return c.baseURL
}