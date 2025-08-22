package glui

import (
	"context"
	"strings"

	"emperror.dev/errors"
	"github.com/aviator-co/av/internal/actions"
	"github.com/aviator-co/av/internal/gl"
	"github.com/aviator-co/av/internal/git"
	"github.com/aviator-co/av/internal/meta"
	"github.com/aviator-co/av/internal/utils/colors"
	"github.com/aviator-co/av/internal/utils/stackutils"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func NewGitLabFetchModel(
	repo *git.Repo,
	db meta.DB,
	client *gl.Client,
	currentBranch plumbing.ReferenceName,
	targetBranches []plumbing.ReferenceName,
) *GitLabFetchModel {
	return &GitLabFetchModel{
		repo:           repo,
		db:             db,
		client:         client,
		currentBranch:  currentBranch,
		targetBranches: targetBranches,
		spinner:        spinner.New(spinner.WithSpinner(spinner.Dot)),

		runningGitFetch:             true,
		runningGitLabAPIBranch:      -1,
		runningCheckCommitHistory:   false,
		runningPropagateMergeCommit: false,
	}
}

type GitLabFetchProgress struct {
	gitFetchIsDone               bool
	apiFetchIsDone               bool
	checkCommitHistoryIsDone     bool
	mergeCommitPropagationIsDone bool
}

type GitLabFetchDone struct{}

type GitLabFetchModel struct {
	repo           *git.Repo
	db             meta.DB
	client         *gl.Client
	currentBranch  plumbing.ReferenceName
	targetBranches []plumbing.ReferenceName
	spinner        spinner.Model

	runningGitFetch             bool
	runningGitLabAPIBranch      int
	runningCheckCommitHistory   bool
	runningPropagateMergeCommit bool
}

func (vm *GitLabFetchModel) Init() tea.Cmd {
	return tea.Batch(vm.spinner.Tick, vm.runGitFetch)
}

func (vm *GitLabFetchModel) Update(msg tea.Msg) (*GitLabFetchModel, tea.Cmd) {
	switch msg := msg.(type) {
	case *GitLabFetchProgress:
		if msg.gitFetchIsDone {
			vm.runningGitFetch = false
			vm.runningGitLabAPIBranch = 0
			return vm, vm.runGitLabAPIFetch
		}
		if msg.apiFetchIsDone {
			vm.runningGitLabAPIBranch++
			if len(vm.targetBranches) <= vm.runningGitLabAPIBranch {
				vm.runningCheckCommitHistory = true
				return vm, vm.updateMergeCommitsFromCommitMessage
			}
			return vm, vm.runGitLabAPIFetch
		}
		if msg.checkCommitHistoryIsDone {
			vm.runningCheckCommitHistory = false
			vm.runningPropagateMergeCommit = true
			return vm, vm.updateMergeCommitsFromChildren
		}
		if msg.mergeCommitPropagationIsDone {
			vm.runningPropagateMergeCommit = false
			return vm, func() tea.Msg { return &GitLabFetchDone{} }
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		vm.spinner, cmd = vm.spinner.Update(msg)
		return vm, cmd
	}
	return vm, nil
}

func (vm *GitLabFetchModel) View() string {
	sb := strings.Builder{}
	showTree := false
	if vm.runningGitFetch {
		sb.WriteString(colors.ProgressStyle.Render(vm.spinner.View() + "Running git fetch..."))
		showTree = true
	} else if vm.runningGitLabAPIBranch >= 0 && vm.runningGitLabAPIBranch < len(vm.targetBranches) {
		sb.WriteString(colors.ProgressStyle.Render(vm.spinner.View() + "Querying GitLab API for " + vm.targetBranches[vm.runningGitLabAPIBranch].Short() + "..."))
		showTree = true
	} else if vm.runningCheckCommitHistory {
		sb.WriteString(colors.ProgressStyle.Render(vm.spinner.View() + "Checking commit history for merge commits..."))
		showTree = true
	} else if vm.runningPropagateMergeCommit {
		sb.WriteString(colors.ProgressStyle.Render(vm.spinner.View() + "Checking if sub-stacks are merged already..."))
		showTree = true
	} else {
		sb.WriteString(colors.SuccessStyle.Render("✓ GitLab fetch is done"))
	}

	if showTree {
		sb.WriteString("\n")

		syncedBranches := map[plumbing.ReferenceName]bool{}
		pendingBranches := map[plumbing.ReferenceName]bool{}
		for i, br := range vm.targetBranches {
			if i > vm.runningGitLabAPIBranch {
				pendingBranches[br] = true
			} else if i < vm.runningGitLabAPIBranch {
				syncedBranches[br] = true
			}
		}
		var brs []string
		for _, br := range vm.targetBranches {
			brs = append(brs, br.Short())
		}
		var nodes []*stackutils.StackTreeNode
		var err error
		nodes, err = stackutils.BuildStackTreeRelatedBranchStacks(
			vm.db.ReadTx(),
			vm.currentBranch.Short(),
			true,
			brs,
		)
		if err != nil {
			sb.WriteString("Failed to build stack tree: " + err.Error())
		} else {
			sb.WriteString("\n")
			for _, node := range nodes {
				sb.WriteString(stackutils.RenderTree(node, func(branchName string, isTrunk bool) string {
					var suffix string
					avbr, _ := vm.db.ReadTx().Branch(branchName)
					if avbr.MergeCommit != "" {
						suffix = " (merged)"
					}
					bn := plumbing.NewBranchReferenceName(branchName)
					if syncedBranches[bn] {
						return colors.SuccessStyle.Render("✓ " + branchName + suffix)
					}
					if pendingBranches[bn] {
						return colors.ProgressStyle.Render(branchName + suffix)
					}
					if vm.runningGitLabAPIBranch > 0 && vm.runningGitLabAPIBranch < len(vm.targetBranches) && vm.targetBranches[vm.runningGitLabAPIBranch] == bn {
						return colors.ProgressStyle.Render(vm.spinner.View() + branchName + suffix)
					}
					return branchName
				}))
			}
		}
	}
	return sb.String()
}

func (vm *GitLabFetchModel) runGitFetch() tea.Msg {
	remote := vm.repo.GetRemoteName()
	if _, err := vm.repo.Git(context.Background(), "fetch", remote); err != nil {
		return errors.Errorf("failed to fetch from %s: %v", remote, err)
	}
	return &GitLabFetchProgress{gitFetchIsDone: true}
}

func (vm *GitLabFetchModel) runGitLabAPIFetch() tea.Msg {
	if len(vm.targetBranches) <= vm.runningGitLabAPIBranch {
		return &GitLabFetchProgress{apiFetchIsDone: true}
	}
	br := vm.targetBranches[vm.runningGitLabAPIBranch]
	tx := vm.db.WriteTx()
	defer tx.Abort()
	avbr, _ := tx.Branch(br.Short())
	if avbr.MergeCommit != "" {
		return &GitLabFetchProgress{apiFetchIsDone: true}
	}
	// TODO: Implement UpdateMergeRequestState function for GitLab
	// This is a placeholder that will be implemented in later steps
	// _, err := actions.UpdateMergeRequestState(context.Background(), vm.client, tx, br.Short())
	// if err != nil {
	//     return err
	// }
	if err := tx.Commit(); err != nil {
		return errors.Errorf("failed to commit: %v", err)
	}
	return &GitLabFetchProgress{apiFetchIsDone: true}
}

func (vm *GitLabFetchModel) updateMergeCommitsFromCommitMessage() tea.Msg {
	trunkRefs := map[plumbing.ReferenceName]bool{}
	for _, br := range vm.targetBranches {
		avbr, _ := vm.db.ReadTx().Branch(br.Short())
		if avbr.Parent.Trunk {
			trunkRefs[plumbing.NewBranchReferenceName(avbr.Parent.Name)] = true
		}
	}

	repo := vm.repo.GoGitRepo()
	remote, err := repo.Remote(vm.repo.GetRemoteName())
	if err != nil {
		return errors.Errorf("failed to get remote %s: %v", vm.repo.GetRemoteName(), err)
	}
	remoteConfig := remote.Config()

	// For each trunk commits, look first 10000 commits to find the recently merged
	// commit. This is a best-effort approach for GitLab merge request detection.
	mergedMRs := map[int64]plumbing.Hash{}
	for trunkRef := range trunkRefs {
		rtb := mapToRemoteTrackingBranch(remoteConfig, trunkRef)
		if rtb == nil {
			// No remote tracking branch. Skip.
			continue
		}
		ref, err := repo.Reference(*rtb, true)
		if err != nil {
			return errors.Errorf("failed to get reference %q: %v", rtb, err)
		}
		c, err := repo.CommitObject(ref.Hash())
		if err != nil {
			return errors.Errorf("failed to get commit %q: %v", ref.Hash(), err)
		}
		visited := 0
		_ = object.NewCommitPreorderIter(c, nil, nil).ForEach(func(c *object.Commit) error {
			// TODO: Implement GitLab merge request detection from commit messages
			// This would look for patterns like "Merge branch 'feature' into 'main'"
			// or GitLab-specific merge commit patterns
			visited += 1
			if visited >= 10000 {
				return errors.New("stop")
			}
			return nil
		})
	}
	for _, br := range vm.targetBranches {
		tx := vm.db.WriteTx()
		avbr, _ := tx.Branch(br.Short())
		if avbr.MergeCommit != "" {
			tx.Abort()
			continue
		}
		if avbr.PullRequest != nil && avbr.PullRequest.Number != 0 {
			if hash, ok := mergedMRs[avbr.PullRequest.Number]; ok {
				avbr.MergeCommit = hash.String()
				tx.SetBranch(avbr)
			}
		}
		if err := tx.Commit(); err != nil {
			return errors.Errorf("failed to commit: %v", err)
		}
	}
	return &GitLabFetchProgress{checkCommitHistoryIsDone: true}
}

func (vm *GitLabFetchModel) updateMergeCommitsFromChildren() tea.Msg {
	// If child branches are merged, the parent branch is also merged.
	for _, br := range vm.targetBranches {
		tx := vm.db.WriteTx()
		avbr, _ := tx.Branch(br.Short())
		if avbr.MergeCommit == "" {
			tx.Abort()
			continue
		}
		parent := avbr.Parent
		for !parent.Trunk {
			parentBr, ok := tx.Branch(parent.Name)
			if !ok {
				break
			}
			if parentBr.MergeCommit != "" {
				break
			}
			parentBr.MergeCommit = avbr.MergeCommit
			tx.SetBranch(parentBr)
			parent = parentBr.Parent
		}
		if err := tx.Commit(); err != nil {
			return errors.Errorf("failed to commit: %v", err)
		}
	}
	return &GitLabFetchProgress{mergeCommitPropagationIsDone: true}
}

func mapToRemoteTrackingBranch(
	remoteConfig *config.RemoteConfig,
	refName plumbing.ReferenceName,
) *plumbing.ReferenceName {
	for _, fetch := range remoteConfig.Fetch {
		if fetch.Match(refName) {
			dst := fetch.Dst(refName)
			return &dst
		}
	}
	return nil
}