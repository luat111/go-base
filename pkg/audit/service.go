package audit

import (
	"context"
	"encoding/json"
	"fmt"
)

// IAuditService defines the interface for audit service operations
type IAuditService interface {
	// Log creates an audit log entry
	Log(ctx context.Context, entry *AuditEntry) error

	// GetByID retrieves an audit log by ID
	GetByID(ctx context.Context, id uint) (*AuditLog, error)

	// GetByActor retrieves audit logs for a specific actor
	GetByActor(ctx context.Context, actorID string, limit int) ([]*AuditLog, error)

	// GetByResource retrieves audit logs for a specific resource
	GetByResource(ctx context.Context, resource, resourceID string, limit int) ([]*AuditLog, error)

	// GetByCorrelationID retrieves audit logs by correlation ID
	GetByCorrelationID(ctx context.Context, correlationID string) ([]*AuditLog, error)

	// Query retrieves paginated audit logs with filters
	Query(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*AuditLog, int64, error)
}

// AuditEntry represents the data needed to create an audit log
type AuditEntry struct {
	ActorID       string
	ActorType     string
	Action        string
	Resource      string
	ResourceID    string
	Description   string
	Method        string
	Path          string
	Query         string
	RequestBody   any
	RequestHeader map[string]string
	StatusCode    int
	ResponseBody  any
	ResponseTime  int64
	IPAddress     string
	UserAgent     string
	CorrelationID string
	Metadata      map[string]interface{}
}

// AuditService implements IAuditService
type AuditService struct {
	repo IAuditRepository
}

// NewAuditService creates a new audit service instance
func NewAuditService(repo IAuditRepository) IAuditService {
	return &AuditService{repo: repo}
}

// Log creates an audit log entry
func (s *AuditService) Log(ctx context.Context, entry *AuditEntry) error {
	// Marshal request body
	requestBody, err := marshalToJSON(entry.RequestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Marshal response body
	responseBody, err := marshalToJSON(entry.ResponseBody)
	if err != nil {
		return fmt.Errorf("failed to marshal response body: %w", err)
	}

	// Marshal request headers
	requestHeader, err := json.Marshal(entry.RequestHeader)
	if err != nil {
		return fmt.Errorf("failed to marshal request headers: %w", err)
	}

	// Marshal metadata
	metadata, err := json.Marshal(entry.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	log := &AuditLog{
		ActorID:       entry.ActorID,
		ActorType:     entry.ActorType,
		Action:        entry.Action,
		Resource:      entry.Resource,
		ResourceID:    entry.ResourceID,
		Description:   entry.Description,
		Method:        entry.Method,
		Path:          entry.Path,
		Query:         entry.Query,
		RequestBody:   requestBody,
		RequestHeader: string(requestHeader),
		StatusCode:    entry.StatusCode,
		ResponseBody:  responseBody,
		ResponseTime:  entry.ResponseTime,
		IPAddress:     entry.IPAddress,
		UserAgent:     entry.UserAgent,
		CorrelationID: entry.CorrelationID,
		Metadata:      string(metadata),
	}

	return s.repo.Create(ctx, log)
}

// GetByID retrieves an audit log by ID
func (s *AuditService) GetByID(ctx context.Context, id uint) (*AuditLog, error) {
	return s.repo.FindByID(ctx, id)
}

// GetByActor retrieves audit logs for a specific actor
func (s *AuditService) GetByActor(ctx context.Context, actorID string, limit int) ([]*AuditLog, error) {
	return s.repo.FindByActor(ctx, actorID, limit)
}

// GetByResource retrieves audit logs for a specific resource
func (s *AuditService) GetByResource(ctx context.Context, resource, resourceID string, limit int) ([]*AuditLog, error) {
	return s.repo.FindByResource(ctx, resource, resourceID, limit)
}

// GetByCorrelationID retrieves audit logs by correlation ID
func (s *AuditService) GetByCorrelationID(ctx context.Context, correlationID string) ([]*AuditLog, error) {
	return s.repo.FindByCorrelationID(ctx, correlationID)
}

// Query retrieves paginated audit logs with filters
func (s *AuditService) Query(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*AuditLog, int64, error) {
	return s.repo.Paginate(ctx, page, pageSize, filters)
}

// marshalToJSON converts any value to JSON string, handling nil and empty cases
func marshalToJSON(v interface{}) (string, error) {
	if v == nil {
		return "", nil
	}

	// Check if it's already a string
	if str, ok := v.(string); ok {
		return str, nil
	}

	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
