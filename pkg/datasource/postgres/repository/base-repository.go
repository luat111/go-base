package repository

import (
	"context"

	"gorm.io/gorm"
)

type BaseRepository struct {
	DB *gorm.DB
}

func NewBaseRepository(db *gorm.DB) *BaseRepository {
	return &BaseRepository{DB: db}
}

// Transaction handling
func (r *BaseRepository) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// Create a record
func (r *BaseRepository) Create(ctx context.Context, entity any) error {
	return r.DB.WithContext(ctx).Create(entity).Error
}

// Update a record
func (r *BaseRepository) Update(ctx context.Context, entity any) error {
	return r.DB.WithContext(ctx).Save(entity).Error
}

// Delete a record
func (r *BaseRepository) Delete(ctx context.Context, entity any) error {
	return r.DB.WithContext(ctx).Delete(entity).Error
}

// Find by ID
func (r *BaseRepository) FindByID(ctx context.Context, id any, entity any) error {
	return r.DB.WithContext(ctx).First(entity, id).Error
}

// Pagination handling
func (r *BaseRepository) Paginate(ctx context.Context, page, pageSize int, result any, queryFunc func(db *gorm.DB) *gorm.DB) (int64, error) {
	var total int64
	db := r.DB.WithContext(ctx)

	// Apply query function
	db = queryFunc(db)

	// Count total records
	if err := db.Count(&total).Error; err != nil {
		return 0, err
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Find(result).Error; err != nil {
		return 0, err
	}

	return total, nil
}
