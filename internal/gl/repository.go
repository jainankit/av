package gl

import (
	"context"
	"fmt"
	"strings"
	"time"

	"emperror.dev/errors"
)

type Visibility string

const (
	VisibilityPrivate  Visibility = "private"
	VisibilityInternal Visibility = "internal"
	VisibilityPublic   Visibility = "public"
)

type Project struct {
	ID                                        int          `json:"id"`
	Name                                      string       `json:"name"`
	Path                                      string       `json:"path"`
	Description                               string       `json:"description"`
	DefaultBranch                             string       `json:"default_branch"`
	Visibility                                Visibility   `json:"visibility"`
	VisibilityLevel                           int          `json:"visibility_level"`
	Namespace                                 *Namespace   `json:"namespace"`
	Owner                                     *User        `json:"owner"`
	SSHURLToRepo                              string       `json:"ssh_url_to_repo"`
	HTTPURLToRepo                             string       `json:"http_url_to_repo"`
	WebURL                                    string       `json:"web_url"`
	ReadmeURL                                 string       `json:"readme_url"`
	TagList                                   []string     `json:"tag_list"`
	Topics                                    []string     `json:"topics"`
	NameWithNamespace                         string       `json:"name_with_namespace"`
	PathWithNamespace                         string       `json:"path_with_namespace"`
	IssuesEnabled                             bool         `json:"issues_enabled"`
	MergeRequestsEnabled                      bool         `json:"merge_requests_enabled"`
	WikiEnabled                               bool         `json:"wiki_enabled"`
	JobsEnabled                               bool         `json:"jobs_enabled"`
	SnippetsEnabled                           bool         `json:"snippets_enabled"`
	ContainerRegistryEnabled                  bool         `json:"container_registry_enabled"`
	ServiceDeskEnabled                        bool         `json:"service_desk_enabled"`
	CanCreateMergeRequestIn                   bool         `json:"can_create_merge_request_in"`
	IssuesAccessLevel                         string       `json:"issues_access_level"`
	RepositoryAccessLevel                     string       `json:"repository_access_level"`
	MergeRequestsAccessLevel                  string       `json:"merge_requests_access_level"`
	ForkingAccessLevel                        string       `json:"forking_access_level"`
	WikiAccessLevel                           string       `json:"wiki_access_level"`
	BuildsAccessLevel                         string       `json:"builds_access_level"`
	SnippetsAccessLevel                       string       `json:"snippets_access_level"`
	PagesAccessLevel                          string       `json:"pages_access_level"`
	OperationsAccessLevel                     string       `json:"operations_access_level"`
	AnalyticsAccessLevel                      string       `json:"analytics_access_level"`
	ContainerRegistryAccessLevel              string       `json:"container_registry_access_level"`
	SecurityAndComplianceAccessLevel          string       `json:"security_and_compliance_access_level"`
	ReleasesAccessLevel                       string       `json:"releases_access_level"`
	EnvironmentsAccessLevel                   string       `json:"environments_access_level"`
	FeatureFlagsAccessLevel                   string       `json:"feature_flags_access_level"`
	InfrastructureAccessLevel                 string       `json:"infrastructure_access_level"`
	MonitorAccessLevel                        string       `json:"monitor_access_level"`
	ModelExperimentsAccessLevel               string       `json:"model_experiments_access_level"`
	ModelRegistryAccessLevel                  string       `json:"model_registry_access_level"`
	CreatedAt                                 time.Time    `json:"created_at"`
	LastActivityAt                            time.Time    `json:"last_activity_at"`
	CreatorID                                 int          `json:"creator_id"`
	ImportURL                                 string       `json:"import_url"`
	ImportType                                string       `json:"import_type"`
	ImportStatus                              string       `json:"import_status"`
	ImportError                               string       `json:"import_error"`
	OpenIssuesCount                           int          `json:"open_issues_count"`
	RunnersToken                              string       `json:"runners_token"`
	CIDefaultGitDepth                         int          `json:"ci_default_git_depth"`
	CIForwardDeploymentEnabled                bool         `json:"ci_forward_deployment_enabled"`
	CIJobTokenScopeEnabled                    bool         `json:"ci_job_token_scope_enabled"`
	CISeparatedCaches                         bool         `json:"ci_separated_caches"`
	CIAllowForkPipelinesToRunInParentProject  bool         `json:"ci_allow_fork_pipelines_to_run_in_parent_project"`
	PublicJobs                                bool         `json:"public_jobs"`
	BuildTimeout                              int          `json:"build_timeout"`
	AutoCancelPendingPipelines                string       `json:"auto_cancel_pending_pipelines"`
	CICancelRedundantPipelines                bool         `json:"ci_cancel_redundant_pipelines"`
	BuildCoverageRegex                        string       `json:"build_coverage_regex"`
	CIConfigPath                              string       `json:"ci_config_path"`
	SharedWithGroups                          []SharedGroup `json:"shared_with_groups"`
	OnlyAllowMergeIfPipelineSucceeds          bool         `json:"only_allow_merge_if_pipeline_succeeds"`
	AllowMergeOnSkippedPipeline               bool         `json:"allow_merge_on_skipped_pipeline"`
	RestrictUserDefinedVariables              bool         `json:"restrict_user_defined_variables"`
	OnlyAllowMergeIfAllDiscussionsAreResolved bool         `json:"only_allow_merge_if_all_discussions_are_resolved"`
	RemoveSourceBranchAfterMerge              bool         `json:"remove_source_branch_after_merge"`
	PrintingMergeRequestLinkEnabled           bool         `json:"printing_merge_request_link_enabled"`
	MergeMethod                               string       `json:"merge_method"`
	SquashOption                              string       `json:"squash_option"`
	EnforceAuthChecksOnUploads                bool         `json:"enforce_auth_checks_on_uploads"`
	SuggestionCommitMessage                   string       `json:"suggestion_commit_message"`
	MergeCommitTemplate                       string       `json:"merge_commit_template"`
	SquashCommitTemplate                      string       `json:"squash_commit_template"`
	IssuesBranchTemplate                      string       `json:"issue_branch_template"`
	AutoDevopsEnabled                         bool         `json:"auto_devops_enabled"`
	AutoDevopsDeployStrategy                  string       `json:"auto_devops_deploy_strategy"`
	AutocloseReferencedIssues                 bool         `json:"autoclose_referenced_issues"`
	KeepLatestArtifact                        bool         `json:"keep_latest_artifact"`
	EmailsEnabled                             bool         `json:"emails_enabled"`
	ExternalAuthorizationClassificationLabel  string       `json:"external_authorization_classification_label"`
	RequirementsEnabled                       bool         `json:"requirements_enabled"`
	RequirementsAccessLevel                   string       `json:"requirements_access_level"`
	SecurityAndComplianceEnabled              bool         `json:"security_and_compliance_enabled"`
	ComplianceFrameworks                      []string     `json:"compliance_frameworks"`
	IssuesTemplate                            string       `json:"issues_template"`
	MergeRequestsTemplate                     string       `json:"merge_requests_template"`
	MergePipelinesEnabled                     bool         `json:"merge_pipelines_enabled"`
	MergeTrainsEnabled                        bool         `json:"merge_trains_enabled"`
	RestrictedVisibilityLevels                []int        `json:"restricted_visibility_levels"`
	Mirror                                    bool         `json:"mirror"`
	MirrorUserID                              int          `json:"mirror_user_id"`
	MirrorTriggerBuilds                       bool         `json:"mirror_trigger_builds"`
	OnlyMirrorProtectedBranches               bool         `json:"only_mirror_protected_branches"`
	MirrorOverwritesDivergedBranches          bool         `json:"mirror_overwrites_diverged_branches"`
	PackagesEnabled                           bool         `json:"packages_enabled"`
	ServiceDeskAddress                        string       `json:"service_desk_address"`
	IssuesCountLimit                          int          `json:"issues_count_limit"`
	RepositoryObjectFormat                    string       `json:"repository_object_format"`
	RepositorySize                            int64        `json:"repository_size"`
	LfsObjectsSize                            int64        `json:"lfs_objects_size"`
	JobArtifactsSize                          int64        `json:"job_artifacts_size"`
	PipelineArtifactsSize                     int64        `json:"pipeline_artifacts_size"`
	PackagesSize                              int64        `json:"packages_size"`
	SnippetsSize                              int64        `json:"snippets_size"`
	WikiSize                                  int64        `json:"wiki_size"`
	UploadsSize                               int64        `json:"uploads_size"`
	Statistics                                *ProjectStatistics `json:"statistics"`
}

