// Package search provides helper types and functions to facilitate searching processes, including
// listing a list of entries from the database.
//
// It contains the PaginatedList type and utilities, which provide a list that holds data and metadata in it.
package search

import "math"

// PaginatedList is a helper class to return the result of a query with all the metadata needed.
// See also PaginatedListMetadata and CreateQueryResultMetadata to create query result metadata
// which contains the length of the result.
type PaginatedList struct {
	// The result data, must be a slice.
	Data interface{} `json:"data"`

	// The metadata, which contains information like total pages.
	Metadata *PaginatedListMetadata `json:"metadata"`
}

type PaginatedListMetadata struct {
	// The length of the result.
	Subtotal int `json:"subtotal"`

	// The total entries there are that would fit in the result if it was not paginated.
	Total int `json:"total"`

	// The TotalPages count how many pages there is. Can be -1 if not calculated.
	TotalPages int `json:"total_pages"`

	// PageNumber indicates the page number that you are on. It starts with 1.
	PageNumber int `json:"page_number"`

	// HasNext indicates that there is still the next page.
	// If next page is unknown, set this to true.
	HasNext bool `json:"has_next"`
}

// CreatePaginatedListMetadata calculates number of pages from pageNumber and countPerPage, and not start and end. Users
// are more familiar with specifying pageNumber and countPerPage directly rather than start and end row numbers.
func CreatePaginatedListMetadata(pageNumber, countPerPage, resultLength, total int) *PaginatedListMetadata {
	totalPages := int(math.Ceil(float64(total) / float64(countPerPage)))
	hasNext := pageNumber < totalPages
	meta := &PaginatedListMetadata{
		Subtotal:   resultLength,
		Total:      total,
		TotalPages: totalPages,
		PageNumber: pageNumber,
		HasNext:    hasNext,
	}
	return meta
}

// CreatePaginatedListMetadataNoTotal works just like CreatePaginatedListMetadata, but with total entries unknown.
// This will cause TotalPages to be set to -1, due to, for some reasons, total pages
// cannot be calculated.
func CreatePaginatedListMetadataNoTotal(pageNumber, resultLength int) *PaginatedListMetadata {
	meta := &PaginatedListMetadata{
		Subtotal:   resultLength,
		Total:      -1,
		TotalPages: -1,
		PageNumber: pageNumber,
		HasNext:    true,
	}
	return meta
}

// CreatePaginatedListMetadataNoTotalNext works just like CreatePaginatedListMetadataNoTotal, but with next page known.
// This will cause TotalPages to be set to -1, due to, for some reasons, total pages
// cannot be calculated. However, hasNext can be set to false this way, saying that there is no next page.
func CreatePaginatedListMetadataNoTotalNext(pageNumber, resultLength int, hasNext bool) *PaginatedListMetadata {
	meta := &PaginatedListMetadata{
		Subtotal:   resultLength,
		Total:      -1,
		TotalPages: -1,
		PageNumber: pageNumber,
		HasNext:    hasNext,
	}
	return meta
}
