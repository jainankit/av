package gh

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// pullrequest_bench_test.go provides performance benchmarking tests for GitHub API operations.
// This file focuses specifically on measuring throughput, concurrent request handling,
// pagination performance, and rate limiting behavior for the RepoPullRequests API call.
//
// The benchmarks are designed to:
// - Measure requests per second under various load conditions
// - Test pagination efficiency with large datasets
// - Evaluate performance under rate limiting scenarios
// - Provide baseline performance metrics for regression detection
//
// Run benchmarks with: go test -bench=BenchmarkRepoPullRequests -benchmem ./internal/gh/

const (
	// GraphQL query for pull requests used in benchmarking
	benchmarkPRQuery = "query($after:String$first:Int!$owner:String!$repo:String!$states:[PullRequestState!]){repository(owner: $owner, name: $repo){pullRequests(states: $states, first: $first, after: $after){totalCount,nodes{id,number,headRefName,baseRefName,isDraft,permalink,state,title,body,mergeCommit{oid},timelineItems(last: 10, itemTypes: [CLOSED_EVENT, MERGED_EVENT]){nodes{... on ClosedEvent{closer{... on Commit{oid}}},... on MergedEvent{commit{oid}}}}},pageInfo{endCursor,hasNextPage,hasPreviousPage,startCursor}}}}"
)

// MockGitHubServerForBench provides a high-performance mock GitHub GraphQL server
// optimized for benchmarking scenarios. It supports configurable response generation,
// pagination simulation, response delays, and concurrent request tracking.
type MockGitHubServerForBench struct {
	server *httptest.Server
	
	// Configuration
	totalPRs        int
	pageSize        int
	responseDelayMs int64
	
	// Request tracking
	requestCount    int64
	concurrentReqs  int64
	maxConcurrent   int64
	
	// Data generation
	prDatasets map[string][]benchmarkPR
	
	// State
	mutex sync.RWMutex
}

// benchmarkPR represents a pull request optimized for benchmark testing
type benchmarkPR struct {
	ID             string
	Number         int
	HeadRefName    string
	BaseRefName    string
	IsDraft        bool
	State          string
	Title          string
	Body           string
	MergeCommitOID string
	HasTimeline    bool
}

// benchmarkGraphQLRequest represents the GraphQL request structure
type benchmarkGraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

// benchmarkGraphQLResponse represents the GraphQL response structure
type benchmarkGraphQLResponse struct {
	Data map[string]interface{} `json:"data"`
}

// NewMockGitHubServerForBench creates a new mock server configured for benchmark testing
func NewMockGitHubServerForBench(totalPRs, pageSize int, responseDelayMs int64) *MockGitHubServerForBench {
	mock := &MockGitHubServerForBench{
		totalPRs:        totalPRs,
		pageSize:        pageSize,
		responseDelayMs: responseDelayMs,
		prDatasets:      make(map[string][]benchmarkPR),
	}
	
	mock.server = httptest.NewServer(mock)
	return mock
}

// URL returns the mock server's URL
func (m *MockGitHubServerForBench) URL() string {
	return m.server.URL
}

// Close shuts down the mock server
func (m *MockGitHubServerForBench) Close() {
	m.server.Close()
}

// GetStats returns current request statistics
func (m *MockGitHubServerForBench) GetStats() (totalRequests, maxConcurrent int64) {
	return atomic.LoadInt64(&m.requestCount), atomic.LoadInt64(&m.maxConcurrent)
}

// ResetStats resets all request tracking statistics
func (m *MockGitHubServerForBench) ResetStats() {
	atomic.StoreInt64(&m.requestCount, 0)
	atomic.StoreInt64(&m.concurrentReqs, 0)
	atomic.StoreInt64(&m.maxConcurrent, 0)
}

