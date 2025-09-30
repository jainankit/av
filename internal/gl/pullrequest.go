package gl

import (
	"context"
	"strings"

	"emperror.dev/errors"
)

// MergeRequestState represents the different states a GitLab merge request can be in
type MergeRequestState string

const (
	MergeRequestStateOpened MergeRequestState = "opened"
	MergeRequestStateClosed MergeRequestState = "closed"
	MergeRequestStateMerged MergeRequestState = "merged"
)

type MergeRequest struct {
	ID              string
	Iid             int64
	SourceBranch    string
	TargetBranch    string
	State           MergeRequestState
	Title           string
	Description     string
	WebUrl          string
	Draft           bool
	WorkInProgress  bool
	MergeCommitSha  string
}

func (mr *MergeRequest) HeadBranchName() string {
	// GitLab branch names don't include the "refs/heads/" prefix in merge requests
	return strings.TrimPrefix(mr.SourceBranch, "refs/heads/")
}

func (mr *MergeRequest) BaseBranchName() string {
	return strings.TrimPrefix(mr.TargetBranch, "refs/heads/")
}

func (mr *MergeRequest) GetMergeCommit() string {
	if mr.State == MergeRequestStateMerged {
		return mr.MergeCommitSha
	}
	return ""
}

type MergeRequestOpts struct {
	ProjectPath string
	Iid         int64
}

func (c *Client) MergeRequest(ctx context.Context, id string) (*MergeRequest, error) {
	var query struct {
		MergeRequest struct {
			ID              string            `graphql:"id"`
			Iid             int64             `graphql:"iid"`
			SourceBranch    string            `graphql:"sourceBranch"`
			TargetBranch    string            `graphql:"targetBranch"`
			State           MergeRequestState `graphql:"state"`
			Title           string            `graphql:"title"`
			Description     string            `graphql:"description"`
			WebUrl          string            `graphql:"webUrl"`
			Draft           bool              `graphql:"draft"`
			WorkInProgress  bool              `graphql:"workInProgress"`
			MergeCommitSha  string            `graphql:"mergeCommitSha"`
		} `graphql:"mergeRequest(id: $id)"`
	}
	if err := c.query(ctx, &query, map[string]interface{}{
		"id": id,
	}); err != nil {
		return nil, errors.Wrap(err, "failed to query merge request")
	}
	if query.MergeRequest.ID == "" {
		return nil, errors.Errorf("merge request %q not found", id)
	}
	return &MergeRequest{
		ID:              query.MergeRequest.ID,
		Iid:             query.MergeRequest.Iid,
		SourceBranch:    query.MergeRequest.SourceBranch,
		TargetBranch:    query.MergeRequest.TargetBranch,
		State:           query.MergeRequest.State,
		Title:           query.MergeRequest.Title,
		Description:     query.MergeRequest.Description,
		WebUrl:          query.MergeRequest.WebUrl,
		Draft:           query.MergeRequest.Draft,
		WorkInProgress:  query.MergeRequest.WorkInProgress,
		MergeCommitSha:  query.MergeRequest.MergeCommitSha,
	}, nil
}

type GetMergeRequestsInput struct {
	// REQUIRED
	ProjectPath string
	// OPTIONAL
	SourceBranch string
	TargetBranch string
	State        MergeRequestState
	First        int32
	After        string
}

type GetMergeRequestsPage struct {
	PageInfo
	MergeRequests []MergeRequest
}

func (c *Client) GetMergeRequests(
	ctx context.Context,
	input GetMergeRequestsInput,
) (*GetMergeRequestsPage, error) {
	if input.First == 0 {
		input.First = 50
	}
	var query struct {
		Project struct {
			MergeRequests struct {
				Nodes []struct {
					ID              string            `graphql:"id"`
					Iid             int64             `graphql:"iid"`
					SourceBranch    string            `graphql:"sourceBranch"`
					TargetBranch    string            `graphql:"targetBranch"`
					State           MergeRequestState `graphql:"state"`
					Title           string            `graphql:"title"`
					Description     string            `graphql:"description"`
					WebUrl          string            `graphql:"webUrl"`
					Draft           bool              `graphql:"draft"`
					WorkInProgress  bool              `graphql:"workInProgress"`
					MergeCommitSha  string            `graphql:"mergeCommitSha"`
				}
				PageInfo PageInfo
			} `graphql:"mergeRequests(sourceBranches: $sourceBranches, targetBranches: $targetBranches, state: $state, first: $first, after: $after)"`
		} `graphql:"project(fullPath: $projectPath)"`
	}
	
	variables := map[string]interface{}{
		"projectPath": input.ProjectPath,
		"first":       int32(input.First),
		"after":       nullable(input.After),
	}
	
	if input.SourceBranch != "" {
		variables["sourceBranches"] = []string{input.SourceBranch}
	}
	if input.TargetBranch != "" {
		variables["targetBranches"] = []string{input.TargetBranch}
	}
	if input.State != "" {
		variables["state"] = input.State
	}
	
	if err := c.query(ctx, &query, variables); err != nil {
		return nil, errors.Wrap(err, "failed to query merge requests")
	}
	
	mergeRequests := make([]MergeRequest, len(query.Project.MergeRequests.Nodes))
	for i, node := range query.Project.MergeRequests.Nodes {
		mergeRequests[i] = MergeRequest{
			ID:              node.ID,
			Iid:             node.Iid,
			SourceBranch:    node.SourceBranch,
			TargetBranch:    node.TargetBranch,
			State:           node.State,
			Title:           node.Title,
			Description:     node.Description,
			WebUrl:          node.WebUrl,
			Draft:           node.Draft,
			WorkInProgress:  node.WorkInProgress,
			MergeCommitSha:  node.MergeCommitSha,
		}
	}
	
	return &GetMergeRequestsPage{
		PageInfo:      query.Project.MergeRequests.PageInfo,
		MergeRequests: mergeRequests,
	}, nil
}

