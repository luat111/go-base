package audit

import "github.com/gin-gonic/gin"

// AuditConfig defines the configuration for the audit middleware
type AuditConfig struct {
	// Skip defines a function to skip audit logging for certain requests
	Skip func(*gin.Context) bool

	// GetActorID extracts the actor ID from the request context
	// Default: extracts from "user_id" context key
	GetActorID func(*gin.Context) string

	// GetActorType extracts the actor type from the request context
	// Default: returns "user"
	GetActorType func(*gin.Context) string

	// GetAction extracts the action from the request
	// Default: uses HTTP method
	GetAction func(*gin.Context) string

	// GetResource extracts the resource name from the request
	// Default: uses the first path segment
	GetResource func(*gin.Context) string

	// GetResourceID extracts the resource ID from the request
	// Default: extracts from path parameters or query
	GetResourceID func(*gin.Context) string

	// GetDescription generates a description for the audit log
	// Default: combines method and path
	GetDescription func(*gin.Context) string

	// IncludeRequestBody determines if request body should be logged
	// Default: true
	IncludeRequestBody bool

	// IncludeResponseBody determines if response body should be logged
	// Default: true
	IncludeResponseBody bool

	// IncludeRequestHeaders determines if request headers should be logged
	// Default: false
	IncludeRequestHeaders bool

	// SensitiveHeaders lists headers that should not be logged
	// Default: ["Authorization", "Cookie", "X-Api-Key"]
	SensitiveHeaders []string

	// AsyncLogging enables asynchronous audit log writing
	// Default: true
	AsyncLogging bool

	// GetMetadata extracts additional metadata from the request
	// Default: nil
	GetMetadata func(*gin.Context) map[string]interface{}
}

// DefaultConfig returns the default configuration
func DefaultConfig() AuditConfig {
	return AuditConfig{
		Skip: func(c *gin.Context) bool {
			// Skip health check and metrics endpoints by default
			path := c.Request.URL.Path
			return path == "/health" || path == "/metrics" || path == "/ping"
		},
		GetActorID: func(c *gin.Context) string {
			if userID, exists := c.Get("user_id"); exists {
				if id, ok := userID.(string); ok {
					return id
				}
			}
			return "anonymous"
		},
		GetActorType: func(c *gin.Context) string {
			if actorType, exists := c.Get("actor_type"); exists {
				if t, ok := actorType.(string); ok {
					return t
				}
			}
			return "user"
		},
		GetAction: func(c *gin.Context) string {
			return c.Request.Method
		},
		GetResource: func(c *gin.Context) string {
			// Extract first path segment as resource
			path := c.Request.URL.Path
			if len(path) > 1 {
				segments := splitPath(path)
				if len(segments) > 0 {
					return segments[0]
				}
			}
			return "unknown"
		},
		GetResourceID: func(c *gin.Context) string {
			// Try to get from common parameter names
			if id := c.Param("id"); id != "" {
				return id
			}
			if id := c.Query("id"); id != "" {
				return id
			}
			return ""
		},
		GetDescription: func(c *gin.Context) string {
			return c.Request.Method + " " + c.Request.URL.Path
		},
		IncludeRequestBody:    true,
		IncludeResponseBody:   true,
		IncludeRequestHeaders: false,
		SensitiveHeaders:      []string{"Authorization", "Cookie", "X-Api-Key", "X-Auth-Token"},
		AsyncLogging:          true,
		GetMetadata:           nil,
	}
}

// splitPath splits a URL path into segments
func splitPath(path string) []string {
	var segments []string
	current := ""

	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			if current != "" {
				segments = append(segments, current)
				current = ""
			}
		} else {
			current += string(path[i])
		}
	}

	if current != "" {
		segments = append(segments, current)
	}

	return segments
}
