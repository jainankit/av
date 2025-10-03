package provider

import "errors"

// Common provider errors
var (
	// ErrUnsupportedProvider indicates the provider type is not supported
	ErrUnsupportedProvider = errors.New("unsupported provider")
	
	// ErrProviderNotDetected indicates no provider could be detected
	ErrProviderNotDetected = errors.New("provider not detected")
	
	// ErrNoAuthentication indicates no authentication token was found
	ErrNoAuthentication = errors.New("no authentication found")
	
	// ErrInvalidConfiguration indicates the provider configuration is invalid
	ErrInvalidConfiguration = errors.New("invalid provider configuration")
	
	// ErrNotImplemented indicates the feature is not yet implemented
	ErrNotImplemented = errors.New("feature not implemented")
	
	// ErrNoGitRepository indicates no git repository was found
	ErrNoGitRepository = errors.New("no git repository found")
	
	// ErrInvalidRepositoryURL indicates the repository URL is invalid
	ErrInvalidRepositoryURL = errors.New("invalid repository URL")
)
