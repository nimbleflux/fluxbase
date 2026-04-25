package api

import "github.com/gofiber/fiber/v3"

// ParseBody binds the request body to req and returns a fiber.Error on failure.
func ParseBody(c fiber.Ctx, req any) error {
	if err := c.Bind().Body(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	return nil
}

// NormalizePaginationParams validates and normalizes limit/offset pagination parameters.
// It enforces the maximum limit and ensures offset is non-negative.
// Returns the normalized (limit, offset) values.
func NormalizePaginationParams(limit, offset, defaultLimit, maxLimit int) (int, int) {
	if limit <= 0 || limit > maxLimit {
		limit = defaultLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
