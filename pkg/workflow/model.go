package workflow

import (
	entity "go-base/pkg/datasource/postgres/entities"
	"time"
)

type Workflow struct {
	entity.BaseEntity
	ID             string         `gorm:"primaryKey;type:varchar"`
	WorkflowName   string         `gorm:"not null"`
	MaxAttempts    int            `gorm:"not null"`
	CurrentAttempt int            `gorm:"default:0"`
	Payload        string         `gorm:"type:json"`
	ProcessResults string         `gorm:"type:json"`
	Status         WorkflowResult `gorm:"default:'NEW'"`
	Finished       bool           `gorm:"default:false"`
	Duration       time.Duration  `gorm:"type:int64"`
	StartedTime    time.Time      `gorm:"type:timestamptz"`
	FinishedTime   time.Time      `gorm:"type:timestamptz"`
}
