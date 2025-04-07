package backend

import (
	"encoding/json"
	"errors"
	"infinity-subtitle/backend/database"
	"time"
)

type Movie struct {
	ID              int               `json:"id"`
	Title           string            `json:"title"`
	DefaultLanguage string            `json:"default_language"`
	Languages       map[string]string `json:"languages"`
	CreatedAt       time.Time         `json:"created_at"`
}

type ListMoviesResponse struct {
	Movies     []Movie    `json:"movies"`
	Pagination Pagination `json:"pagination"`
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

	jsonLanguages, err := json.Marshal(languages)
	if err != nil {
		return err
	}

	m.Title = title
	m.DefaultLanguage = defaultLanguage

	db := database.GetDB()

	_, err = db.Exec("INSERT INTO movies (title, default_language, languages) VALUES (?, ?, ?)", m.Title, m.DefaultLanguage, jsonLanguages)
	if err != nil {
		return err
	}

	return nil
}

func (m *Movie) GetMovieByID(id int) (*Movie, error) {
	db := database.GetDB()
	row := db.QueryRow("SELECT * FROM movies WHERE id = ?", id)
	var movie Movie
	var languages []byte

	err := row.Scan(&movie.ID, &movie.Title, &movie.DefaultLanguage, &languages, &movie.CreatedAt)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(languages, &movie.Languages)
	if err != nil {
		return nil, err
	}

	return &movie, nil
}

func (m *Movie) UpdateMovie(id int, title string, defaultLanguage string, languages map[string]string) error {
	db := database.GetDB()

	jsonLanguages, err := json.Marshal(languages)
	if err != nil {
		return err
	}

	_, err = db.Exec("UPDATE movies SET title = ?, default_language = ?, languages = ? WHERE id = ?", m.Title, m.DefaultLanguage, jsonLanguages, id)
	if err != nil {
		return err
	}

	return nil
}

func (m *Movie) DeleteMovie(id int) error {
	db := database.GetDB()
	_, err := db.Exec("DELETE FROM movies WHERE id = ?", id)
	return err
}

func (m *Movie) ListMovies(title string, pagination Pagination) (*ListMoviesResponse, error) {
	db := database.GetDB()

	query := "SELECT COUNT(id) FROM movies"
	args := []any{}
	if title != "" {
		query += " WHERE title LIKE ?"
		args = append(args, "%"+title+"%")
	}
	row := db.QueryRow(query, args...)
	var rowsNumber int
	err := row.Scan(&rowsNumber)
	if err != nil {
		return nil, err
	}
	pagination.RowsNumber = rowsNumber

	query = "SELECT * FROM movies"
	args = []any{}

	// Handle search by title if provided
	if title != "" {
		query += " WHERE title LIKE ?"
		args = append(args, "%"+title+"%")
	}

	// Handle sorting
	if pagination.SortBy != "" {
		query += " ORDER BY " + pagination.SortBy
		if pagination.Descending {
			query += " DESC"
		} else {
			query += " ASC"
		}
	}

	// Add pagination
	offset := (pagination.Page - 1) * pagination.RowsPerPage
	query += " LIMIT ? OFFSET ?"
	args = append(args, pagination.RowsPerPage, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []Movie
	for rows.Next() {
		var movie Movie
		var languages []byte
		err := rows.Scan(&movie.ID, &movie.Title, &movie.DefaultLanguage, &languages, &movie.CreatedAt)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(languages, &movie.Languages)
		if err != nil {
			return nil, err
		}

		movies = append(movies, movie)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &ListMoviesResponse{
		Movies:     movies,
		Pagination: pagination,
	}, nil
}
