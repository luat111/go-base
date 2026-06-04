package middlewares

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// Tracing returns a Gin middleware that starts an OpenTelemetry span for every
// HTTP request and propagates W3C TraceContext headers both upstream and
// downstream. The span name follows the pattern "<HTTP method> <route>".
//
// serviceName should match the value used when calling tracing.Init().
func Tracing(serviceName string) gin.HandlerFunc {
	return otelgin.Middleware(serviceName)
}
