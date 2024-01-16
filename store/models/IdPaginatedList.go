package models

import (
	"encoding/json"

	"github.com/fazrithe/siasn-jf-backend-git/libs/search"
)

// IdPaginatedList is an alias of search.PaginatedList, but can be encoded into JSON with Bahasa
// Indonesia fields.
type IdPaginatedList[T any] search.PaginatedList[T]

type rawIdPaginatedList[T any] struct {
	Data     []T                      `json:"data"`
	Metadata *IdPaginatedListMetadata `json:"metadata"`
}

func (i IdPaginatedList[T]) MarshalJSON() ([]byte, error) {
	r := &rawIdPaginatedList[T]{Data: i.Data, Metadata: (*IdPaginatedListMetadata)(i.Metadata)}
	return json.Marshal(r)
}

// IdPaginatedListMetadata is an alias of search.PaginatedListMetadata, but can be encoded into JSON with Bahasa
// Indonesia fields.
type IdPaginatedListMetadata search.PaginatedListMetadata

type rawIdPaginatedListMetadata struct {
	// The length of the result.
	Subtotal int `json:"subtotal"`

	// The total entries there are that would fit in the result if it was not paginated.
	Total int `json:"total"`

	// The TotalPages count how many pages there is. Can be -1 if not calculated.
	TotalPages int `json:"total_halaman"`

	// PageNumber indicates the page number that you are on. It starts with 1.
	PageNumber int `json:"halaman"`

	// False to indicate that there is no next page.
	HasNext bool `json:"halaman_berikutnya"`
}

func (i IdPaginatedListMetadata) MarshalJSON() ([]byte, error) {
	m := rawIdPaginatedListMetadata(i)
	return json.Marshal(m)
}
