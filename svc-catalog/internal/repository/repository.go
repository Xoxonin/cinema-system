package repository

import (
	"context"
	"errors"

	"cinema/svc-catalog/internal/domain"
)

var ErrNotFound = errors.New("resource not found")

type MovieRepository interface {
	List(ctx context.Context) ([]domain.Movie, error)
	GetByID(ctx context.Context, id string) (domain.Movie, error)
	Create(ctx context.Context, movie domain.Movie) (domain.Movie, error)
	Update(ctx context.Context, movie domain.Movie) (domain.Movie, error)
	Delete(ctx context.Context, id string) error
}
