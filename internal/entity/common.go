package entity

import "math"

type PaginationMeta struct {
	Page       int   `json:"page" example:"1"`
	Size       int   `json:"size" example:"10"`
	TotalItems int64 `json:"totalItems" example:"42"`
	TotalPages int   `json:"totalPages" example:"5"`
}

func NewPaginationMeta(page, size int, total int64) PaginationMeta {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}

	totalPages := 0
	if size > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(size)))
	}

	return PaginationMeta{
		Page:       page,
		Size:       size,
		TotalItems: total,
		TotalPages: totalPages,
	}
}

// RequestMeta carries transport-level information captured by handlers.
type RequestMeta struct {
	IP        string
	UserAgent string
}