// generatePRDataset creates a dataset of mock pull requests for the specified owner/repo
// using the enhanced generateMockPullRequests function for realistic data variation
func (m *MockGitHubServerForBench) generatePRDataset(owner, repo string, count int) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	key := fmt.Sprintf("%s/%s", owner, repo)
	prs := generateMockPullRequests(count)
	
	m.prDatasets[key] = prs
}

// SetupLargeDataset configures the mock server with a large dataset for stress testing
func (m *MockGitHubServerForBench) SetupLargeDataset(owner, repo string, sizes ...int) {
	if len(sizes) == 0 {
		sizes = []int{100, 1000, 5000, 10000}
	}
	
	maxSize := 0
	for _, size := range sizes {
		if size > maxSize {
			maxSize = size
		}
	}
	
	m.generatePRDataset(owner, repo, maxSize)
}

// createPagedResponse creates a paginated response for the given parameters
func (m *MockGitHubServerForBench) createPagedResponse(owner, repo, after string, first int) benchmarkGraphQLResponse {
	key := fmt.Sprintf("%s/%s", owner, repo)
	
	// Check if dataset exists, create if needed
	m.mutex.RLock()
	allPRs := m.prDatasets[key]
	m.mutex.RUnlock()
	
	if len(allPRs) == 0 {
		m.generatePRDataset(owner, repo, m.totalPRs)
		m.mutex.RLock()
		allPRs = m.prDatasets[key]
		m.mutex.RUnlock()
	}
	
	// Handle pagination
	startIdx := 0
	if after != "" {
		// Decode cursor (simple base64-like encoding for benchmark purposes)
		if decoded, err := strconv.Atoi(after); err == nil && decoded < len(allPRs) {
			startIdx = decoded
		}
	}
	
	// Determine page size
	pageSize := first
	if pageSize <= 0 || pageSize > m.pageSize {
		pageSize = m.pageSize
	}
	
	endIdx := startIdx + pageSize
	if endIdx > len(allPRs) {
		endIdx = len(allPRs)
	}
	
	// Create page data
	pageData := allPRs[startIdx:endIdx]
	prs := make([]interface{}, 0, len(pageData))
	
	for _, pr := range pageData {
		gqlPR := map[string]interface{}{
			"id":          pr.ID,
			"number":      pr.Number,
			"headRefName": pr.HeadRefName,
			"baseRefName": pr.BaseRefName,
			"isDraft":     pr.IsDraft,
			"permalink":   fmt.Sprintf("https://github.com/%s/%s/pull/%d", owner, repo, pr.Number),
			"state":       pr.State,
			"title":       pr.Title,
			"body":        pr.Body,
		}
		
		// Add merge commit if present
		if pr.MergeCommitOID != "" {
			gqlPR["mergeCommit"] = map[string]interface{}{
				"oid": pr.MergeCommitOID,
			}
		}
		
		// Add timeline items for closed/merged PRs
		if pr.HasTimeline {
			timelineNodes := []interface{}{}
			if pr.State == "MERGED" && pr.MergeCommitOID != "" {
				timelineNodes = append(timelineNodes, map[string]interface{}{
					"__typename": "MergedEvent",
					"commit": map[string]interface{}{
						"oid": pr.MergeCommitOID,
					},
				})
			} else if pr.State == "CLOSED" {
				timelineNodes = append(timelineNodes, map[string]interface{}{
					"__typename": "ClosedEvent",
					"closer": map[string]interface{}{
						"__typename": "Commit",
						"oid":        fmt.Sprintf("close%d", pr.Number),
					},
				})
			}
			
			gqlPR["timelineItems"] = map[string]interface{}{
				"nodes": timelineNodes,
			}
		}
		
		prs = append(prs, gqlPR)
	}
	
	// Create page info
	pageInfo := map[string]interface{}{
		"startCursor":     strconv.Itoa(startIdx),
		"endCursor":       strconv.Itoa(endIdx),
		"hasNextPage":     endIdx < len(allPRs),
		"hasPreviousPage": startIdx > 0,
	}
	
	return benchmarkGraphQLResponse{
		Data: map[string]interface{}{
			"repository": map[string]interface{}{
				"pullRequests": map[string]interface{}{
					"totalCount": int64(len(allPRs)),
					"nodes":      prs,
					"pageInfo":   pageInfo,
				},
			},
		},
	}
}

