package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	allowedHeaders = "Authorization, Content-Type, x-requested-with, origin, true-client-ip, X-Correlation-ID"
)

// CORS is a middleware that adds CORS (Cross-Origin Resource Sharing) headers to the response.
func CORS(middlewareConfigs map[string]string, routes *[]string) func(c *gin.Context) {
	return func(c *gin.Context) {
		setMiddlewareHeaders(middlewareConfigs, *routes, c.Writer)

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

func setMiddlewareHeaders(middlewareConfigs map[string]string, routes []string, w http.ResponseWriter) {
	routes = append(routes, "OPTIONS")

	// Set default headers
	defaultHeaders := map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": strings.Join(routes, ", "),
		"Access-Control-Allow-Headers": allowedHeaders,
	}

	// Add custom headers to the default headers
	for header, defaultValue := range defaultHeaders {
		if customValue, ok := middlewareConfigs[header]; ok && customValue != "" {
			if header == "Access-Control-Allow-Headers" {
				w.Header().Set(header, defaultValue+", "+customValue)
			} else {
				w.Header().Set(header, customValue)
			}
		} else {
			w.Header().Set(header, defaultValue)
		}
	}

	// Handle additional custom headers (not part of defaultHeaders)
	for header, customValue := range middlewareConfigs {
		if _, ok := defaultHeaders[header]; !ok {
			w.Header().Set(header, customValue)
		}
	}
}
