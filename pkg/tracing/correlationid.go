package tracing

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type key string

const (
	DefaultHeaderName key = "X-Correlation-Id"
	ContextKey        key = "CorrelationId"
)

type CorrelationIDService struct {
	HeaderName    string
	EnforceHeader bool
	IdGenerator   func() string
}

func New() CorrelationIDService {
	return CorrelationIDService{
		HeaderName:    string(DefaultHeaderName),
		EnforceHeader: true,
		IdGenerator:   DefaultGenerator,
	}
}

// CorrelationMiddleware extracts or generates a correlation ID and stores it in
// the request context. When an active OTel span is present (injected by the
// otelgin middleware that runs before this one) its trace ID is used as the
// correlation ID so that logs and traces share the same identifier. If no span
// is active the value from the incoming header is used, or a new UUID is
// generated when EnforceHeader is true.
func (m *CorrelationIDService) CorrelationMiddleware(c *gin.Context) {
	headerName := m.getHeaderName()

	// corrId := c.Request.Header.Get(headerName)
	// Prefer the OTel trace ID – it is already set by otelgin which runs first.
	corrId := TraceIDFromContext(c.Request.Context())

	// Fall back to the incoming header value.
	if corrId == "" {
		corrId = c.Request.Header.Get(headerName)
	}

	// Last resort: generate a new ID.
	if corrId == "" && m.EnforceHeader {
		corrId = m.generateId()
	}

	// Expose the resolved ID as a response header so callers can correlate.
	c.Header(headerName, corrId)

	updCtx := WithCorrelationId(c.Request.Context(), corrId)
	c.Request = c.Request.WithContext(updCtx)

	c.Next()
}

// TraceIDFromContext returns the W3C-formatted trace ID of the active OTel span,
// or an empty string when no span is present.
func TraceIDFromContext(ctx context.Context) string {
	spanCtx := trace.SpanFromContext(ctx).SpanContext()
	if spanCtx.IsValid() {
		return spanCtx.TraceID().String()
	}
	return ""
}

func FromContext(ctx context.Context) string {
	corrId, ok := ctx.Value(ContextKey).(string)
	if ok {
		return corrId
	}
	return ""
}

func WithCorrelationId(ctx context.Context, correlationId string) context.Context {
	return context.WithValue(ctx, ContextKey, correlationId)
}

func DefaultGenerator() string {
	return uuid.NewString()
}

func AttachHeaderTracking(ctx context.Context, headers map[string]string) {
	id := FromContext(ctx)
	headers[string(DefaultHeaderName)] = id
}

func (m *CorrelationIDService) getHeaderName() string {
	if m.HeaderName == "" {
		return string(DefaultHeaderName)
	}

	return m.HeaderName
}

func (m *CorrelationIDService) generateId() string {
	if m.IdGenerator != nil {
		return m.IdGenerator()
	}

	return DefaultGenerator()
}
