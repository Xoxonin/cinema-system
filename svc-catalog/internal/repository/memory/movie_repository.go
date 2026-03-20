package memory

import (
	"context"
	"sort"
	"sync"

	"cinema/svc-catalog/internal/domain"
	"cinema/svc-catalog/internal/repository"
)

type MovieRepository struct {
	mu     sync.RWMutex
	movies map[string]domain.Movie
}

func NewMovieRepository() *MovieRepository {
	return &MovieRepository{movies: make(map[string]domain.Movie)}
}

func (r *MovieRepository) List(_ context.Context) ([]domain.Movie, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]domain.Movie, 0, len(r.movies))
	for _, movie := range r.movies {
		out = append(out, movie)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})

	return out, nil
}

func (r *MovieRepository) GetByID(_ context.Context, id string) (domain.Movie, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	movie, ok := r.movies[id]
	if !ok {
		return domain.Movie{}, repository.ErrNotFound
	}

	return movie, nil
}

func (r *MovieRepository) Create(_ context.Context, movie domain.Movie) (domain.Movie, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.movies[movie.ID] = movie
	return movie, nil
}

func (r *MovieRepository) Update(_ context.Context, movie domain.Movie) (domain.Movie, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.movies[movie.ID]; !ok {
		return domain.Movie{}, repository.ErrNotFound
	}

	r.movies[movie.ID] = movie
	return movie, nil
}

func (r *MovieRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.movies[id]; !ok {
		return repository.ErrNotFound
	}

	delete(r.movies, id)
	return nil
}
