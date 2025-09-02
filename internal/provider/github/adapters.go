package github

import (
	"github.com/aviator-co/av/internal/provider"
	"github.com/shurcooL/githubv4"
)

// Adapter functions to convert between GitHub-specific types and provider types

// ConvertGithubPRToProvider converts a GitHub pull request to provider format
func ConvertGithubPRToProvider(ghPR *PullRequest) *provider.PullRequest {
	if ghPR == nil {
		return nil
	}
	return convertGithubPRToProvider((*githubPullRequest)(ghPR))
}

// ConvertProviderPRToGithub converts a provider pull request to GitHub format
func ConvertProviderPRToGithub(providerPR *provider.PullRequest) *PullRequest {
	if providerPR == nil {
		return nil
	}
	
	ghState := githubv4.PullRequestStateOpen
	switch providerPR.State {
	case provider.PullRequestStateOpen:
		ghState = githubv4.PullRequestStateOpen
	case provider.PullRequestStateClosed:
		ghState = githubv4.PullRequestStateClosed
	case provider.PullRequestStateMerged:
		ghState = githubv4.PullRequestStateMerged
	}

	ghPR := &githubPullRequest{
		ID:          providerPR.ID,
		Number:      providerPR.Number,
		HeadRefName: providerPR.HeadRefName,
		BaseRefName: providerPR.BaseRefName,
		IsDraft:     providerPR.IsDraft,
		Permalink:   providerPR.Permalink,
		State:       ghState,
		Title:       providerPR.Title,
		Body:        providerPR.Body,
	}
	
	// Set merge commit if available
	if providerPR.MergeCommit != "" {
		ghPR.PRIVATE_MergeCommit.Oid = providerPR.MergeCommit
	}
	
	return (*PullRequest)(ghPR)
}

// ConvertGithubStateToProvider converts GitHub PR state to provider state
func ConvertGithubStateToProvider(ghState githubv4.PullRequestState) provider.PullRequestState {
	return convertGithubStateToProvider(ghState)
}

// ConvertProviderStateToGithub converts provider PR state to GitHub state
func ConvertProviderStateToGithub(providerState provider.PullRequestState) (githubv4.PullRequestState, error) {
	return convertProviderStateToGithub(providerState)
}

// ConvertGithubRepoToProvider converts a GitHub repository to provider format
func ConvertGithubRepoToProvider(ghRepo *Repository) *provider.Repository {
	if ghRepo == nil {
		return nil
	}
	return &provider.Repository{
		ID:    ghRepo.ID,
		Owner: ghRepo.Owner.Login,
		Name:  ghRepo.Name,
	}
}

// ConvertGithubUserToProvider converts a GitHub user to provider format
func ConvertGithubUserToProvider(ghUser *User) *provider.User {
	if ghUser == nil {
		return nil
	}
	return &provider.User{
		ID:    string(ghUser.ID),
		Login: ghUser.Login,
	}
}

// ConvertGithubViewerToProvider converts a GitHub viewer to provider format
func ConvertGithubViewerToProvider(ghViewer *Viewer) *provider.Viewer {
	if ghViewer == nil {
		return nil
	}
	return &provider.Viewer{
		Name:  ghViewer.Name,
		Login: ghViewer.Login,
	}
}

// ConvertGithubPageInfoToProvider converts GitHub page info to provider format
func ConvertGithubPageInfoToProvider(ghPageInfo *PageInfo) provider.PageInfo {
	return provider.PageInfo{
		HasNextPage:     ghPageInfo.HasNextPage,
		HasPreviousPage: ghPageInfo.HasPreviousPage,
		StartCursor:     ghPageInfo.StartCursor,
		EndCursor:       ghPageInfo.EndCursor,
	}
}

// ConvertProviderCreateInputToGithub converts provider create input to GitHub format
func ConvertProviderCreateInputToGithub(input provider.CreatePullRequestInput) githubv4.CreatePullRequestInput {
	return githubv4.CreatePullRequestInput{
		RepositoryID: githubv4.ID(input.RepositoryID),
		Title:        githubv4.String(input.Title),
		Body:         nullable(githubv4.String(input.Body)),
		HeadRefName:  githubv4.String(input.HeadRefName),
		BaseRefName:  githubv4.String(input.BaseRefName),
		Draft:        nullable(githubv4.Boolean(input.IsDraft)),
	}
}

// ConvertProviderUpdateInputToGithub converts provider update input to GitHub format
func ConvertProviderUpdateInputToGithub(input provider.UpdatePullRequestInput) githubv4.UpdatePullRequestInput {
	ghInput := githubv4.UpdatePullRequestInput{
		PullRequestID: githubv4.ID(input.PullRequestID),
	}
	
	if input.Title != nil {
		ghInput.Title = nullable(githubv4.String(*input.Title))
	}
	if input.Body != nil {
		ghInput.Body = nullable(githubv4.String(*input.Body))
	}
	
	return ghInput
}

// ConvertProviderGetPRsInputToGithub converts provider get PRs input to GitHub format
func ConvertProviderGetPRsInputToGithub(input provider.GetPullRequestsInput) GetPullRequestsInput {
	var ghStates []githubv4.PullRequestState
	for _, state := range input.States {
		if ghState, err := convertProviderStateToGithub(state); err == nil {
			ghStates = append(ghStates, ghState)
		}
	}
	
	return GetPullRequestsInput{
		Owner:       input.Owner,
		Repo:        input.Repo,
		HeadRefName: input.HeadRefName,
		BaseRefName: input.BaseRefName,
		States:      ghStates,
		First:       input.First,
		After:       input.After,
	}
}

// ConvertGithubGetPRsPageToProvider converts GitHub get PRs page to provider format
func ConvertGithubGetPRsPageToProvider(ghPage *GetPullRequestsPage) *provider.GetPullRequestsPage {
	if ghPage == nil {
		return nil
	}
	
	var providerPRs []provider.PullRequest
	for _, ghPR := range ghPage.PullRequests {
		providerPRs = append(providerPRs, *convertGithubPRToProvider((*githubPullRequest)(&ghPR)))
	}
	
	return &provider.GetPullRequestsPage{
		PageInfo:     ConvertGithubPageInfoToProvider(&ghPage.PageInfo),
		PullRequests: providerPRs,
	}
}