# Provider Package

The `provider` package implements provider detection and selection logic to choose between GitHub and GitLab for the av CLI tool. This package is part of Step 2.2 of the GitLab support implementation.

## Features

- **Automatic Provider Detection**: Detects provider type from git remote URLs
- **Manual Override**: Supports forcing a specific provider through configuration or command-line flags
- **Multi-Remote Support**: Checks all git remotes, not just origin, with preference ordering
- **Authentication Discovery**: Finds tokens from config, environment variables, and CLI tools
- **Provider Validation**: Validates authentication and connectivity
- **Enterprise Support**: Handles both SaaS and self-hosted instances

## Architecture

### Core Components

1. **provider.go**: Main provider detection logic
2. **interface.go**: Provider interface abstraction (for future extensibility)
3. **remotes.go**: Git remote parsing and analysis
4. **config.go**: Configuration override support
5. **flags.go**: Command-line flag support
6. **validation.go**: Comprehensive provider validation
7. **errors.go**: Provider-specific error definitions

### Provider Types

```go
type ProviderType string

const (
    GitHub  ProviderType = "github"
    GitLab  ProviderType = "gitlab"  
    Unknown ProviderType = "unknown"
)
```

### Provider Structure

```go
type Provider struct {
    Type         ProviderType // GitHub or GitLab
    BaseURL      string       // Base URL (e.g., https://github.com)
    RepoSlug     string       // Repository slug (owner/repo)
    IsEnterprise bool         // Enterprise/self-hosted instance
    Token        string       // Authentication token
}
```

## Usage Examples

### Basic Provider Detection

```go
ctx := context.Background()
repo, err := git.GetRepo()
if err != nil {
    return err
}

provider, err := provider.DetectProvider(ctx, repo)
if err != nil {
    return err
}

fmt.Printf("Detected %s provider at %s\n", provider.Type, provider.BaseURL)
```

### Provider Detection with Options

```go
options := provider.DetectionOptions{
    ForceProvider:         provider.GitLab,      // Force GitLab
    AllowFallback:        true,                  // Allow config fallback
    RequireAuthentication: true,                 // Require valid token
}

provider, err := provider.DetectProviderWithOptions(ctx, repo, options)
```

### Configuration Overrides

```go
overrides := &provider.ConfigOverrides{
    Provider:    provider.GitHub,
    GitHubToken: "ghp_custom_token",
    BaseURL:     "https://github.enterprise.com",
}

provider, err := provider.DetectProviderWithConfig(ctx, repo, overrides)
```

### Command-Line Flag Support

```go
var flags provider.ProviderFlags
provider.AddProviderFlags(cmd, &flags)

// Later, in the command execution:
overrides, err := provider.ConfigFromFlags(&flags)
if err != nil {
    return err
}

provider, err := provider.DetectProviderWithConfig(ctx, repo, overrides)
```

### Comprehensive Validation

```go
result := provider.ValidateProviderComprehensive(ctx, provider)
if !result.IsValid {
    for _, err := range result.Errors {
        fmt.Printf("Error: %v\n", err)
    }
}

fmt.Printf("Authentication: %s\n", result.AuthenticationStatus)
fmt.Printf("Connectivity: %s\n", result.ConnectivityStatus)
```

## Detection Logic

The provider detection follows this priority order:

1. **Manual Override**: Command-line flags or configuration overrides
2. **Git Remote Detection**: Analyze git remote URLs in preference order:
   - `origin` remote (highest priority)
   - `upstream` remote
   - Any other remotes with recognized providers
3. **Configuration Fallback**: Use configured provider if no remote detection

### URL Pattern Recognition

- **GitHub**: 
  - `github.com` or hostnames containing `github`
  - SSH: `git@github.com:owner/repo.git`
  - HTTPS: `https://github.com/owner/repo.git`

- **GitLab**:
  - `gitlab.com` or hostnames containing `gitlab`
  - SSH: `git@gitlab.com:namespace/project.git`
  - HTTPS: `https://gitlab.com/namespace/project.git`

## Authentication Discovery

Token discovery follows this precedence:

### GitHub
1. Configuration: `config.Av.GitHub.Token`
2. Environment: `AV_GITHUB_TOKEN` → `GITHUB_TOKEN`
3. GitHub CLI: `gh auth token`

### GitLab
1. Configuration: `config.Av.GitLab.Token`
2. Environment: `AV_GITLAB_TOKEN` → `GITLAB_TOKEN`
3. GitLab CLI: `glab auth print-access-token`

## Enterprise Support

The package automatically detects enterprise instances:

- **GitHub Enterprise**: Any hostname other than `github.com`
- **GitLab Self-hosted**: Any hostname other than `gitlab.com`

Enterprise instances require proper base URL configuration and may have different API endpoints or authentication requirements.

## Error Handling

The package provides specific error types for different failure scenarios:

```go
var (
    ErrUnsupportedProvider   = errors.New("unsupported provider")
    ErrProviderNotDetected   = errors.New("provider not detected")
    ErrNoAuthentication     = errors.New("no authentication found")
    ErrInvalidConfiguration = errors.New("invalid provider configuration")
    // ... more errors
)
```

## Integration Points

This package integrates with:

- **internal/config**: Provider configuration and mutual exclusivity validation
- **internal/git**: Git repository and remote analysis
- **internal/gitlab**: GitLab-specific utilities and URL parsing
- Future integration with GitHub/GitLab client packages

## Testing

The package includes comprehensive tests covering:

- URL pattern detection
- Provider type parsing
- Token format validation
- Configuration override logic
- Flag validation
- Error scenarios

Run tests with:
```bash
go test ./internal/provider/...
```

## Future Extensibility

The `ClientInterface` abstraction provides a foundation for future provider-agnostic operations, allowing the same code to work with different providers through a common interface.

## Implementation Status

✅ **Completed Features:**
- Provider detection from git remotes
- Configuration and flag override support
- Token discovery from multiple sources
- Comprehensive validation framework
- Enterprise instance support
- Multi-remote analysis with preference ordering

⏳ **Future Enhancements:**
- Full API connectivity validation
- Repository access permission checks
- ClientInterface implementations
- Rate limiting integration
- Enhanced CLI tool integration
