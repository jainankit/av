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
func (m *MockGitHubServerForBench) generatePRDataset(owner, repo string, count int) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	key := fmt.Sprintf("%s/%s", owner, repo)
	prs := make([]benchmarkPR, 0, count)
	
	states := []string{"OPEN", "CLOSED", "MERGED"}
	branches := []string{"feature/auth", "bugfix/validation", "feat/api", "hotfix/security", "develop", "staging"}
	
	for i := 0; i < count; i++ {
		stateIdx := i % len(states)
		branchIdx := i % len(branches)
		
		pr := benchmarkPR{
			ID:          fmt.Sprintf("PR_%d_%s_%s", i+1, owner, repo),
			Number:      i + 1,
			HeadRefName: branches[branchIdx],
			BaseRefName: "main",
			IsDraft:     i%10 == 0, // 10% are drafts
			State:       states[stateIdx],
			Title:       fmt.Sprintf("Pull request #%d: Implement feature %d", i+1, (i%20)+1),
			Body:        fmt.Sprintf("This PR implements feature %d with comprehensive changes.\n\nChanges include:\n- Added new functionality\n- Updated tests\n- Fixed documentation", (i%20)+1),
			HasTimeline: states[stateIdx] != "OPEN",
		}
		
		// Add merge commit for merged PRs
		if pr.State == "MERGED" {
			pr.MergeCommitOID = fmt.Sprintf("abc%04d%04d", i+1, len(pr.Title))
		}
		
		prs = append(prs, pr)
	}
	
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

// createLargePRDataset is a utility function to create large datasets for stress testing
func createLargePRDataset(mock *MockGitHubServerForBench, owner, repo string, size int) {
	mock.SetupLargeDataset(owner, repo, size)
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