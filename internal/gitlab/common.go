package gitlab

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// PageInfo contains information about the current/previous/next page of
// results when using GitLab's GraphQL paginated APIs.
type PageInfo struct {
	EndCursor       string
	HasNextPage     bool
	HasPreviousPage bool
	StartCursor     string
}

// Ptr returns a pointer to the argument.
//
// It's a convenience function to make working with the API easier: since Go
// disallows pointers-to-literals, and optional input fields are expressed as
// pointers, this function can be used to easily set optional fields to non-nil
// primitives.
//
// For example, `CreateMergeRequestInput{Draft: Ptr(true)}`.
func Ptr[T any](v T) *T {
	return &v
}

// nullable returns a pointer to the argument if it's not the zero value,
// otherwise it returns nil.
// This is useful to translate between Golang-style "unset is zero" and GraphQL
// which distinguishes between unset (null) and zero values.
func nullable[T comparable](v T) *T {
	var zero T
	if v == zero {
		return nil
	}
	return &v
}

// ProjectSlug represents a GitLab project in namespace/project format
type ProjectSlug struct {
	Namespace string
	Project   string
}

// String returns the full project slug in namespace/project format
func (p ProjectSlug) String() string {
	return fmt.Sprintf("%s/%s", p.Namespace, p.Project)
}

// ParseProjectSlug parses a project slug string in the format
// "namespace/project" and returns a ProjectSlug struct. Returns an error if
// the format is invalid.
func ParseProjectSlug(slug string) (ProjectSlug, error) {
	parts := strings.Split(slug, "/")
	if len(parts) != 2 {
		return ProjectSlug{}, fmt.Errorf(
			"invalid project slug format: expected 'namespace/project', got '%s'",
			slug,
		)
	}
	
	if parts[0] == "" || parts[1] == "" {
		return ProjectSlug{}, fmt.Errorf(
			"invalid project slug format: namespace and project cannot be empty",
		)
	}
	
	return ProjectSlug{
		Namespace: parts[0],
		Project:   parts[1],
	}, nil
}

// ParseGitLabURL extracts the project slug from a GitLab URL.
// Supports various GitLab URL formats including SSH and HTTPS.
func ParseGitLabURL(gitURL string) (ProjectSlug, error) {
	// Handle SSH URLs like git@gitlab.com:namespace/project.git
	sshRegex := regexp.MustCompile(`^git@([^:]+):(.+)\.git$`)
	if matches := sshRegex.FindStringSubmatch(gitURL); matches != nil {
		return ParseProjectSlug(matches[2])
	}
	
	// Handle HTTPS URLs
	parsedURL, err := url.Parse(gitURL)
	if err != nil {
		return ProjectSlug{}, fmt.Errorf("invalid URL format: %w", err)
	}
	
	// Remove leading slash and .git suffix
	path := strings.TrimPrefix(parsedURL.Path, "/")
	path = strings.TrimSuffix(path, ".git")
	
	if path == "" {
		return ProjectSlug{}, fmt.Errorf("no project path found in URL")
	}
	
	return ParseProjectSlug(path)
}

// IsGitLabURL checks if a given URL is a GitLab URL.
// It checks for common GitLab patterns including gitlab.com and self-hosted
// instances.
func IsGitLabURL(gitURL string) bool {
	// Handle SSH URLs
	sshRegex := regexp.MustCompile(`^git@([^:]+):(.+)\.git$`)
	if matches := sshRegex.FindStringSubmatch(gitURL); matches != nil {
		hostname := matches[1]
		// Check for gitlab.com or common GitLab patterns
		return strings.Contains(hostname, "gitlab") || hostname == "gitlab.com"
	}
	
	// Handle HTTPS URLs
	parsedURL, err := url.Parse(gitURL)
	if err != nil {
		return false
	}
	
	hostname := parsedURL.Hostname()
	// Check for gitlab.com or common GitLab patterns
	return strings.Contains(hostname, "gitlab") || hostname == "gitlab.com"
}

