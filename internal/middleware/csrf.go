package middleware

import "github.com/gofiber/fiber/v2"

const (
	CSRFCookieName = "csrf_token"
	CSRFHeaderName = "X-CSRF-Token"
)

// SkipCSRFPaths adalah Routes yang dilewati CSRF
var SkipCSRFPaths = map[string]bool{
	"/api/v1/auth/login":           true,
	"/api/v1/auth/register":        true,
	"/api/v1/auth/forgot-password": true,
	"/api/v1/auth/reset-password":  true,
}

func CSRF() fiber.Handler {
	return func(c *fiber.Ctx) error {

		// Skip safe methods
		switch c.Method() {
		case fiber.MethodGet, fiber.MethodHead, fiber.MethodOptions:
			return c.Next()
		}

		// Skip route tertentu
		if SkipCSRFPaths[c.Path()] {
			return c.Next()
		}

		cookieToken := c.Cookies(CSRFCookieName)
		if cookieToken == "" {
			return fiber.ErrForbidden
		}

		// Ambil header token, kalau kosong gunakan cookie sebagai fallback
		headerToken := c.Get(CSRFHeaderName)
		if headerToken == "" {
			headerToken = cookieToken
		}

		if cookieToken != headerToken {
			return fiber.ErrForbidden
		}

		return c.Next()
	}
}