// ServeHTTP implements the http.Handler interface for the mock server
func (m *MockGitHubServerForBench) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Track concurrent requests
	current := atomic.AddInt64(&m.concurrentReqs, 1)
	defer atomic.AddInt64(&m.concurrentReqs, -1)
	
	// Update max concurrent requests if necessary
	for {
		max := atomic.LoadInt64(&m.maxConcurrent)
		if current <= max || atomic.CompareAndSwapInt64(&m.maxConcurrent, max, current) {
			break
		}
	}
	
	// Increment total request count
	atomic.AddInt64(&m.requestCount, 1)
	
	// Simulate response delay for rate limiting testing
	if m.responseDelayMs > 0 {
		time.Sleep(time.Duration(m.responseDelayMs) * time.Millisecond)
	}
	
	// Parse GraphQL request
	var req benchmarkGraphQLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}
	
	// Handle pull requests query
	if req.Query == benchmarkPRQuery {
		response := m.handlePullRequestsQuery(req)
		
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
		return
	}
	
	// Return error for unsupported queries
	http.Error(w, "Unsupported GraphQL query", http.StatusBadRequest)
}

// handlePullRequestsQuery processes pull requests GraphQL queries
func (m *MockGitHubServerForBench) handlePullRequestsQuery(req benchmarkGraphQLRequest) benchmarkGraphQLResponse {
	vars := req.Variables
	
	// Extract variables with defaults
	owner := "testowner"
	repo := "testrepo"
	after := ""
	first := m.pageSize
	
	if ownerVar, ok := vars["owner"].(string); ok {
		owner = ownerVar
	}
	if repoVar, ok := vars["repo"].(string); ok {
		repo = repoVar
	}
	if afterVar, ok := vars["after"].(string); ok {
		after = afterVar
	}
	if firstVar, ok := vars["first"]; ok {
		switch v := firstVar.(type) {
		case float64:
			first = int(v)
		case int:
			first = v
		}
	}
	
	return m.createPagedResponse(owner, repo, after, first)
}

// SetResponseDelay configures the mock server to simulate response delays
func (m *MockGitHubServerForBench) SetResponseDelay(delayMs int64) {
	m.responseDelayMs = delayMs
}

// TestGenerateMockPullRequests verifies the generateMockPullRequests function
func TestGenerateMockPullRequests(t *testing.T) {
	// Test with various counts
	testCases := []struct {
		name  string
		count int
	}{
		{"zero PRs", 0},
		{"single PR", 1},
		{"small dataset", 10},
		{"medium dataset", 100},
		{"large dataset", 1000},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			prs := generateMockPullRequests(tc.count)
			
			if tc.count == 0 {
				if prs != nil {
					t.Errorf("Expected nil for count 0, got %d PRs", len(prs))
				}
				return
			}
			
			if len(prs) != tc.count {
				t.Errorf("Expected %d PRs, got %d", tc.count, len(prs))
			}
			
			// Verify data variation
			if tc.count > 1 {
				states := make(map[string]bool)
				branches := make(map[string]bool)
				titles := make(map[string]bool)
				
				for _, pr := range prs {
					states[pr.State] = true
					branches[pr.HeadRefName] = true
					titles[pr.Title] = true
					
					// Verify required fields are populated
					if pr.ID == "" || pr.Number <= 0 || pr.HeadRefName == "" ||
						pr.BaseRefName == "" || pr.Title == "" {
						t.Errorf("PR %d missing required fields", pr.Number)
					}
					
					// Verify timeline consistency
					if pr.State != "OPEN" && !pr.HasTimeline {
						t.Errorf("PR %d should have timeline for state %s", pr.Number, pr.State)
					}
					
					// Verify merge commit for merged PRs
					if pr.State == "MERGED" && pr.MergeCommitOID == "" {
						t.Errorf("Merged PR %d missing merge commit", pr.Number)
					}
				}
				
				// Verify we have variation in data (for datasets > 10)
				if tc.count >= 10 {
					if len(states) < 2 {
						t.Errorf("Expected multiple states, got %d", len(states))
					}
					if len(branches) < 2 {
						t.Errorf("Expected multiple branches, got %d", len(branches))
					}
				}
			}
		})
	}
}