// NormalizeBaseURL ensures the base URL is properly formatted for GitLab API
// usage. It adds the API path if not present and ensures proper trailing
// slash handling.
func NormalizeBaseURL(baseURL string) string {
	if baseURL == "" {
		return "https://gitlab.com"
	}
	
	// Remove trailing slash
	baseURL = strings.TrimSuffix(baseURL, "/")
	
	// Add API path if not present
	if !strings.HasSuffix(baseURL, "/api/graphql") &&
		!strings.Contains(baseURL, "/api/") {
		baseURL = baseURL + "/api/graphql"
	}
	
	return baseURL
}

// RetryConfig configures retry behavior for GitLab API requests
type RetryConfig struct {
	// Maximum number of retry attempts (including initial attempt)
	MaxAttempts int
	InitialDelay    time.Duration // Initial delay before first retry
	MaxDelay        time.Duration // Maximum delay between retries
	BackoffFactor   float64       // Exponential backoff multiplier
	JitterEnabled   bool          // Whether to add random jitter to delays
	RateLimitDelay  time.Duration // Additional delay for rate limit errors
}

// DefaultRetryConfig returns the default retry configuration for GitLab API
// requests
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:     3,
		InitialDelay:    1 * time.Second,
		MaxDelay:        30 * time.Second,
		BackoffFactor:   2.0,
		JitterEnabled:   true,
		RateLimitDelay:  60 * time.Second, // GitLab rate limits reset every minute
	}
}

// RetryableFunc is a function that can be retried with exponential backoff
type RetryableFunc func(ctx context.Context) error

// WithRetry executes a function with exponential backoff retry logic.
// It automatically handles GitLab rate limiting and temporary errors.
func WithRetry(
	ctx context.Context,
	config RetryConfig,
	fn RetryableFunc,
) error {
	var lastErr error
	
	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// Execute the function
		err := fn(ctx)
		if err == nil {
			return nil // Success
		}
		
		lastErr = err
		
		// Don't retry on the last attempt
		if attempt == config.MaxAttempts-1 {
			break
		}
		
		// Check if this error is retryable
		if !IsRetryableError(err) {
			return err // Non-retryable error, fail immediately
		}
		
		// Calculate delay
		delay := calculateDelay(attempt, config, IsHTTPRateLimited(err))
		
		// Check if context is cancelled before sleeping
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}
	
	return lastErr
}

// calculateDelay computes the delay for a retry attempt using exponential
// backoff
func calculateDelay(
	attempt int,
	config RetryConfig,
	isRateLimit bool,
) time.Duration {
	// Base delay calculation with exponential backoff
	baseDelay := float64(config.InitialDelay)
	backoffMultiplier := math.Pow(config.BackoffFactor, float64(attempt))
	delay := time.Duration(baseDelay * backoffMultiplier)
	
	// Apply maximum delay cap
	if delay > config.MaxDelay {
		delay = config.MaxDelay
	}
	
	// Add additional delay for rate limit errors
	if isRateLimit {
		delay += config.RateLimitDelay
	}
	
	// Add jitter to prevent thundering herd
	if config.JitterEnabled {
		jitter := time.Duration(rand.Float64() * float64(delay) * 0.1) // 10% jitter
		delay += jitter
	}
	
	return delay
}

// RateLimiter implements a simple rate limiter for GitLab API requests
type RateLimiter struct {
	requests    chan struct{}
	ticker      *time.Ticker
	maxRequests int
	interval    time.Duration
}

// NewRateLimiter creates a new rate limiter with the specified rate.
// GitLab.com has a rate limit of 2000 requests per minute for authenticated
// users.
func NewRateLimiter(maxRequests int, interval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests:    make(chan struct{}, maxRequests),
		maxRequests: maxRequests,
		interval:    interval,
	}
	
	// Fill the initial bucket
	for i := 0; i < maxRequests; i++ {
		rl.requests <- struct{}{}
	}
	
	// Start the refill ticker
	rl.ticker = time.NewTicker(interval / time.Duration(maxRequests))
	go rl.refill()
	
	return rl
}

// Wait blocks until a request slot is available
func (rl *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-rl.requests:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// refill replenishes the rate limiter tokens
func (rl *RateLimiter) refill() {
	for range rl.ticker.C {
		select {
		case rl.requests <- struct{}{}:
			// Token added
		default:
			// Bucket is full, skip
		}
	}
}

// Close stops the rate limiter and cleans up resources
func (rl *RateLimiter) Close() {
	if rl.ticker != nil {
		rl.ticker.Stop()
	}
	close(rl.requests)
}
