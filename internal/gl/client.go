package gl

import (
	"context"
	"net/http"
	"time"

	"emperror.dev/errors"
	"github.com/aviator-co/av/internal/config"
	"github.com/aviator-co/av/internal/utils/logutils"
	"github.com/sirupsen/logrus"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

// NewClient creates a new GitLab client.
// It takes configuration from the global config.Av.GitLab variable.
func NewClient(ctx context.Context, token string) (*Client, error) {
	if token == "" {
		return nil, errors.Errorf("no GitLab token provided (do you need to configure one?)")
	}
	
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	// Set base URL, defaulting to gitlab.com if not specified
	// TODO: Update this when GitLab configuration is added to config.Av in step 1.2
	baseURL := "https://gitlab.com"
	// if config.Av.GitLab != nil && config.Av.GitLab.BaseURL != "" {
	//     baseURL = config.Av.GitLab.BaseURL
	// }
	
	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
		token:      token,
	}, nil
}

func (c *Client) request(ctx context.Context, method, endpoint string, body interface{}) (reterr error) {
	log := logrus.WithFields(logrus.Fields{
		"method":   method,
		"endpoint": endpoint,
		"body":     logutils.Format("%#+v", body),
	})
	log.Debug("executing GitLab API request...")
	startTime := time.Now()
	defer func() {
		log := log.WithFields(logrus.Fields{
			"elapsed": time.Since(startTime),
		})
		if reterr != nil {
			log.WithError(reterr).Debug("GitLab API request failed")
		} else {
			log.Debug("GitLab API request succeeded")
		}
	}()
	
	// TODO: Implement actual HTTP request logic with proper authentication
	// This is a placeholder implementation that will be expanded in later steps
	return nil
}