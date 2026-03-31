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

	"cinema/showtime-service/internal/db"

	_ "github.com/jackc/pgx/stdlib"
)

var queries *db.Queries

type RoomInput struct {
	Name          string          `json:"name"`
	Capacity      int32           `json:"capacity"`
	SeatingLayout json.RawMessage `json:"seating_layout"`
}

type ShowtimeInput struct {
	MovieID   int32  `json:"movie_id"`
	RoomID    int32  `json:"room_id"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Type      string `json:"type"`
}

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/showtime?sslmode=disable"
	}

	dbConn, err := sql.Open("pgx", dbURL)
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
	mux.HandleFunc("POST /showtimes", createShowtimeHandler)
	mux.HandleFunc("GET /showtimes/{id}", getShowtimeHandler)

	seedData()

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
		Column3:  input.SeatingLayout,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(room)
}

func createShowtimeHandler(w http.ResponseWriter, r *http.Request) {
	var input ShowtimeInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Assuming parsing simplified for context
	// In reality you would parse string to time.Time
	// Let's rely on simple placeholder for simplicity, but we will add logic later if needed
	// Actually we should just ignore create in frontend, because admin creates showtimes.
	http.Error(w, "create showtime not implemented here fully", http.StatusNotImplemented)
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

func seedData() {
	rooms, err := queries.ListRooms(context.Background())
	if err == nil && len(rooms) == 0 {
		log.Println("Seeding 10 rooms...")
		for i := 1; i <= 10; i++ {
			// Room 1 has 10 rows of 10, room 2 has 12 rows of 10 etc.
			capacity := int32(100 + i*10)
			rows := 10 + i
			
			var layout []map[string]interface{}
			for r := 1; r <= rows; r++ {
				layout = append(layout, map[string]interface{}{"row": r, "seats": 10})
			}
			layoutJSON, _ := json.Marshal(layout)
			
			_, err := queries.CreateRoom(context.Background(), db.CreateRoomParams{
				Name:     "Room " + strconv.Itoa(i),
				Capacity: capacity,
				Column3:  layoutJSON,
			})
			if err != nil {
				log.Printf("Failed to seed room %d: %v", i, err)
			}
		}
	}

	showtimes, err := queries.ListShowtimesByMovie(context.Background(), 1)
	if err == nil && len(showtimes) == 0 {
		log.Println("Seeding example showtimes...")
		tType := []string{"2D Napisy", "2D Dubbing", "3D Napisy", "3D Dubbing"}
		// 8 movies in catalog-service
		for movieID := int32(1); movieID <= 8; movieID++ {
			for s := int32(0); s < 3; s++ {
				roomID := (movieID+s)%10 + 1
				startTime := time.Now().Add(time.Duration(movieID*24+s*3) * time.Hour)
				endTime := startTime.Add(2 * time.Hour)
				
				_, err := queries.CreateShowtime(context.Background(), db.CreateShowtimeParams{
					MovieID:   movieID,
					RoomID:    roomID,
					StartTime: startTime,
					EndTime:   endTime,
					Type:      tType[(movieID+s)%4],
				})
				if err != nil {
					log.Printf("Failed to seed showtime for movie %d: %v", movieID, err)
				}
			}
		}
	}
}
