-- name: CreateBooking :one
INSERT INTO bookings (user_id, showtime_id, seat_number, status)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetBooking :one
SELECT * FROM bookings WHERE id = $1 LIMIT 1;

-- name: ListBookingsByUser :many
SELECT * FROM bookings WHERE user_id = $1 ORDER BY created_at DESC;

-- name: ListBookingsByShowtime :many
SELECT * FROM bookings WHERE showtime_id = $1;
