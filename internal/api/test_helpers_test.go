package api

import (
	"errors"
	"testing"

	"github.com/gofiber/fiber/v3"
)

// newTestApp creates a Fiber app with the standard test error handler.
func newTestApp(t *testing.T) *fiber.App {
	t.Helper()
	return fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})
}

// newTestAppWithUser creates a Fiber app with user context middleware.
// The middleware sets user_id and user_role into c.Locals().
func newTestAppWithUser(t *testing.T, userID, role string) *fiber.App {
	t.Helper()
	app := newTestApp(t)
	app.Use(func(c fiber.Ctx) error {
		c.Locals("user_id", userID)
		c.Locals("user_role", role)
		return c.Next()
	})
	return app
}