// TestCreateLargePRDataset verifies the createLargePRDataset function
func TestCreateLargePRDataset(t *testing.T) {
	mock := NewMockGitHubServerForBench(1000, 50, 0)
	defer mock.Close()
	
	owner, repo := "testowner", "testrepo"
	
	// Test with default sizes
	createLargePRDataset(mock, owner, repo)
	
	key := fmt.Sprintf("%s/%s", owner, repo)
	mock.mutex.RLock()
	prs := mock.prDatasets[key]
	mock.mutex.RUnlock()
	
	if len(prs) < 25000 { // Default max size is 25000
		t.Errorf("Expected at least 25000 PRs for stress testing, got %d", len(prs))
	}
	
	// Test with custom sizes
	customSizes := []int{500, 2000, 5000}
	createLargePRDataset(mock, "custom", "repo", customSizes...)
	
	customKey := "custom/repo"
	mock.mutex.RLock()
	customPrs := mock.prDatasets[customKey]
	mock.mutex.RUnlock()
	
	if len(customPrs) != 5000 { // Max of custom sizes
		t.Errorf("Expected 5000 PRs for custom sizes, got %d", len(customPrs))
	}
}

// TestMockGitHubServerForBench_BasicFunctionality verifies the mock server basic operations
func TestMockGitHubServerForBench_BasicFunctionality(t *testing.T) {
	// Create mock server with small dataset for testing
	mock := NewMockGitHubServerForBench(100, 25, 0)
	defer mock.Close()
	
	// Setup test data
	mock.SetupLargeDataset("testowner", "testrepo", 50)
	
	// Test basic request handling
	requestBody := `{
		"query": "query($after:String$first:Int!$owner:String!$repo:String!$states:[PullRequestState!]){repository(owner: $owner, name: $repo){pullRequests(states: $states, first: $first, after: $after){totalCount,nodes{id,number,headRefName,baseRefName,isDraft,permalink,state,title,body,mergeCommit{oid},timelineItems(last: 10, itemTypes: [CLOSED_EVENT, MERGED_EVENT]){nodes{... on ClosedEvent{closer{... on Commit{oid}}},... on MergedEvent{commit{oid}}}}},pageInfo{endCursor,hasNextPage,hasPreviousPage,startCursor}}}}",
		"variables": {
			"owner": "testowner",
			"repo": "testrepo",
			"first": 10,
			"after": null
		}
	}`
	
	resp, err := http.Post(mock.URL(), "application/json", strings.NewReader(requestBody))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}
	
	var response benchmarkGraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	// Verify response structure
	if response.Data == nil {
		t.Fatal("Response data is nil")
	}
	
	totalReqs, maxConcurrent := mock.GetStats()
	if totalReqs != 1 {
		t.Errorf("Expected 1 request, got %d", totalReqs)
	}
	if maxConcurrent != 1 {
		t.Errorf("Expected max concurrent 1, got %d", maxConcurrent)
	}
}

