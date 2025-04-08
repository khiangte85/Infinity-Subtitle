package backend

import (
	"encoding/json"
	"errors"
	"fmt"
	"infinity-subtitle/backend/database"
	"strings"
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
	return &Movie{
		Languages: make(map[string]string),
	}
}

func (m *Movie) CreateMovie(title string, defaultLanguage string, languages map[string]string) error {
	// Input validation
	if strings.TrimSpace(title) == "" {
		return errors.New("title is required")
	}

	if strings.TrimSpace(defaultLanguage) == "" {
		return errors.New("default language is required")
	}

	if len(languages) == 0 {
		return errors.New("subtitle languages are required")
	}

	jsonLanguages, err := json.Marshal(languages)
	if err != nil {
		return fmt.Errorf("failed to marshal languages: %w", err)
	}

	m.Title = title
	m.DefaultLanguage = defaultLanguage

	db := database.GetDB()
	if db == nil {
		return errors.New("database connection is nil")
	}

	_, err = db.Exec("INSERT INTO movies (title, default_language, languages) VALUES (?, ?, ?)",
		m.Title,
		m.DefaultLanguage,
		jsonLanguages,
	)
	if err != nil {
		return fmt.Errorf("failed to insert movie: %w", err)
	}

	return nil
}

func (m *Movie) GetMovieByID(id int) (*Movie, error) {
	if id <= 0 {
		return nil, errors.New("invalid movie ID")
	}

	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database connection is nil")
	}

	// Only select needed fields
	row := db.QueryRow("SELECT id, title, default_language, languages, created_at FROM movies WHERE id = ?", id)
	var languages []byte

	err := row.Scan(&m.ID, &m.Title, &m.DefaultLanguage, &languages, &m.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to scan movie: %w", err)
	}

	err = json.Unmarshal(languages, &m.Languages)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal languages: %w", err)
	}

	return m, nil
}

func (m *Movie) UpdateMovie(movie Movie) error {
	if movie.ID <= 0 {
		return errors.New("invalid movie ID")
	}

	db := database.GetDB()
	if db == nil {
		return errors.New("database connection is nil")
	}

	jsonLanguages, err := json.Marshal(movie.Languages)
	if err != nil {
		return fmt.Errorf("failed to marshal languages: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	_, err = tx.Exec("UPDATE movies SET title = ?, default_language = ?, languages = ? WHERE id = ?",
		movie.Title,
		movie.DefaultLanguage,
		jsonLanguages,
		movie.ID)
	if err != nil {
		return err
	}

	// rows, err := tx.Query("SELECT * FROM subtitles WHERE movie_id = ?", movie.ID)
	// if err != nil {
	// 	return err
	// }
	// defer rows.Close()

	// for rows.Next() {
	// 	var subtitle Subtitle
	// 	var contentBytes []byte
	// 	err := rows.Scan(&subtitle.ID, &subtitle.MovieID, &subtitle.SlNo, &subtitle.StartTime, &subtitle.EndTime, &contentBytes, &subtitle.CreatedAt, &subtitle.UpdatedAt)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	err = json.Unmarshal(contentBytes, &subtitle.Content)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	// Update content, if language key doesnt exist, add key with empty value
	// 	for key := range movie.Languages {
	// 		// check if key exists in content
	// 		if _, ok := subtitle.Content[key]; !ok {
	// 			subtitle.Content[key] = ""
	// 		}
	// 	}

	// 	// if language key removed, remove key from content
	// 	for key := range subtitle.Content {
	// 		if _, ok := movie.Languages[key]; !ok {
	// 			delete(subtitle.Content, key)
	// 		}
	// 	}

	// 	// update subtitle
	// 	contentJson, err := json.Marshal(subtitle.Content)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	_, err = tx.Exec("UPDATE subtitles SET content = ? WHERE id = ?", contentJson, subtitle.ID)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	err = tx.Commit()
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
