# Audit Middleware Package

The `pkg/audit` package provides comprehensive audit logging functionality for tracking actions, actors, requests, and responses in your Go application.

## Features

- ✅ **Generic Interface Support**: Flexible type handling for request/response bodies
- ✅ **Repository Pattern**: Clean separation of concerns with interface-based design
- ✅ **Gin Middleware**: Easy integration with Gin framework
- ✅ **Async/Sync Logging**: Choose between blocking and non-blocking audit logging
- ✅ **Request/Response Capture**: Automatic capture of request and response data
- ✅ **Size Limits**: Configurable limits to prevent excessive data storage
- ✅ **Sensitive Header Filtering**: Automatically filters sensitive headers
- ✅ **Correlation ID Tracking**: Built-in support for request tracing
- ✅ **Custom Metadata**: Extensible metadata support
- ✅ **Flexible Extractors**: Customizable functions for extracting audit data

## Installation

The package is already part of your `go-base` project under `pkg/audit`.

## Quick Start

### 1. Database Migration

First, ensure the `audit_logs` table is created by running migrations:

```go
import "go-base/pkg/audit"

// In your app initialization
app.DB().MigrateEntities([]any{&audit.AuditLog{}})
```

### 2. Basic Usage

```go
import (
    "go-base/pkg/audit"
    "github.com/gin-gonic/gin"
)

func main() {
    router := gin.Default()
    
    // Initialize audit service (request/response bodies can be any type)
    auditService := audit.New(db)
    
    // Use default configuration
    router.Use(audit.Middleware(auditService, audit.DefaultConfig()))
    
    // Your routes here
    router.GET("/users/:id", getUserHandler)
    router.POST("/users", createUserHandler)
    
    router.Run(":8080")
}
```

### 3. Custom Configuration

```go
// Create custom configuration
config := audit.DefaultConfig()

// Customize actor extraction
config.GetActorID = func(c *gin.Context) string {
    if userID, exists := c.Get("user_id"); exists {
        return userID.(string)
    }
    return "anonymous"
}

config.GetActorType = func(c *gin.Context) string {
    if isAdmin, _ := c.Get("is_admin"); isAdmin == true {
        return "admin"
    }
    return "user"
}

// Customize resource extraction
config.GetResource = func(c *gin.Context) string {
    // Extract from path: /api/v1/users/123 -> "users"
    segments := strings.Split(c.Request.URL.Path, "/")
    if len(segments) >= 4 {
        return segments[3]
    }
    return "unknown"
}

config.GetResourceID = func(c *gin.Context) string {
    return c.Param("id")
}

// Add custom metadata
config.GetMetadata = func(c *gin.Context) map[string]interface{} {
    return map[string]interface{}{
        "tenant_id": c.GetHeader("X-Tenant-ID"),
        "api_version": "v1",
    }
}

// Configure body capture
config.IncludeRequestBody = true
config.IncludeResponseBody = true
config.MaxBodySize = 20480 // 20KB

// Use synchronous logging for critical operations
config.AsyncLogging = false

// Apply middleware
router.Use(audit.Middleware(auditService, config))
```

### 4. Selective Auditing

Apply audit middleware only to specific routes:

```go
// Audit only API routes
apiGroup := router.Group("/api/v1")
apiGroup.Use(audit.Middleware(auditService, config))
{
    apiGroup.GET("/users/:id", getUserHandler)
    apiGroup.POST("/users", createUserHandler)
    apiGroup.PUT("/users/:id", updateUserHandler)
    apiGroup.DELETE("/users/:id", deleteUserHandler)
}

// Public routes without audit
router.GET("/health", healthCheckHandler)
router.GET("/ping", pingHandler)
```

### 5. Skip Certain Requests

```go
config := audit.DefaultConfig()

// Skip health checks and static files
config.Skip = func(c *gin.Context) bool {
    path := c.Request.URL.Path
    return path == "/health" || 
           path == "/metrics" || 
           strings.HasPrefix(path, "/static/")
}
```

## Querying Audit Logs

### Get Logs by Actor

```go
logs, err := auditService.GetByActor(ctx, "user-123", 50)
if err != nil {
    // handle error
}

for _, log := range logs {
    fmt.Printf("Action: %s, Resource: %s, Time: %s\n", 
        log.Action, log.Resource, log.CreatedAt)
}
```

### Get Logs by Resource

```go
logs, err := auditService.GetByResource(ctx, "users", "user-123", 50)
if err != nil {
    // handle error
}
```

### Get Logs by Correlation ID

```go
// Get all logs for a specific request trace
logs, err := auditService.GetByCorrelationID(ctx, "correlation-id-123")
if err != nil {
    // handle error
}
```

### Paginated Query with Filters

```go
filters := map[string]interface{}{
    "actor_id": "user-123",
    "action": "DELETE",
}

logs, total, err := auditService.Query(ctx, 1, 20, filters)
if err != nil {
    // handle error
}

fmt.Printf("Found %d total logs, showing page 1\n", total)
```

