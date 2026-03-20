package httpapi

import (
	"net/http"

	"cinema/svc-catalog/internal/service"
)

func NewRouter(catalog *service.CatalogService) http.Handler {
	h := NewHandler(catalog)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("GET /movies", h.ListMovies)
	mux.HandleFunc("GET /movies/{id}", h.GetMovie)
	mux.HandleFunc("POST /movies", h.CreateMovie)
	mux.HandleFunc("PUT /movies/{id}", h.UpdateMovie)
	mux.HandleFunc("DELETE /movies/{id}", h.DeleteMovie)

	return mux
}
