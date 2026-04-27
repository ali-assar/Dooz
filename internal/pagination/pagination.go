package pagination

import (
	"encoding/base64"
	"fmt"

	"dooz/internal/constants"

	"github.com/google/uuid"
)

type Direction string

const (
	DirectionNext Direction = "next"
	DirectionPrev Direction = "prev"
)

type SortOrder string

const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
)

type Cursor struct {
	ID uuid.UUID `json:"id"`
}

type Request struct {
	Limit     uint32                 `json:"limit" form:"limit"`
	Cursor    string                 `json:"cursor,omitempty" form:"cursor"`
	Direction Direction              `json:"direction,omitempty" form:"direction"`
	SortOrder SortOrder              `json:"sort_order,omitempty" form:"sort_order"`
	Filters   map[string]interface{} `json:"filters,omitempty" form:"-"`
}

type Response[T any] struct {
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	Limit      uint32 `json:"limit"`
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
	HasNext    bool   `json:"has_next"`
	HasPrev    bool   `json:"has_prev"`
}

func (c *Cursor) Encode() (string, error) {
	if c.ID == uuid.Nil {
		return "", nil
	}
	data := c.ID[:]
	return base64.RawURLEncoding.EncodeToString(data), nil
}

func DecodeCursor(encoded string) (*Cursor, error) {
	if encoded == "" {
		return nil, nil
	}
	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor format: %w", err)
	}
	if len(data) != 16 {
		return nil, fmt.Errorf("invalid cursor length: expected 16 bytes")
	}
	var id uuid.UUID
	copy(id[:], data)
	return &Cursor{ID: id}, nil
}

func (r *Request) Validate() error {
	if r.Limit <= 0 {
		r.Limit = constants.DefaultPaginationLimit
	}
	if r.Limit > constants.MaxPaginationLimit {
		return fmt.Errorf("limit cannot exceed %d", constants.MaxPaginationLimit)
	}
	if r.Direction == "" {
		r.Direction = DirectionNext
	}
	if r.Direction != DirectionNext && r.Direction != DirectionPrev {
		return fmt.Errorf("invalid direction: must be 'next' or 'prev'")
	}
	if r.SortOrder == "" {
		r.SortOrder = SortOrderDesc
	}
	if r.SortOrder != SortOrderAsc && r.SortOrder != SortOrderDesc {
		return fmt.Errorf("invalid sort_order: must be 'asc' or 'desc'")
	}
	return nil
}

func (r *Request) GetCursor() (*Cursor, error) {
	return DecodeCursor(r.Cursor)
}

func (r *Request) ShouldSortDesc() bool {
	if r.Direction == DirectionNext {
		return r.SortOrder == SortOrderDesc
	}
	return r.SortOrder == SortOrderAsc
}

func BuildResponse[T any](
	items []T,
	req *Request,
	getCursor func(T) uuid.UUID,
) (*Response[T], error) {
	if len(items) == 0 {
		return &Response[T]{
			Data: []T{},
			Pagination: Pagination{
				Limit:   req.Limit,
				HasNext: false,
				HasPrev: req.Cursor != "",
			},
		}, nil
	}

	if req.Direction == DirectionPrev {
		reverseSlice(items)
	}

	hasMore := uint32(len(items)) > req.Limit
	if hasMore {
		items = items[:req.Limit]
	}

	pag := Pagination{Limit: req.Limit}

	if req.Direction == DirectionNext {
		pag.HasNext = hasMore
		pag.HasPrev = req.Cursor != ""
	} else {
		pag.HasNext = req.Cursor != ""
		pag.HasPrev = hasMore
	}

	if pag.HasNext && len(items) > 0 {
		lastID := getCursor(items[len(items)-1])
		cursor := &Cursor{ID: lastID}
		encoded, err := cursor.Encode()
		if err != nil {
			return nil, fmt.Errorf("failed to encode next cursor: %w", err)
		}
		pag.NextCursor = encoded
	}

	if pag.HasPrev && len(items) > 0 {
		firstID := getCursor(items[0])
		cursor := &Cursor{ID: firstID}
		encoded, err := cursor.Encode()
		if err != nil {
			return nil, fmt.Errorf("failed to encode prev cursor: %w", err)
		}
		pag.PrevCursor = encoded
	}

	return &Response[T]{Data: items, Pagination: pag}, nil
}

func reverseSlice[T any](s []T) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

type QueryParams struct {
	CursorID  uuid.UUID
	HasCursor bool
	SortDesc  bool
	Limit     uint32
}

func (r *Request) GetQueryParams() (*QueryParams, error) {
	cursor, err := r.GetCursor()
	if err != nil {
		return nil, err
	}
	params := &QueryParams{
		SortDesc: r.ShouldSortDesc(),
		Limit:    r.Limit + 1,
	}
	if cursor != nil {
		params.CursorID = cursor.ID
		params.HasCursor = true
	}
	return params, nil
}

func (r *Request) SetFilter(key string, value interface{}) {
	if r.Filters == nil {
		r.Filters = make(map[string]interface{})
	}
	r.Filters[key] = value
}
