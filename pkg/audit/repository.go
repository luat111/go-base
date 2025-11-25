package audit

import (
	"context"
	"go-base/pkg/datasource/postgres/repository"

	"gorm.io/gorm"
)

// IAuditRepository defines the interface for audit log repository operations
type IAuditRepository interface {
	// Create creates a new audit log entry
	Create(ctx context.Context, log *AuditLog) error

	// FindByID retrieves an audit log by ID
	FindByID(ctx context.Context, id uint) (*AuditLog, error)

	// FindByActor retrieves audit logs for a specific actor
	FindByActor(ctx context.Context, actorID string, limit int) ([]*AuditLog, error)

	// FindByResource retrieves audit logs for a specific resource
	FindByResource(ctx context.Context, resource, resourceID string, limit int) ([]*AuditLog, error)

	// FindByCorrelationID retrieves audit logs by correlation ID
	FindByCorrelationID(ctx context.Context, correlationID string) ([]*AuditLog, error)

	// Paginate retrieves paginated audit logs with optional filters
	Paginate(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*AuditLog, int64, error)
}

// AuditRepository implements IAuditRepository by extending BaseRepository
type AuditRepository struct {
	*repository.BaseRepository
}

// NewAuditRepository creates a new audit repository instance
func NewAuditRepository(db *gorm.DB) IAuditRepository {
	return &AuditRepository{
		BaseRepository: repository.NewBaseRepository(db, &AuditLog{}),
	}
}

// Create creates a new audit log entry
func (r *AuditRepository) Create(ctx context.Context, log *AuditLog) error {
	return r.BaseRepository.Create(ctx, log)
}

// FindByID retrieves an audit log by ID
func (r *AuditRepository) FindByID(ctx context.Context, id uint) (*AuditLog, error) {
	var log AuditLog
	err := r.BaseRepository.DB.WithContext(ctx).First(&log, id).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// FindByActor retrieves audit logs for a specific actor
func (r *AuditRepository) FindByActor(ctx context.Context, actorID string, limit int) ([]*AuditLog, error) {
	var logs []*AuditLog
	err := r.BaseRepository.Find(ctx, &logs, func(db *gorm.DB) *gorm.DB {
		query := db.Where("actor_id = ?", actorID).Order("created_at DESC")
		if limit > 0 {
			query = query.Limit(limit)
		}
		return query
	})
	return logs, err
}

// FindByResource retrieves audit logs for a specific resource
func (r *AuditRepository) FindByResource(ctx context.Context, resource, resourceID string, limit int) ([]*AuditLog, error) {
	var logs []*AuditLog
	err := r.BaseRepository.Find(ctx, &logs, func(db *gorm.DB) *gorm.DB {
		query := db.Where("resource = ? AND resource_id = ?", resource, resourceID).Order("created_at DESC")
		if limit > 0 {
			query = query.Limit(limit)
		}
		return query
	})
	return logs, err
}

// FindByCorrelationID retrieves audit logs by correlation ID
func (r *AuditRepository) FindByCorrelationID(ctx context.Context, correlationID string) ([]*AuditLog, error) {
	var logs []*AuditLog
	err := r.BaseRepository.Find(ctx, &logs, func(db *gorm.DB) *gorm.DB {
		return db.Where("correlation_id = ?", correlationID).Order("created_at ASC")
	})
	return logs, err
}

// Paginate retrieves paginated audit logs with optional filters
func (r *AuditRepository) Paginate(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*AuditLog, int64, error) {
	var logs []*AuditLog
	total, err := r.BaseRepository.Paginate(ctx, page, pageSize, &logs, func(db *gorm.DB) *gorm.DB {
		// Apply filters
		for key, value := range filters {
			db = db.Where(key+" = ?", value)
		}
		return db
	})

	return logs, total, err
}
