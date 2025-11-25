package audit

import (
	"gorm.io/gorm"
)

// Package audit provides comprehensive audit logging functionality for tracking
// actions, actors, requests, and responses in your application.
//
// Features:
// - Generic interface support for flexible type handling
// - Repository pattern for data persistence
// - Configurable middleware for Gin framework
// - Async/sync logging options
// - Request/response body capture with size limits
// - Sensitive header filtering
// - Correlation ID tracking
// - Custom metadata support
//
// Example usage:
//
//	// Initialize repository and service
//	repo := audit.NewAuditRepository(db)
//	service := audit.NewAuditService(repo)
//
//	// Configure middleware
//	config := audit.DefaultConfig()
//	config.GetActorID = func(c *gin.Context) string {
//		userID, _ := c.Get("user_id")
//		return userID.(string)
//	}
//
//	// Apply middleware
//	router.Use(audit.Middleware(service, config))

// New creates a new audit service with the default repository
func New(db *gorm.DB) IAuditService {
	repo := NewAuditRepository(db)
	return NewAuditService(repo)
}
