package workflow

import (
	entity "go-base/pkg/datasource/postgres/entities"
)

type Workflow struct {
	entity.BaseEntity
	WorkflowName   string `gorm:"not null"`
	MaxAttempts    int    `gorm:"not null"`
	CurrentAttempt int    `gorm:"default:0"`
	Payload        string `gorm:"type:json"`
	ProcessResults string `gorm:"type:json"`
	Status         string `gorm:"default:'NEW'"`
}
