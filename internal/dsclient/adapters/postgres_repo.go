// Package adapters provides a generic PostgreSQL repository implementation using GORM.
package adapters

import (
	"context"

	"gorm.io/gorm"
)

type PGRepository[T any] struct {
	db *gorm.DB
}

func NewPGRepository[T any](db *gorm.DB) *PGRepository[T] {
	return &PGRepository[T]{db: db}
}

func (r *PGRepository[T]) Create(ctx context.Context, m *T) error {
	return r.db.WithContext(ctx).Create(m).Error
}
func (r *PGRepository[T]) GetByID(ctx context.Context, id string) (*T, error) {
	var out T
	if err := r.db.WithContext(ctx).First(&out, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &out, nil
}
func (r *PGRepository[T]) Update(ctx context.Context, m *T) error {
	return r.db.WithContext(ctx).Save(m).Error
}
func (r *PGRepository[T]) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(new(T), "id = ?", id).Error
}
func (r *PGRepository[T]) List(ctx context.Context, cond any) ([]T, error) {
	var list []T
	if err := r.db.WithContext(ctx).Find(&list, cond).Error; err != nil {
		return nil, err
	}
	return list, nil
}


