package provider

// PullRequest represents a pull request/merge request across different providers
type PullRequest struct {
	ID          string
	Number      int64
	HeadRefName string
	BaseRefName string
	IsDraft     bool
	Permalink   string
	State       string
	Title       string
	Body        string
	MergeCommitSHA string
}

// HeadBranchName returns the head branch name without refs/heads/ prefix
func (p *PullRequest) HeadBranchName() string {
	return trimRefPrefix(p.HeadRefName)
}

// BaseBranchName returns the base branch name without refs/heads/ prefix  
func (p *PullRequest) BaseBranchName() string {
	return trimRefPrefix(p.BaseRefName)
}

func trimRefPrefix(ref string) string {
	if ref == "" {
		return ref
	}
	if len(ref) > 11 && ref[:11] == "refs/heads/" {
		return ref[11:]
	}
	return ref
}

// Repository represents a repository across different providers
type Repository struct {
	ID    string
	Owner string
	Name  string
}

// User represents a user across different providers
type User struct {
	ID    string
	Login string
}

// Viewer represents the authenticated user across different providers
type Viewer struct {
	Name  string
	Login string
}

// Pull request states
const (
	PullRequestStateOpen   = "OPEN"
	PullRequestStateClosed = "CLOSED"
	PullRequestStateMerged = "MERGED"
)