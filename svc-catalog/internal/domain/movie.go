package domain

import "time"

type Movie struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	DurationMinutes int       `json:"durationMinutes"`
	ReleaseYear     int       `json:"releaseYear"`
	Genres          []string  `json:"genres"`
	Language        string    `json:"language"`
	Country         string    `json:"country"`
	Director        string    `json:"director"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type CreateMovieInput struct {
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	DurationMinutes int      `json:"durationMinutes"`
	ReleaseYear     int      `json:"releaseYear"`
	Genres          []string `json:"genres"`
	Language        string   `json:"language"`
	Country         string   `json:"country"`
	Director        string   `json:"director"`
}

type UpdateMovieInput struct {
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	DurationMinutes int      `json:"durationMinutes"`
	ReleaseYear     int      `json:"releaseYear"`
	Genres          []string `json:"genres"`
	Language        string   `json:"language"`
	Country         string   `json:"country"`
	Director        string   `json:"director"`
}
