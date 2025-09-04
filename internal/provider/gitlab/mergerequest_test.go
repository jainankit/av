package gitlab

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xanzy/go-gitlab"
)

func TestParseProjectID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{
			name:     "numeric project ID",
			input:    "123",
			expected: 123,
		},
		{
			name:     "project path",
			input:    "group/project",
			expected: "group/project",
		},
		{
			name:     "complex project path",
			input:    "group/subgroup/project",
			expected: "group/subgroup/project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseProjectID(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseMergeRequestID(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectProject interface{}
		expectIID     int
		expectError   bool
	}{
		{
			name:          "numeric project ID with IID",
			input:         "123:456",
			expectProject: 123,
			expectIID:     456,
			expectError:   false,
		},
		{
			name:          "project path with IID",
			input:         "group/project:789",
			expectProject: "group/project",
			expectIID:     789,
			expectError:   false,
		},
		{
			name:        "invalid format - no colon",
			input:       "123456",
			expectError: true,
		},
		{
			name:        "invalid format - non-numeric IID",
			input:       "123:abc",
			expectError: true,
		},
		{
			name:        "empty input",
			input:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project, iid, err := parseMergeRequestID(tt.input)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectProject, project)
				assert.Equal(t, tt.expectIID, iid)
			}
		})
	}
}

func TestConvertMergeRequestToPullRequest(t *testing.T) {
	client := &Client{
		baseURL: "https://gitlab.example.com",
	}

	tests := []struct {
		name     string
		mr       *gitlab.MergeRequest
		expected string // expected state
	}{
		{
			name: "open merge request",
			mr: &gitlab.MergeRequest{
				ID:           123,
				IID:          456,
				Title:        "Test MR",
				Description:  "Test description",
				State:        "opened",
				SourceBranch: "feature",
				TargetBranch: "main",
				WebURL:       "https://gitlab.example.com/project/merge_requests/456",
				Draft:        false,
			},
			expected: "OPEN",
		},
		{
			name: "closed merge request",
			mr: &gitlab.MergeRequest{
				ID:           124,
				IID:          457,
				Title:        "Closed MR",
				Description:  "Closed description",
				State:        "closed",
				SourceBranch: "feature2",
				TargetBranch: "main",
				WebURL:       "https://gitlab.example.com/project/merge_requests/457",
				Draft:        false,
			},
			expected: "CLOSED",
		},
		{
			name: "merged merge request",
			mr: &gitlab.MergeRequest{
				ID:             125,
				IID:            458,
				Title:          "Merged MR",
				Description:    "Merged description",
				State:          "merged",
				SourceBranch:   "feature3",
				TargetBranch:   "main",
				WebURL:         "https://gitlab.example.com/project/merge_requests/458",
				Draft:          false,
				MergeCommitSHA: "abc123",
			},
			expected: "MERGED",
		},
		{
			name: "draft merge request",
			mr: &gitlab.MergeRequest{
				ID:           126,
				IID:          459,
				Title:        "Draft: WIP Feature",
				Description:  "Work in progress",
				State:        "opened",
				SourceBranch: "wip-feature",
				TargetBranch: "main",
				WebURL:       "https://gitlab.example.com/project/merge_requests/459",
				Draft:        true,
			},
			expected: "OPEN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.convertMergeRequestToPullRequest(tt.mr, 789)
			
			assert.Equal(t, "789:456", result.ID)
			assert.Equal(t, int64(tt.mr.IID), result.Number)
			assert.Equal(t, tt.mr.SourceBranch, result.HeadRefName)
			assert.Equal(t, tt.mr.TargetBranch, result.BaseRefName)
			assert.Equal(t, tt.mr.Title, result.Title)
			assert.Equal(t, tt.mr.Description, result.Body)
			assert.Equal(t, tt.expected, result.State)
			assert.Equal(t, tt.mr.WebURL, result.Permalink)
			
			if tt.mr.Draft || tt.mr.Title == "Draft: WIP Feature" {
				assert.True(t, result.IsDraft)
			}
			
			if tt.mr.MergeCommitSHA != "" {
				assert.Equal(t, tt.mr.MergeCommitSHA, result.MergeCommitSHA)
			}
		})
	}
}

