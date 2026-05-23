package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"cinema/booking-service/internal/db"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var queries *db.Queries

type BookingInput struct {
	UserID     int32  `json:"user_id"`
	ShowtimeID int32  `json:"showtime_id"`
	SeatNumber string `json:"seat_number"`
}

type ConfirmInput struct {
	BookingID  int32  `json:"booking_id"`
	UserID     int32  `json:"user_id"`
	TicketType string `json:"ticket_type"`
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
	mux.HandleFunc("GET /api/bookings", listBookingsHandler)
	mux.HandleFunc("POST /api/bookings/lock", lockSeatHandler)
	mux.HandleFunc("POST /api/bookings/confirm", confirmBookingHandler)

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

func lockSeatHandler(w http.ResponseWriter, r *http.Request) {
	var input BookingInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Lock for 5 minutes
	lockedUntil := sql.NullTime{
		Time:  javaTimeNowPlus5Mins(),
		Valid: true,
	}
	booking, err := queries.LockSeat(context.Background(), db.LockSeatParams{
		UserID:      input.UserID,
		ShowtimeID:  input.ShowtimeID,
		SeatNumber:  input.SeatNumber,
		LockedUntil: lockedUntil,
	})
	if err != nil {
		http.Error(w, "Seat already locked or reserved: "+err.Error(), http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(booking)
}

func confirmBookingHandler(w http.ResponseWriter, r *http.Request) {
	var input ConfirmInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ticketType := sql.NullString{
		String: input.TicketType,
		Valid:  input.TicketType != "",
	}
	booking, err := queries.ConfirmBooking(context.Background(), db.ConfirmBookingParams{
		ID:         input.BookingID,
		UserID:     input.UserID,
		TicketType: ticketType,
	})
	if err != nil {
		http.Error(w, "Failed to confirm booking: "+err.Error(), http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(booking)
}

// Helper time function
func javaTimeNowPlus5Mins() time.Time {
	// Import "time" should be added
	return time.Now().Add(5 * time.Minute)
}
