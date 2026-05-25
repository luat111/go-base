package repository

import (
	"context"

	"gorm.io/gorm"
)

type BaseRepository[T any] struct {
	DB     *gorm.DB
	entity *T
}

func NewBaseRepository[T any](db *gorm.DB, entity *T) *BaseRepository[T] {
	return &BaseRepository[T]{DB: db, entity: entity}
}

// Transaction handling
func (r *BaseRepository[T]) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// Create a record
func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) error {
	return gorm.G[T](r.DB).Create(ctx, entity)
}

// Update a record
func (r *BaseRepository[T]) Update(ctx context.Context, id string, entity T) error {
	_, err := gorm.G[T](r.DB).Where("id = ?", id).Updates(ctx, entity)
	return err
}

// Delete a record
func (r *BaseRepository[T]) DeleteById(ctx context.Context, id string) error {
	_, err := gorm.G[T](r.DB).Where("id = ?", id).Delete(ctx)
	return err
}

// Find records based on a query function
func (r *BaseRepository[T]) Find(ctx context.Context, result any, queryFunc func(db *gorm.DB) *gorm.DB) error {
	db := r.DB.Model(r.entity).WithContext(ctx)

	// Apply query function
	db = queryFunc(db)

	// Execute the query
	return db.Find(result).Error
}

// Find by ID
func (r *BaseRepository[T]) FindByID(ctx context.Context, id any) (T, error) {
	return gorm.G[T](r.DB).Where("id = ?", id).First(ctx)
}

// Check Exist
func (r *BaseRepository[T]) IsIdExisted(ctx context.Context, id string) bool {
	_, err := r.FindByID(ctx, id)
	return err == nil
}

// Pagination handling
func (r *BaseRepository[T]) Paginate(ctx context.Context, page, pageSize int, result any, queryFunc func(db *gorm.DB) *gorm.DB) (int64, error) {
	var total int64
	db := r.DB.Model(r.entity).WithContext(ctx)

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