// generateMockPullRequests creates a slice of realistic mock pull requests for benchmarking.
// This function generates varied PR data to prevent unrealistic caching effects and includes
// realistic field values that mirror production GitHub API responses.
func generateMockPullRequests(count int) []benchmarkPR {
	if count <= 0 {
		return nil
	}
	
	prs := make([]benchmarkPR, 0, count)
	
	// Realistic variations for PR data
	states := []string{"OPEN", "CLOSED", "MERGED"}
	stateWeights := []int{40, 30, 30} // 40% open, 30% closed, 30% merged
	
	headBranches := []string{
		"feature/user-authentication", "bugfix/memory-leak", "feat/api-versioning",
		"hotfix/security-patch", "develop", "staging", "feature/dark-mode",
		"bugfix/validation-error", "feat/performance-optimization", "feature/dashboard-ui",
		"bugfix/race-condition", "feat/webhook-integration", "feature/mobile-responsive",
		"hotfix/critical-fix", "feat/logging-improvements", "feature/search-functionality",
		"bugfix/timeout-handling", "feat/caching-layer", "feature/user-preferences",
		"bugfix/data-corruption", "feat/monitoring-metrics", "feature/admin-panel",
	}
	
	baseBranches := []string{"main", "master", "develop", "staging"}
	baseWeights := []int{60, 25, 10, 5} // 60% main, 25% master, 10% develop, 5% staging
	
	titles := []string{
		"Add user authentication system", "Fix memory leak in data processor",
		"Implement API versioning support", "Security patch for XSS vulnerability",
		"Add dark mode support", "Fix validation error in form handling",
		"Optimize database query performance", "Update dashboard UI components",
		"Fix race condition in concurrent processing", "Integrate webhook notifications",
		"Make interface mobile responsive", "Critical fix for production issue",
		"Improve logging and error handling", "Implement search functionality",
		"Fix timeout handling in HTTP client", "Add Redis caching layer",
		"Create user preferences system", "Fix data corruption in migration",
		"Add monitoring and metrics collection", "Build admin control panel",
		"Refactor legacy authentication code", "Update dependencies and security patches",
		"Implement real-time notifications", "Add comprehensive test coverage",
		"Optimize image loading and processing", "Fix cross-browser compatibility issues",
		"Add internationalization support", "Implement feature flag system",
		"Fix SSL certificate validation", "Add automated backup system",
	}
	
	bodyTemplates := []string{
		"This PR addresses %s by implementing comprehensive changes.\n\n## Changes\n- Added %s functionality\n- Updated related tests\n- Fixed documentation\n- Improved error handling\n\n## Testing\n- Unit tests passing\n- Integration tests added\n- Manual testing completed",
		"## Summary\nThis change %s to improve system %s.\n\n## Implementation Details\n- Modified core %s logic\n- Added validation for %s scenarios\n- Updated API documentation\n\n## Breaking Changes\nNone\n\n## Performance Impact\nMinimal performance improvement expected",
		"**Problem**: %s was causing issues in production.\n\n**Solution**: Implemented %s approach using best practices.\n\n**Changes Made**:\n1. Refactored %s component\n2. Added comprehensive error handling\n3. Improved logging and monitoring\n4. Updated configuration management\n\n**Testing Strategy**:\n- Automated tests cover all scenarios\n- Load testing performed\n- Security review completed",
	}
	
	// Generate varied PRs with realistic data distribution
	for i := 0; i < count; i++ {
		// Select state based on realistic weights
		stateIdx := weightedRandom(stateWeights, i)
		state := states[stateIdx]
		
		// Select base branch based on weights
		baseIdx := weightedRandom(baseWeights, i+count)
		baseBranch := baseBranches[baseIdx]
		
		// Select head branch (avoid same as base)
		headBranch := headBranches[i%len(headBranches)]
		for headBranch == baseBranch && len(headBranches) > 1 {
			headBranch = headBranches[(i+1)%len(headBranches)]
		}
		
		// Generate realistic titles and descriptions
		title := titles[i%len(titles)]
		if i >= len(titles) {
			title = fmt.Sprintf("%s (v%d)", title, (i/len(titles))+1)
		}
		
		bodyTemplate := bodyTemplates[i%len(bodyTemplates)]
		feature := []string{"authentication", "validation", "processing", "UI", "API", "security", "performance", "monitoring"}[i%8]
		system := []string{"core", "frontend", "backend", "database", "cache", "network", "storage", "logging"}[i%8]
		body := fmt.Sprintf(bodyTemplate, feature, system, feature, system)
		
		// Create PR with variation to prevent caching effects
		pr := benchmarkPR{
			ID:          fmt.Sprintf("PR_%08d_%016d", i+1, hashString(title)%10000000000000000),
			Number:      i + 1,
			HeadRefName: headBranch,
			BaseRefName: baseBranch,
			IsDraft:     (i+7)%13 == 0, // ~7.7% are drafts (realistic percentage)
			State:       state,
			Title:       title,
			Body:        body,
			HasTimeline: state != "OPEN",
		}
		
		// Add merge commit for merged PRs with varied commit SHAs
		if state == "MERGED" {
			pr.MergeCommitOID = fmt.Sprintf("%040x", hashString(fmt.Sprintf("merge_%d_%s", i+1, title)))
		}
		
		prs = append(prs, pr)
	}
	
	return prs
}

