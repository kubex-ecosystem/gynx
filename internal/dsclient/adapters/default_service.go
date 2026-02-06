// Package adapters defines a generic repository interface for database operations.
package adapters

import "context"

type DefaultService[T any] struct {
	repo Repository[T]
}

func NewDefaultService[T any](r Repository[T]) *DefaultService[T] {
	return &DefaultService[T]{repo: r}
}

func (s *DefaultService[T]) Create(ctx context.Context, m *T) error {
	return s.repo.Create(ctx, m)
}
func (s *DefaultService[T]) GetByID(ctx context.Context, id string) (*T, error) {
	return s.repo.GetByID(ctx, id)
}
func (s *DefaultService[T]) Update(ctx context.Context, m *T) error {
	return s.repo.Update(ctx, m)
}
func (s *DefaultService[T]) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
func (s *DefaultService[T]) List(ctx context.Context, cond map[string]any) ([]T, error) {
	return s.repo.List(ctx, cond)
}
func (s *DefaultService[T]) BeginTx(ctx context.Context) (context.Context, error) {
	// return s.repo.BeginTx(ctx)
	return nil, nil // Placeholder implementation
}

type Service[T any] interface {
	Create(ctx context.Context, m *T) error
	GetByID(ctx context.Context, id string) (*T, error)
	Update(ctx context.Context, m *T) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filters map[string]any) ([]T, error)
	BeginTx(ctx context.Context) (context.Context, error)
}

type Repository[T any] interface {
	Create(ctx context.Context, m *T) error
	GetByID(ctx context.Context, id string) (*T, error)
	Update(ctx context.Context, m *T) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filters map[string]any) ([]T, error)
}