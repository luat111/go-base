package tracing

import (
	"context"
	"net/http"

	"github.com/google/uuid"
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

func (m *CorrelationIDService) CorrelationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		headerName := m.getHeaderName()
		corrId := r.Header.Get(headerName)
		if corrId == "" && m.EnforceHeader {
			corrId = m.generateId()
		}

		updCtx := WithCorrelationId(r.Context(), corrId)
		next.ServeHTTP(rw, r.WithContext(updCtx))
	})
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
