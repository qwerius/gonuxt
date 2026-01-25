package utils

import (
	"github.com/gofiber/fiber/v2"
)

// PaginationParams menyimpan parameter pagination
type PaginationParams struct {
	Page   int
	Limit  int
	Offset int
}

// GetPagination mengambil query params "page" dan "limit" dari request Fiber
// defaultPage & defaultLimit digunakan jika query param tidak ada
// maxLimit membatasi agar user tidak bisa minta terlalu banyak data
func GetPagination(c *fiber.Ctx, defaultPage, defaultLimit, maxLimit int) PaginationParams {
	page := c.QueryInt("page", defaultPage)
	limit := c.QueryInt("limit", defaultLimit)

	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = defaultLimit
	} else if limit > maxLimit {
		limit = maxLimit
	}

	offset := (page - 1) * limit

	return PaginationParams{
		Page:   page,
		Limit:  limit,
		Offset: offset,
	}
}

// PaginationMeta struct untuk meta info
type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// GetPaginatedResponse membungkus hasil data dengan informasi pagination
// Mengembalikan: items dan meta terpisah agar tidak nested data.data
func GetPaginatedResponse(items interface{}, total, page, limit int) (data interface{}, meta PaginationMeta) {
	meta = PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: (total + limit - 1) / limit,
	}
	data = items
	return
}
