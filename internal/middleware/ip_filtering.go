package middleware

import (
	"net"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type IPFilterConfig struct {
	Whitelist []string
	Blacklist []string
}

func IPFilterMiddleware(cfg *IPFilterConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()

		// cek blacklist dulu
		for _, b := range cfg.Blacklist {
			if matchIP(ip, b) {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "access denied",
				})
			}
		}

		// jika whitelist ada, hanya izinkan IP di whitelist
		if len(cfg.Whitelist) > 0 {
			allowed := false
			for _, w := range cfg.Whitelist {
				if matchIP(ip, w) {
					allowed = true
					break
				}
			}
			if !allowed {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "access denied",
				})
			}
		}

		return c.Next()
	}
}

// matchIP bisa mendukung IP tunggal atau subnet
func matchIP(ip, rule string) bool {
	if strings.Contains(rule, "/") {
		_, cidr, err := net.ParseCIDR(rule)
		if err != nil {
			return false
		}
		return cidr.Contains(net.ParseIP(ip))
	}
	return ip == rule
}
