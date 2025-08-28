// Package provider defines abstractions for version control hosting providers.
//
// This package provides a common interface for interacting with different
// version control hosting platforms (GitHub, GitLab, etc.) in a provider-agnostic way.
// 
// The main components are:
//   - Provider interface: Core operations for merge requests, repositories, and users
//   - Common types: Provider-agnostic data structures for MRs, users, and repos
//   - Factory pattern: Creation and configuration of provider instances
//   - Auto-detection: Automatic provider detection from git remote URLs
//
// Example usage:
//
//   config := ProviderConfig{
//       Name:  ProviderGitHub,
//       Token: "ghp_...",
//   }
//   provider, err := NewProvider(ctx, config)
//   if err != nil {
//       return err
//   }
//   
//   mr, err := provider.GetMergeRequest(ctx, "pr-id")
//   if err != nil {
//       return err
//   }
//   fmt.Printf("MR: %s\n", mr.Title)
//
package provider