package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"cinema/svc-catalog/internal/domain"
	"cinema/svc-catalog/internal/repository"
	"cinema/svc-catalog/internal/service"
)

type Handler struct {
	catalog *service.CatalogService
}

func NewHandler(catalog *service.CatalogService) *Handler {
	return &Handler{catalog: catalog}
}

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) ListMovies(w http.ResponseWriter, r *http.Request) {
	movies, err := h.catalog.ListMovies(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list movies")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"items": movies})
}

func (h *Handler) GetMovie(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	movie, err := h.catalog.GetMovie(r.Context(), id)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, movie)
}

func (h *Handler) CreateMovie(w http.ResponseWriter, r *http.Request) {
	var in domain.CreateMovieInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	movie, err := h.catalog.CreateMovie(r.Context(), in)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, movie)
}

func (h *Handler) UpdateMovie(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var in domain.UpdateMovieInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	movie, err := h.catalog.UpdateMovie(r.Context(), id, in)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, movie)
}

func (h *Handler) DeleteMovie(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.catalog.DeleteMovie(r.Context(), id); err != nil {
		writeDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeDomainError(w http.ResponseWriter, err error) {
	if errors.Is(err, service.ErrValidation) {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		respondError(w, http.StatusNotFound, "movie not found")
		return
	}

	respondError(w, http.StatusInternalServerError, "internal error")
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
