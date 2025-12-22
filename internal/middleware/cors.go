package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{})
	for _, o := range allowedOrigins {
		allowed[strings.TrimSpace(o)] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// ✅ Allow only configured origins
		if origin != "" {
			if _, ok := allowed[origin]; ok {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Vary", "Origin")
			}
		}

		// ✅ Always declare methods
		c.Header(
			"Access-Control-Allow-Methods",
			"GET, POST, PUT, PATCH, DELETE, OPTIONS",
		)

		// ✅ IMPORTANT: allow ALL headers browser may send
		// Especially custom headers like ngrok-skip-browser-warning
		reqHeaders := c.GetHeader("Access-Control-Request-Headers")
		if reqHeaders != "" {
			c.Header("Access-Control-Allow-Headers", reqHeaders)
		} else {
			c.Header(
				"Access-Control-Allow-Headers",
				"Authorization, Content-Type, Accept, ngrok-skip-browser-warning",
			)
		}

		// ✅ Handle preflight without aborting middleware chain incorrectly
		if c.Request.Method == http.MethodOptions {
			c.Status(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
