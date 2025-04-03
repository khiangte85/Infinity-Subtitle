package backend

import (
	"encoding/json"
	"errors"
	"infinity-subtitle/backend/database"
	"log"
	"strings"
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

type ListMoviesResponse struct {
	Movies []Movie `json:"movies"`
	LastID int64   `json:"last_id"`
}

func (m *Movie) ListMovies(title string, sortBy string, sortDesc bool, lastID int64, limit int) (*ListMoviesResponse, error) {
	db := database.GetDB()
	query := "SELECT * FROM movies"
	args := []any{}

	// Handle search by title if provided
	if title != "" {
		query += " WHERE title LIKE ?"
		args = append(args, "%"+title+"%")
	}

	// Handle sorting
	if sortBy != "" {
		query += " ORDER BY " + sortBy
		if sortDesc {
			query += " DESC"
		} else {
			query += " ASC" 
		}
	}

	// Add cursor pagination
	if lastID > 0 {
		if strings.Contains(query, "WHERE") {
			query += " AND"
		} else {
			query += " WHERE"
		}
		
		if sortDesc {
			query += " id < ?"
		} else {
			query += " id > ?"
		}
		args = append(args, lastID)
	}

	// Add limit
	query += " LIMIT ?"
	args = append(args, limit)

	
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
	
	var lastMovieID int64
	if len(movies) > 0 {
		log.Printf("Movies %+v", movies)
		lastMovieID = int64(movies[len(movies)-1].ID)
		log.Printf("Last Movie ID %d", lastMovieID)
	}

	return &ListMoviesResponse{
		Movies: movies,
		LastID: lastMovieID,
	}, nil
}
