package utils

import (
	"github.com/gofiber/fiber/v2"
	"math"
)

// PaginationParams menyimpan parameter pagination
type PaginationParams struct {
	Page  int
	Limit int
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
		Page:  page,
		Limit: limit,
		Offset: offset,
	}
}

// GetPaginatedResponse membungkus hasil data dengan informasi pagination
func GetPaginatedResponse(data interface{}, total int, page, limit int) fiber.Map {
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return fiber.Map{
		"data":        data,
		"page":        page,
		"limit":       limit,
		"total":       total,
		"total_pages": totalPages,
	}
}
