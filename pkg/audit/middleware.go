package audit

import (
	"bytes"
	"go-base/pkg/tracing"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Middleware creates a Gin middleware for audit logging
func Middleware(service IAuditService, config AuditConfig) gin.HandlerFunc {
	// Merge with default config
	config = mergeWithDefaults(config)

	return func(c *gin.Context) {
		// Skip if configured
		if config.Skip != nil && config.Skip(c) {
			c.Next()
			return
		}

		start := time.Now()

		// Capture request body
		var requestBody string
		if config.IncludeRequestBody && c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				// Restore the body for downstream handlers
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				requestBody = string(bodyBytes)
			}
		}

		// Capture request headers (filtered)
		var requestHeaders map[string]string
		if config.IncludeRequestHeaders {
			requestHeaders = make(map[string]string)
			for key, values := range c.Request.Header {
				// Skip sensitive headers
				if !isSensitiveHeader(key, config.SensitiveHeaders) {
					requestHeaders[key] = strings.Join(values, ", ")
				}
			}
		}

		// Capture response using custom writer
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
			captureBody:    config.IncludeResponseBody,
		}
		c.Writer = writer

		// Process request
		c.Next()

		// Calculate response time
		responseTime := time.Since(start).Milliseconds()

		// Get response body
		var responseBody string
		if config.IncludeResponseBody && writer.body.Len() > 0 {
			responseBody = writer.body.String()
		}

		// Extract IP address
		ipAddress := getIPAddress(c.Request)

		// Get correlation ID
		correlationID := tracing.FromContext(c.Request.Context())

		// Extract metadata
		var metadata map[string]interface{}
		if config.GetMetadata != nil {
			metadata = config.GetMetadata(c)
		}

		// Create audit entry
		entry := &AuditEntry{
			ActorID:       config.GetActorID(c),
			ActorType:     config.GetActorType(c),
			Action:        config.GetAction(c),
			Resource:      config.GetResource(c),
			ResourceID:    config.GetResourceID(c),
			Description:   config.GetDescription(c),
			Method:        c.Request.Method,
			Path:          c.Request.URL.Path,
			Query:         c.Request.URL.RawQuery,
			RequestBody:   requestBody,
			RequestHeader: requestHeaders,
			StatusCode:    writer.Status(),
			ResponseBody:  responseBody,
			ResponseTime:  responseTime,
			IPAddress:     ipAddress,
			UserAgent:     c.Request.UserAgent(),
			CorrelationID: correlationID,
			Metadata:      metadata,
		}

		// Log audit entry
		if config.AsyncLogging {
			// Asynchronous logging (non-blocking)
			go func() {
				_ = service.Log(c.Request.Context(), entry)
			}()
		} else {
			// Synchronous logging
			_ = service.Log(c.Request.Context(), entry)
		}
	}
}

// mergeWithDefaults merges the provided config with default values
func mergeWithDefaults(config AuditConfig) AuditConfig {
	defaults := DefaultConfig()

	if config.GetActorID == nil {
		config.GetActorID = defaults.GetActorID
	}
	if config.GetActorType == nil {
		config.GetActorType = defaults.GetActorType
	}
	if config.GetAction == nil {
		config.GetAction = defaults.GetAction
	}
	if config.GetResource == nil {
		config.GetResource = defaults.GetResource
	}
	if config.GetResourceID == nil {
		config.GetResourceID = defaults.GetResourceID
	}
	if config.GetDescription == nil {
		config.GetDescription = defaults.GetDescription
	}
	if config.SensitiveHeaders == nil {
		config.SensitiveHeaders = defaults.SensitiveHeaders
	}

	return config
}

// responseWriter wraps gin.ResponseWriter to capture response body
type responseWriter struct {
	gin.ResponseWriter
	body        *bytes.Buffer
	captureBody bool
}

func (w *responseWriter) Write(data []byte) (int, error) {
	// Capture body if enabled
	if w.captureBody {
		w.body.Write(data)
	}

	// Write to actual response
	return w.ResponseWriter.Write(data)
}

func (w *responseWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

// isSensitiveHeader checks if a header is in the sensitive list
func isSensitiveHeader(header string, sensitiveHeaders []string) bool {
	headerLower := strings.ToLower(header)
	for _, sensitive := range sensitiveHeaders {
		if strings.ToLower(sensitive) == headerLower {
			return true
		}
	}
	return false
}

// getIPAddress extracts the IP address from the request
func getIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header
	ips := strings.Split(r.Header.Get("X-Forwarded-For"), ",")
	if len(ips) > 0 && ips[0] != "" {
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}

	// Fall back to RemoteAddr
	return strings.TrimSpace(r.RemoteAddr)
}