// createLargePRDataset is an enhanced utility function for stress testing with 10k+ pull requests.
// This function supports multiple dataset sizes and implements data variation strategies
// to prevent unrealistic caching effects during benchmarking.
func createLargePRDataset(mock *MockGitHubServerForBench, owner, repo string, sizes ...int) {
	if len(sizes) == 0 {
		// Default sizes for comprehensive stress testing
		sizes = []int{100, 1000, 5000, 10000, 25000}
	}
	
	// Find the maximum size needed
	maxSize := 0
	for _, size := range sizes {
		if size > maxSize {
			maxSize = size
		}
	}
	
	// Generate the large dataset using the enhanced generation function
	prs := generateMockPullRequests(maxSize)
	
	// Store the dataset in the mock server
	key := fmt.Sprintf("%s/%s", owner, repo)
	mock.mutex.Lock()
	mock.prDatasets[key] = prs
	mock.mutex.Unlock()
}

// weightedRandom selects an index based on weights, using seed for deterministic variation
func weightedRandom(weights []int, seed int) int {
	total := 0
	for _, w := range weights {
		total += w
	}
	
	// Use seed to generate deterministic but varied selection
	target := (hashString(fmt.Sprintf("seed_%d", seed)) % uint64(total))
	
	current := uint64(0)
	for i, weight := range weights {
		current += uint64(weight)
		if target < current {
			return i
		}
	}
	
	return len(weights) - 1 // Fallback
}

// hashString generates a simple hash of a string for deterministic variation
func hashString(s string) uint64 {
	hash := uint64(5381)
	for _, c := range s {
		hash = hash*33 + uint64(c)
	}
	return hash
}

// BenchmarkRepoPullRequests_ConcurrentThroughput measures throughput under concurrent load conditions.
// This benchmark will be implemented to test varying concurrency levels and measure
// requests per second, average response time, and error rates.
func BenchmarkRepoPullRequests_ConcurrentThroughput(b *testing.B) {
	// Placeholder for concurrent throughput benchmark implementation
	// Will be implemented in Step 2.1
	b.Skip("Implementation pending - Step 2.1")
}

// BenchmarkRepoPullRequests_PaginationThroughput measures pagination performance with large result sets.
// This benchmark will test pagination through datasets of varying sizes and measure
// time to paginate through entire result sets and memory allocation per page.
func BenchmarkRepoPullRequests_PaginationThroughput(b *testing.B) {
	// Placeholder for pagination throughput benchmark implementation
	// Will be implemented in Step 2.2
	b.Skip("Implementation pending - Step 2.2")
}

// BenchmarkRepoPullRequests_RateLimitHandling measures performance under rate limiting conditions.
// This benchmark will test exponential backoff performance, retry logic efficiency,
// and measure degraded performance metrics under various error conditions.
func BenchmarkRepoPullRequests_RateLimitHandling(b *testing.B) {
	// Placeholder for rate limiting benchmark implementation
	// Will be implemented in Step 2.3
	b.Skip("Implementation pending - Step 2.3")
}