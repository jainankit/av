package providers

// PageInfo contains pagination information for cursor-based pagination
// This follows the GraphQL Relay specification for consistent pagination across providers
type PageInfo struct {
	// EndCursor points to the last item in the current page
	EndCursor string
	
	// HasNextPage indicates if there are more items available after the current page
	HasNextPage bool
	
	// HasPreviousPage indicates if there are items available before the current page
	HasPreviousPage bool
	
	// StartCursor points to the first item in the current page
	StartCursor string
}

// PaginationInput defines common pagination parameters for list operations
type PaginationInput struct {
	// First specifies the number of items to return from the beginning
	First *int32
	
	// After specifies the cursor to start returning items after
	After *string
	
	// Last specifies the number of items to return from the end (for backward pagination)
	Last *int32
	
	// Before specifies the cursor to start returning items before (for backward pagination)
	Before *string
}

// HasPagination returns true if any pagination parameters are set
func (p PaginationInput) HasPagination() bool {
	return p.First != nil || p.After != nil || p.Last != nil || p.Before != nil
}

// GetFirst returns the First value or a default if not set
func (p PaginationInput) GetFirst(defaultValue int32) int32 {
	if p.First != nil {
		return *p.First
	}
	return defaultValue
}

// GetAfter returns the After cursor or empty string if not set
func (p PaginationInput) GetAfter() string {
	if p.After != nil {
		return *p.After
	}
	return ""
}

// DefaultPageSize is the default number of items to return when no pagination is specified
const DefaultPageSize int32 = 50

// MaxPageSize is the maximum number of items that can be requested in a single page
const MaxPageSize int32 = 100