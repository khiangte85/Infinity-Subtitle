package backend

import (
	"database/sql"
	"fmt"
	"infinity-subtitle/backend/database"
	"time"
)

type MovieQueue struct {
	ID             int        `json:"id"`
	Name           string     `json:"name"`
	Content        string     `json:"content"`
	SourceLanguage string     `json:"source_language"`
	TargetLanguage string     `json:"target_language"`
	Status         int        `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	ProcessedAt    *time.Time `json:"processed_at"`
}

type MovieQueueResponse struct {
	Movies     []MovieQueue `json:"movies"`
	Pagination Pagination   `json:"pagination"`
}

type AddToQueueRequest struct {
	Name           string `json:"name"`
	Content        string `json:"content"`
	SourceLanguage string `json:"source_language"`
	TargetLanguage string `json:"target_language"`
}

func NewMovieQueue() *MovieQueue {
	return &MovieQueue{}
}

func (mq *MovieQueue) GetQueue(pagination Pagination) (MovieQueueResponse, error) {
	db := database.GetDB()
	var response MovieQueueResponse
	var movies []MovieQueue

	// Get total count
	var total int
	err := db.QueryRow("SELECT COUNT(*) FROM movies_queue").Scan(&total)
	if err != nil {
		return response, fmt.Errorf("failed to get total count: %w", err)
	}

	// Calculate offset
	offset := (pagination.Page - 1) * pagination.RowsPerPage

	// Get paginated movies
	rows, err := db.Query(`
		SELECT id, name, source_language, target_language, status, created_at, processed_at 
		FROM movies_queue 
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, pagination.RowsPerPage, offset)
	if err != nil {
		return response, err
	}
	defer rows.Close()

	for rows.Next() {
		var movie MovieQueue
		var processedAt sql.NullTime
		err := rows.Scan(
			&movie.ID,
			&movie.Name,
			&movie.SourceLanguage,
			&movie.TargetLanguage,
			&movie.Status,
			&movie.CreatedAt,
			&processedAt,
		)
		if err != nil {
			return response, err
		}
		if processedAt.Valid {
			movie.ProcessedAt = &processedAt.Time
		}
		movies = append(movies, movie)
	}

	if err = rows.Err(); err != nil {
		return response, err
	}

	response = MovieQueueResponse{
		Movies: movies,
		Pagination: Pagination{
			SortBy:      pagination.SortBy,
			Descending:  pagination.Descending,
			Page:        pagination.Page,
			RowsPerPage: pagination.RowsPerPage,
			RowsNumber:  total,
		},
	}

	return response, nil
}

func (mq *MovieQueue) AddToQueue(req AddToQueueRequest) error {
	db := database.GetDB()

	_, err := db.Exec(`
		INSERT INTO movies_queue (
			name, content, source_language, target_language, status, created_at
		) VALUES (?, ?, ?, ?, 0, CURRENT_TIMESTAMP)
	`, req.Name, req.Content, req.SourceLanguage, req.TargetLanguage)

	if err != nil {
		return fmt.Errorf("failed to add movie to queue: %w", err)
	}

	return nil
}

func (mq *MovieQueue) DeleteFromQueue(id int) error {
	db := database.GetDB()

	_, err := db.Exec("DELETE FROM movies_queue WHERE id = ? AND status = 0", id)
	if err != nil {
		return fmt.Errorf("failed to delete movie from queue: %w", err)
	}

	return nil
}
