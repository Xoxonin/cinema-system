package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"cinema/svc-catalog/internal/domain"
	"cinema/svc-catalog/internal/repository"
)

var ErrValidation = errors.New("validation error")

type CatalogService struct {
	repo repository.MovieRepository
}

func NewCatalogService(repo repository.MovieRepository) *CatalogService {
	return &CatalogService{repo: repo}
}

func (s *CatalogService) ListMovies(ctx context.Context) ([]domain.Movie, error) {
	return s.repo.List(ctx)
}

func (s *CatalogService) GetMovie(ctx context.Context, id string) (domain.Movie, error) {
	if strings.TrimSpace(id) == "" {
		return domain.Movie{}, fmt.Errorf("%w: id is required", ErrValidation)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *CatalogService) CreateMovie(ctx context.Context, in domain.CreateMovieInput) (domain.Movie, error) {
	if err := validateCreateInput(in); err != nil {
		return domain.Movie{}, err
	}

	now := time.Now().UTC()
	movie := domain.Movie{
		ID:              newID(),
		Title:           strings.TrimSpace(in.Title),
		Description:     strings.TrimSpace(in.Description),
		DurationMinutes: in.DurationMinutes,
		ReleaseYear:     in.ReleaseYear,
		Genres:          trimNonEmpty(in.Genres),
		Language:        strings.TrimSpace(in.Language),
		Country:         strings.TrimSpace(in.Country),
		Director:        strings.TrimSpace(in.Director),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	return s.repo.Create(ctx, movie)
}

func (s *CatalogService) UpdateMovie(ctx context.Context, id string, in domain.UpdateMovieInput) (domain.Movie, error) {
	if strings.TrimSpace(id) == "" {
		return domain.Movie{}, fmt.Errorf("%w: id is required", ErrValidation)
	}
	if err := validateUpdateInput(in); err != nil {
		return domain.Movie{}, err
	}

	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Movie{}, err
	}

	current.Title = strings.TrimSpace(in.Title)
	current.Description = strings.TrimSpace(in.Description)
	current.DurationMinutes = in.DurationMinutes
	current.ReleaseYear = in.ReleaseYear
	current.Genres = trimNonEmpty(in.Genres)
	current.Language = strings.TrimSpace(in.Language)
	current.Country = strings.TrimSpace(in.Country)
	current.Director = strings.TrimSpace(in.Director)
	current.UpdatedAt = time.Now().UTC()

	return s.repo.Update(ctx, current)
}

func (s *CatalogService) DeleteMovie(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("%w: id is required", ErrValidation)
	}

	return s.repo.Delete(ctx, id)
}

func validateCreateInput(in domain.CreateMovieInput) error {
	if strings.TrimSpace(in.Title) == "" {
		return fmt.Errorf("%w: title is required", ErrValidation)
	}
	if in.DurationMinutes <= 0 {
		return fmt.Errorf("%w: durationMinutes must be > 0", ErrValidation)
	}
	if in.ReleaseYear < 1888 {
		return fmt.Errorf("%w: releaseYear is invalid", ErrValidation)
	}
	return nil
}

func validateUpdateInput(in domain.UpdateMovieInput) error {
	if strings.TrimSpace(in.Title) == "" {
		return fmt.Errorf("%w: title is required", ErrValidation)
	}
	if in.DurationMinutes <= 0 {
		return fmt.Errorf("%w: durationMinutes must be > 0", ErrValidation)
	}
	if in.ReleaseYear < 1888 {
		return fmt.Errorf("%w: releaseYear is invalid", ErrValidation)
	}
	return nil
}

func trimNonEmpty(input []string) []string {
	out := make([]string, 0, len(input))
	for _, item := range input {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func newID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
