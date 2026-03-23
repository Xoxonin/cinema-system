-- name: CreateMovie :one
INSERT INTO movies (title, description, duration_minutes, release_date)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetMovie :one
SELECT * FROM movies WHERE id = $1 LIMIT 1;

-- name: ListMovies :many
SELECT * FROM movies ORDER BY id LIMIT $1 OFFSET $2;