func TestMRMetadataHandling(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected MRMetadata
	}{
		{
			name: "basic metadata",
			input: `Some description

<!-- av mr metadata
` + "```json\n" + `{"parent":"main","parentHead":"abc123","trunk":"main"}
` + "```\n" + `-->
`,
			expected: MRMetadata{
				Parent:     "main",
				ParentHead: "abc123",
				Trunk:      "main",
			},
		},
		{
			name: "metadata with parent MR",
			input: `Feature description

<!-- av mr metadata
` + "```json\n" + `{"parent":"feature-base","parentHead":"def456","parentMR":123,"trunk":"main"}
` + "```\n" + `-->
`,
			expected: MRMetadata{
				Parent:     "feature-base",
				ParentHead: "def456",
				ParentMR:   123,
				Trunk:      "main",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := ReadMRMetadata(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, metadata)
		})
	}
}

func TestAddMRMetadataAndStack(t *testing.T) {
	metadata := MRMetadata{
		Parent:     "feature-base",
		ParentHead: "abc123",
		ParentMR:   456,
		Trunk:      "main",
	}

	t.Run("add metadata without stack", func(t *testing.T) {
		body := "This is a test merge request"
		result := AddMRMetadataAndStack(body, metadata, "feature-branch", false)
		
		assert.Contains(t, result, "This is a test merge request")
		assert.Contains(t, result, "av mr metadata")
		assert.Contains(t, result, `"parent":"feature-base"`)
		assert.Contains(t, result, `"parentMR":456`)
		assert.NotContains(t, result, "av mr stack begin")
	})

	t.Run("add metadata with stack", func(t *testing.T) {
		body := "This is a test merge request with stack"
		result := AddMRMetadataAndStack(body, metadata, "feature-branch", true)
		
		assert.Contains(t, result, "This is a test merge request with stack")
		assert.Contains(t, result, "av mr metadata")
		assert.Contains(t, result, "av mr stack begin")
		assert.Contains(t, result, "Depends on !456")
		assert.Contains(t, result, "Aviator")
	})

	t.Run("update existing metadata", func(t *testing.T) {
		existingBody := `Original description

<!-- av mr metadata
` + "```json\n" + `{"parent":"old-parent","parentHead":"old123","trunk":"main"}
` + "```\n" + `-->
`
		result := AddMRMetadataAndStack(existingBody, metadata, "feature-branch", false)
		
		assert.Contains(t, result, "Original description")
		assert.Contains(t, result, `"parent":"feature-base"`)
		assert.Contains(t, result, `"parentMR":456`)
		assert.NotContains(t, result, "old-parent")
	})
}

func TestExtractContent(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		start          string
		end            string
		expectedContent string
		expectedOutput  string
	}{
		{
			name:           "extract simple content",
			input:          "before <!-- start --> content <!-- end --> after",
			start:          "<!-- start -->",
			end:            "<!-- end -->",
			expectedContent: "content",
			expectedOutput:  "before\nafter",
		},
		{
			name:           "no start marker",
			input:          "some text without markers",
			start:          "<!-- start -->",
			end:            "<!-- end -->",
			expectedContent: "",
			expectedOutput:  "some text without markers",
		},
		{
			name:           "start but no end marker",
			input:          "before <!-- start --> content without end",
			start:          "<!-- start -->",
			end:            "<!-- end -->",
			expectedContent: "",
			expectedOutput:  "before <!-- start --> content without end",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, output := extractContent(tt.input, tt.start, tt.end)
			assert.Equal(t, tt.expectedContent, content)
			assert.Equal(t, tt.expectedOutput, output)
		})
	}
}