type Namespace struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Kind     string `json:"kind"`
	FullPath string `json:"full_path"`
	ParentID *int   `json:"parent_id"`
	WebURL   string `json:"web_url"`
}

type SharedGroup struct {
	GroupID          int    `json:"group_id"`
	GroupName        string `json:"group_name"`
	GroupFullPath    string `json:"group_full_path"`
	GroupAccessLevel int    `json:"group_access_level"`
}

type ProjectStatistics struct {
	CommitCount           int64 `json:"commit_count"`
	StorageSize           int64 `json:"storage_size"`
	RepositorySize        int64 `json:"repository_size"`
	WikiSize              int64 `json:"wiki_size"`
	LfsObjectsSize        int64 `json:"lfs_objects_size"`
	JobArtifactsSize      int64 `json:"job_artifacts_size"`
	PipelineArtifactsSize int64 `json:"pipeline_artifacts_size"`
	PackagesSize          int64 `json:"packages_size"`
	SnippetsSize          int64 `json:"snippets_size"`
	UploadsSize           int64 `json:"uploads_size"`
}

// Repository is an alias for Project to maintain consistency with GitHub client
type Repository = Project

func (c *Client) GetProjectByID(ctx context.Context, projectID int) (*Project, error) {
	path := fmt.Sprintf("/projects/%d", projectID)
	
	var project Project
	if err := c.get(ctx, path, nil, &project); err != nil {
		return nil, errors.Wrapf(err, "unable to fetch project %d from GitLab", projectID)
	}
	
	return &project, nil
}

func (c *Client) GetProjectBySlug(ctx context.Context, slug string) (*Project, error) {
	// GitLab expects URL-encoded path for project slugs with special characters
	encodedSlug := strings.ReplaceAll(slug, "/", "%2F")
	path := fmt.Sprintf("/projects/%s", encodedSlug)
	
	var project Project
	if err := c.get(ctx, path, nil, &project); err != nil {
		return nil, errors.Wrapf(err, "unable to fetch project %q from GitLab", slug)
	}
	
	return &project, nil
}

func (c *Client) GetRepositoryBySlug(ctx context.Context, slug string) (*Repository, error) {
	return c.GetProjectBySlug(ctx, slug)
}