-- name: CreateRoom :one
INSERT INTO rooms (name, capacity, seating_layout) VALUES ($1, $2, CAST($3 AS JSONB)) RETURNING *;

-- name: ListRooms :many
SELECT * FROM rooms ORDER BY id;

-- name: CreateShowtime :one
INSERT INTO showtimes (movie_id, room_id, start_time, end_time, type)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: GetShowtime :one
SELECT * FROM showtimes WHERE id = $1 LIMIT 1;

-- name: ListShowtimesByMovie :many
SELECT * FROM showtimes WHERE movie_id = $1 ORDER BY start_time;
