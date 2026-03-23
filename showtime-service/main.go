package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"cinema/showtime-service/internal/db"
	_ "github.com/lib/pq"
)

var queries *db.Queries

type RoomInput struct {
	Name     string `json:"name"`
	Capacity int32  `json:"capacity"`
}

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/showtime?sslmode=disable"
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbConn.Close()

	queries = db.New(dbConn)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /rooms", listRoomsHandler)
	mux.HandleFunc("POST /rooms", createRoomHandler)
	mux.HandleFunc("GET /showtimes", listShowtimesHandler)
	mux.HandleFunc("GET /showtimes/{id}", getShowtimeHandler)

	log.Println("Showtime Service starting on port 8083")
	log.Fatal(http.ListenAndServe(":8083", mux))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func listRoomsHandler(w http.ResponseWriter, r *http.Request) {
	rooms, err := queries.ListRooms(context.Background())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(rooms)
}

func createRoomHandler(w http.ResponseWriter, r *http.Request) {
	var input RoomInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	room, err := queries.CreateRoom(context.Background(), db.CreateRoomParams{
		Name:     input.Name,
		Capacity: input.Capacity,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(room)
}

func listShowtimesHandler(w http.ResponseWriter, r *http.Request) {
	movieIDStr := r.URL.Query().Get("movie_id")
	movieID, err := strconv.Atoi(movieIDStr)
	if err != nil {
		http.Error(w, "Invalid movie_id", http.StatusBadRequest)
		return
	}
	showtimes, err := queries.ListShowtimesByMovie(context.Background(), int32(movieID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(showtimes)
}

func getShowtimeHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	st, err := queries.GetShowtime(context.Background(), int32(id))
	if err != nil {
		http.Error(w, "Showtime not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(st)
}
