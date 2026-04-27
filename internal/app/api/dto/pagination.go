package dto

import "dooz/internal/pagination"

type CursorPaginationRequest struct {
	Limit     uint32 `form:"limit"`
	Cursor    string `form:"cursor"`
	Direction string `form:"direction"`
	SortOrder string `form:"sort_order"`
}

func (r *CursorPaginationRequest) ToPaginationRequest() *pagination.Request {
	req := &pagination.Request{
		Limit:     r.Limit,
		Cursor:    r.Cursor,
		Direction: pagination.Direction(r.Direction),
		SortOrder: pagination.SortOrder(r.SortOrder),
	}
	if err := req.Validate(); err != nil {
		req.Limit = 20
		req.Direction = pagination.DirectionNext
		req.SortOrder = pagination.SortOrderDesc
	}
	return req
}
