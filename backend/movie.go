package backend

import (
	"errors"
	"infinity-subtitle/backend/database"
	"time"
)

type Movie struct {
	ID                int               `json:"id"`
	Title             string            `json:"title"`
	DefaultLanguage   string            `json:"default_language"`
	Languages         map[string]string `json:"languages"`
	CreatedAt         time.Time         `json:"created_at"`
}

func NewMovie() *Movie {
	return &Movie{}
}

func (m *Movie) CreateMovie(title string, defaultLanguage string, languages map[string]string) error {
	if title == "" {
		return errors.New("title is required")
	}

	if defaultLanguage == "" {
		return errors.New("default language is required")
	}

	if len(languages) == 0 {
		return errors.New("subtitle languages are required")
	}

	m.Title = title
	m.DefaultLanguage = defaultLanguage
	m.Languages = languages

	db := database.GetDB()

	_, err := db.Exec("INSERT INTO movies (title, default_language, languages) VALUES (?, ?, ?)", m.Title, m.DefaultLanguage, m.Languages)
	if err != nil {
		return err
	}

	return nil
}

func (m *Movie) GetMovieByID(id int) (*Movie, error) {
	db := database.GetDB()
	row := db.QueryRow("SELECT * FROM movies WHERE id = ?", id)
	var movie Movie
	err := row.Scan(&movie.ID, &movie.Title, &movie.DefaultLanguage, &movie.Languages, &movie.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &movie, nil
}
