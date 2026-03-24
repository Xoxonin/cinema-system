package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"cinema/catalog-service/internal/db"

	_ "github.com/jackc/pgx/stdlib"
)

var queries *db.Queries

type MovieInput struct {
	Title           string `json:"title"`
	Description     string `json:"description"`
	DurationMinutes int32  `json:"duration_minutes"`
}

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/catalog?sslmode=disable"
	}

	dbConn, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbConn.Close()

	queries = db.New(dbConn)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /movies", listMoviesHandler)
	mux.HandleFunc("GET /movies/{id}", getMovieHandler)
	mux.HandleFunc("POST /movies", createMovieHandler)

	log.Println("Catalog Service starting on port 8082")
	log.Fatal(http.ListenAndServe(":8082", mux))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func listMoviesHandler(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	limit, _ := strconv.Atoi(limitStr)
	if limit == 0 {
		limit = 100
	}
	offset, _ := strconv.Atoi(offsetStr)

	movies, err := queries.ListMovies(context.Background(), db.ListMoviesParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(movies)
}

func getMovieHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	movie, err := queries.GetMovie(context.Background(), int32(id))
	if err != nil {
		http.Error(w, "Movie not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(movie)
}

func createMovieHandler(w http.ResponseWriter, r *http.Request) {
	var input MovieInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	movie, err := queries.CreateMovie(context.Background(), db.CreateMovieParams{
		Title:           input.Title,
		Description:     sql.NullString{String: input.Description, Valid: input.Description != ""},
		DurationMinutes: input.DurationMinutes,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(movie)
}
