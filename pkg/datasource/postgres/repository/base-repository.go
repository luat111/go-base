package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type BaseRepository struct {
	DB     *gorm.DB
	entity any
}

func NewBaseRepository(db *gorm.DB, entity any) *BaseRepository {
	return &BaseRepository{DB: db, entity: entity}
}

// Transaction handling
func (r *BaseRepository) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.DB.Model(r.entity).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// Create a record
func (r *BaseRepository) Create(ctx context.Context, entity any) error {
	fmt.Println(entity)
	return r.DB.Model(r.entity).WithContext(ctx).Create(entity).Error
}

// Update a record
func (r *BaseRepository) Update(ctx context.Context, id string, entity any) error {
	return r.DB.Model(r.entity).WithContext(ctx).Where("id = ?", id).Updates(entity).Error
}

// Delete a record
func (r *BaseRepository) Delete(ctx context.Context, id string) error {
	return r.DB.Model(r.entity).WithContext(ctx).Delete(r.entity, id).Error
}

// Find records based on a query function
func (r *BaseRepository) Find(ctx context.Context, result any, queryFunc func(db *gorm.DB) *gorm.DB) error {
	db := r.DB.Model(r.entity).WithContext(ctx)

	// Apply query function
	db = queryFunc(db)

	// Execute the query
	return db.Find(result).Error
}

// Find by ID
func (r *BaseRepository) FindByID(ctx context.Context, id any) error {
	return r.DB.Model(r.entity).WithContext(ctx).First(r.entity, id).Error
}

// Pagination handling
func (r *BaseRepository) Paginate(ctx context.Context, page, pageSize int, result any, queryFunc func(db *gorm.DB) *gorm.DB) (int64, error) {
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
