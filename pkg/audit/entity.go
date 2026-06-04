package audit

import (
	entity "go-base/pkg/datasource/postgres/entities"
)

// AuditLog represents an audit trail entry for tracking actions
type AuditLog struct {
	entity.BaseEntity

	// Actor information
	ActorID   string `gorm:"index;size:255" json:"actor_id"`
	ActorType string `gorm:"index;size:100" json:"actor_type"` // e.g., "user", "system", "service"

	// Action details
	Action      string `gorm:"index;size:100" json:"action"`      // e.g., "create", "update", "delete", "read"
	Resource    string `gorm:"index;size:255" json:"resource"`    // e.g., "user", "order", "product"
	ResourceID  string `gorm:"index;size:255" json:"resource_id"` // ID of the affected resource
	Description string `gorm:"type:text" json:"description"`      // Human-readable description

	// Request details
	Method        string `gorm:"size:10" json:"method"`                     // HTTP method
	Path          string `gorm:"size:500" json:"path"`                      // Request path
	Query         string `gorm:"type:text" json:"query,omitempty"`          // Query parameters
	RequestBody   string `gorm:"type:text" json:"request_body,omitempty"`   // Request payload
	RequestHeader string `gorm:"type:text" json:"request_header,omitempty"` // Request headers (selective)

	// Response details
	StatusCode   int    `json:"status_code"`                              // HTTP status code
	ResponseBody string `gorm:"type:text" json:"response_body,omitempty"` // Response payload
	ResponseTime int64  `json:"response_time"`                            // Response time in milliseconds

	// Metadata
	IPAddress     string `gorm:"size:45" json:"ip_address"` // IPv4 or IPv6
	UserAgent     string `gorm:"size:500" json:"user_agent,omitempty"`
	CorrelationID string `gorm:"index;size:255" json:"correlation_id"` // For tracing
	Metadata      string `gorm:"type:jsonb" json:"metadata,omitempty"` // Additional custom data
}

// TableName specifies the table name for AuditLog
func (AuditLog) TableName() string {
	return "audit_logs"
}