## Database Schema

The `AuditLog` entity creates the following table structure:

```sql
CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    
    -- Actor information
    actor_id VARCHAR(255),
    actor_type VARCHAR(100),
    
    -- Action details
    action VARCHAR(100),
    resource VARCHAR(255),
    resource_id VARCHAR(255),
    description TEXT,
    
    -- Request details
    method VARCHAR(10),
    path VARCHAR(500),
    query TEXT,
    request_body TEXT,
    request_header TEXT,
    
    -- Response details
    status_code INTEGER,
    response_body TEXT,
    response_time BIGINT,
    
    -- Metadata
    ip_address VARCHAR(45),
    user_agent VARCHAR(500),
    correlation_id VARCHAR(255),
    metadata JSONB
);

-- Indexes for common queries
CREATE INDEX idx_audit_logs_actor_id ON audit_logs(actor_id);
CREATE INDEX idx_audit_logs_actor_type ON audit_logs(actor_type);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource);
CREATE INDEX idx_audit_logs_resource_id ON audit_logs(resource_id);
CREATE INDEX idx_audit_logs_correlation_id ON audit_logs(correlation_id);
CREATE INDEX idx_audit_logs_deleted_at ON audit_logs(deleted_at);
```

## Advanced Usage

### Manual Audit Logging with Any Type

```go
// Request and response bodies can be any type - string, map, struct, etc.

// With string bodies
entry := &audit.AuditEntry{
    ActorID:      "user-123",
    RequestBody:  `{"username":"john"}`,
    ResponseBody: `{"success":true}`,
    // ... other fields
}

// With map bodies
entry := &audit.AuditEntry{
    ActorID:      "user-123",
    RequestBody:  map[string]interface{}{"username": "john"},
    ResponseBody: map[string]interface{}{"success": true},
    // ... other fields
}

// With struct bodies
type LoginRequest struct {
    Username string `json:"username"`
}
entry := &audit.AuditEntry{
    ActorID:      "user-123",
    RequestBody:  LoginRequest{Username: "john"},
    ResponseBody: map[string]bool{"success": true},
    // ... other fields
}

err := auditService.Log(ctx, entry)
```

### Manual Audit Logging

```go
// Log custom events programmatically
// Request and response bodies can be any type
entry := &audit.AuditEntry{
    ActorID:       "user-123",
    ActorType:     "user",
    Action:        "LOGIN",
    Resource:      "auth",
    ResourceID:    "",
    Description:   "User logged in successfully",
    Method:        "POST",
    Path:          "/auth/login",
    RequestBody:   map[string]string{"username": "john"},  // Can be any type
    ResponseBody:  map[string]bool{"success": true},       // Can be any type
    StatusCode:    200,
    IPAddress:     "192.168.1.1",
    CorrelationID: "trace-123",
    Metadata: map[string]interface{}{
        "login_method": "password",
    },
}

err := auditService.Log(ctx, entry)
```

## Best Practices

1. **Use Async Logging for High-Traffic Routes**: Set `AsyncLogging: true` to avoid blocking requests
2. **Limit Body Size**: Set appropriate `MaxBodySize` to prevent database bloat
3. **Filter Sensitive Data**: Always configure `SensitiveHeaders` to exclude authentication tokens
4. **Index Strategy**: Add database indexes on frequently queried fields
5. **Retention Policy**: Implement a data retention policy to archive or delete old audit logs
6. **Selective Auditing**: Only audit routes that require compliance or security tracking

## Configuration Reference

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Skip` | `func(*gin.Context) bool` | Skip health/metrics | Skip audit for certain requests |
| `GetActorID` | `func(*gin.Context) string` | Extract from "user_id" | Extract actor ID |
| `GetActorType` | `func(*gin.Context) string` | Returns "user" | Extract actor type |
| `GetAction` | `func(*gin.Context) string` | HTTP method | Extract action |
| `GetResource` | `func(*gin.Context) string` | First path segment | Extract resource name |
| `GetResourceID` | `func(*gin.Context) string` | From params/query | Extract resource ID |
| `GetDescription` | `func(*gin.Context) string` | Method + Path | Generate description |
| `IncludeRequestBody` | `bool` | `true` | Log request body |
| `IncludeResponseBody` | `bool` | `true` | Log response body |
| `IncludeRequestHeaders` | `bool` | `false` | Log request headers |
| `SensitiveHeaders` | `[]string` | Auth headers | Headers to exclude |
| `MaxBodySize` | `int` | `10240` (10KB) | Max body size to log |
| `AsyncLogging` | `bool` | `true` | Enable async logging |
| `GetMetadata` | `func(*gin.Context) map[string]interface{}` | `nil` | Extract custom metadata |

## License

Part of the go-base project.
