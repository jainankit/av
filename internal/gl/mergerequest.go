package gl

// MergeRequest represents a GitLab merge request.
// This is a placeholder implementation that will be expanded in later steps.
type MergeRequest struct {
	ID          string
	IID         int64
	Title       string
	Description string
	State       string
	WebURL      string
	
	SourceBranch string
	TargetBranch string
	
	Author *User
	
	// Additional fields will be added as needed
}