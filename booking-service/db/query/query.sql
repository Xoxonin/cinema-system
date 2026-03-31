-- name: CreateBooking :one
INSERT INTO bookings (user_id, showtime_id, seat_number, status, ticket_type)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: LockSeat :one
INSERT INTO bookings (user_id, showtime_id, seat_number, status, locked_until)
VALUES ($1, $2, $3, 'locked', $4)
ON CONFLICT (showtime_id, seat_number) DO UPDATE
SET user_id = EXCLUDED.user_id,
    status = 'locked',
    locked_until = EXCLUDED.locked_until
WHERE bookings.status != 'reserved' AND (bookings.status != 'locked' OR bookings.locked_until < NOW())
RETURNING *;

-- name: ConfirmBooking :one
UPDATE bookings
SET status = 'reserved', ticket_type = $2, locked_until = NULL
WHERE id = $1 AND user_id = $3 AND status = 'locked' AND locked_until >= NOW()
RETURNING *;

-- name: GetBooking :one
SELECT * FROM bookings WHERE id = $1 LIMIT 1;

-- name: ListBookingsByUser :many
SELECT * FROM bookings WHERE user_id = $1 ORDER BY created_at DESC;

-- name: ListBookingsByShowtime :many
SELECT * FROM bookings WHERE showtime_id = $1;
