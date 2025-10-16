package entity

import (
	"time"
)

type BaseEntity struct {
	ID        uint       `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`
	DeletedAt *time.Time `gorm:"index"` // Soft delete support
}
