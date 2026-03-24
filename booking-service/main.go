package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"cinema/booking-service/internal/db"

	_ "github.com/jackc/pgx/stdlib"
)

var queries *db.Queries

type BookingInput struct {
	UserID     int32  `json:"user_id"`
	ShowtimeID int32  `json:"showtime_id"`
	SeatNumber string `json:"seat_number"`
}

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/booking?sslmode=disable"
	}

	dbConn, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbConn.Close()

	queries = db.New(dbConn)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /bookings", listBookingsHandler)
	mux.HandleFunc("POST /bookings", createBookingHandler)

	log.Println("Booking Service starting on port 8084")
	log.Fatal(http.ListenAndServe(":8084", mux))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func listBookingsHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	showtimeIDStr := r.URL.Query().Get("showtime_id")

	var bookings []db.Booking
	var err error

	if userIDStr != "" {
		userID, _ := strconv.Atoi(userIDStr)
		bookings, err = queries.ListBookingsByUser(context.Background(), int32(userID))
	} else if showtimeIDStr != "" {
		showtimeID, _ := strconv.Atoi(showtimeIDStr)
		bookings, err = queries.ListBookingsByShowtime(context.Background(), int32(showtimeID))
	} else {
		http.Error(w, "user_id or showtime_id required", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(bookings)
}

func createBookingHandler(w http.ResponseWriter, r *http.Request) {
	var input BookingInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	booking, err := queries.CreateBooking(context.Background(), db.CreateBookingParams{
		UserID:     input.UserID,
		ShowtimeID: input.ShowtimeID,
		SeatNumber: input.SeatNumber,
		Status:     "reserved",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(booking)
}