type CreateMergeRequestInput struct {
	ProjectPath         string  `json:"projectPath"`
	Title               string  `json:"title"`
	Description         *string `json:"description,omitempty"`
	SourceBranch        string  `json:"sourceBranch"`
	TargetBranch        string  `json:"targetBranch"`
	RemoveSourceBranch  *bool   `json:"removeSourceBranch,omitempty"`
	Squash              *bool   `json:"squash,omitempty"`
}

func (c *Client) CreateMergeRequest(
	ctx context.Context,
	input CreateMergeRequestInput,
) (*MergeRequest, error) {
	var mutation struct {
		MergeRequestCreate struct {
			MergeRequest struct {
				ID              string            `graphql:"id"`
				Iid             int64             `graphql:"iid"`
				SourceBranch    string            `graphql:"sourceBranch"`
				TargetBranch    string            `graphql:"targetBranch"`
				State           MergeRequestState `graphql:"state"`
				Title           string            `graphql:"title"`
				Description     string            `graphql:"description"`
				WebUrl          string            `graphql:"webUrl"`
				Draft           bool              `graphql:"draft"`
				WorkInProgress  bool              `graphql:"workInProgress"`
				MergeCommitSha  string            `graphql:"mergeCommitSha"`
			}
		} `graphql:"mergeRequestCreate(input: $input)"`
	}
	
	if err := c.mutate(ctx, &mutation, map[string]interface{}{
		"input": input,
	}); err != nil {
		return nil, errors.Wrap(err, "failed to create merge request: gitlab error")
	}
	
	mr := &mutation.MergeRequestCreate.MergeRequest
	return &MergeRequest{
		ID:              mr.ID,
		Iid:             mr.Iid,
		SourceBranch:    mr.SourceBranch,
		TargetBranch:    mr.TargetBranch,
		State:           mr.State,
		Title:           mr.Title,
		Description:     mr.Description,
		WebUrl:          mr.WebUrl,
		Draft:           mr.Draft,
		WorkInProgress:  mr.WorkInProgress,
		MergeCommitSha:  mr.MergeCommitSha,
	}, nil
}

type UpdateMergeRequestInput struct {
	ProjectPath string  `json:"projectPath"`
	Iid         int64   `json:"iid"`
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
}

func (c *Client) UpdateMergeRequest(
	ctx context.Context,
	input UpdateMergeRequestInput,
) (*MergeRequest, error) {
	var mutation struct {
		MergeRequestUpdate struct {
			MergeRequest struct {
				ID              string            `graphql:"id"`
				Iid             int64             `graphql:"iid"`
				SourceBranch    string            `graphql:"sourceBranch"`
				TargetBranch    string            `graphql:"targetBranch"`
				State           MergeRequestState `graphql:"state"`
				Title           string            `graphql:"title"`
				Description     string            `graphql:"description"`
				WebUrl          string            `graphql:"webUrl"`
				Draft           bool              `graphql:"draft"`
				WorkInProgress  bool              `graphql:"workInProgress"`
				MergeCommitSha  string            `graphql:"mergeCommitSha"`
			}
		} `graphql:"mergeRequestUpdate(input: $input)"`
	}
	
	if err := c.mutate(ctx, &mutation, map[string]interface{}{
		"input": input,
	}); err != nil {
		return nil, errors.Wrap(err, "failed to update merge request: gitlab error")
	}
	
	mr := &mutation.MergeRequestUpdate.MergeRequest
	return &MergeRequest{
		ID:              mr.ID,
		Iid:             mr.Iid,
		SourceBranch:    mr.SourceBranch,
		TargetBranch:    mr.TargetBranch,
		State:           mr.State,
		Title:           mr.Title,
		Description:     mr.Description,
		WebUrl:          mr.WebUrl,
		Draft:           mr.Draft,
		WorkInProgress:  mr.WorkInProgress,
		MergeCommitSha:  mr.MergeCommitSha,
	}, nil
}