package backend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"infinity-subtitle/backend/database"
	"os"
	"path/filepath"
	"strconv"

	// "strconv"
	"strings"
	"time"
)

type Subtitle struct {
	ID        int               `json:"id"`
	MovieID   int               `json:"movie_id"`
	SlNo      int               `json:"sl_no"`
	StartTime string            `json:"start_time"`
	EndTime   string            `json:"end_time"`
	Content   map[string]string `json:"content"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type SubtitleResponse struct {
	Subtitles  []Subtitle `json:"subtitles"`
	Pagination Pagination `json:"pagination"`
}

type ExportResponse struct {
	FilePath string `json:"file_path"`
}

func NewSubtitle() *Subtitle {
	return &Subtitle{}
}

func (s Subtitle) GetSubtitlesByMovieID(movieID int, pagination Pagination) (SubtitleResponse, error) {
	db := database.GetDB()
	var response SubtitleResponse
	var subtitles []Subtitle

	// Get total count
	var total int
	err := db.QueryRow("SELECT COUNT(*) FROM subtitles WHERE movie_id = ?", movieID).Scan(&total)
	if err != nil {
		return response, fmt.Errorf("failed to get total count: %w", err)
	}

	// Calculate offset
	offset := (pagination.Page - 1) * pagination.RowsPerPage

	// Get paginated subtitles
	rows, err := db.Query(`
		SELECT id, movie_id, sl_no, start_time, end_time, content, created_at, updated_at 
		FROM subtitles 
		WHERE movie_id = ? 
		ORDER BY sl_no ASC
		LIMIT ? OFFSET ?
	`, movieID, pagination.RowsPerPage, offset)
	if err != nil {
		return response, err
	}
	defer rows.Close()

	for rows.Next() {
		var subtitle Subtitle
		var contentJson []byte
		err := rows.Scan(&subtitle.ID, &subtitle.MovieID, &subtitle.SlNo, &subtitle.StartTime, &subtitle.EndTime, &contentJson, &subtitle.CreatedAt, &subtitle.UpdatedAt)
		if err != nil {
			return response, err
		}

		err = json.Unmarshal(contentJson, &subtitle.Content)
		if err != nil {
			return response, fmt.Errorf("failed to unmarshal content: %w", err)
		}
		subtitles = append(subtitles, subtitle)
	}

	if err = rows.Err(); err != nil {
		return response, err
	}

	response = SubtitleResponse{
		Subtitles: subtitles,
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

func (s Subtitle) UpdateSubtitle(subtitle Subtitle) error {
	db := database.GetDB()
	if db == nil {
		return errors.New("database connection is nil")
	}

	contentJson, err := json.Marshal(subtitle.Content)
	if err != nil {
		return fmt.Errorf("failed to marshal content: %w", err)
	}

	_, err = db.Exec("UPDATE subtitles SET content = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", contentJson, subtitle.ID)
	if err != nil {
		return fmt.Errorf("failed to update subtitle: %w", err)
	}

	return nil
}

func (s Subtitle) ImportFromSRTFile(movie Movie, fileContent string) error {
	db := database.GetDB()
	if db == nil {
		return errors.New("database connection is nil")
	}

	fileContent = strings.TrimSpace(fileContent)
	lines := strings.Split(fileContent, "\n")
	var subtitles []Subtitle
	subtitle := &Subtitle{
		SlNo:      0,
		MovieID:   movie.ID,
		StartTime: "",
		EndTime:   "",
		Content:   make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for key := range movie.Languages {
		subtitle.Content[key] = ""
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		slNo, err := strconv.Atoi(line)
		if err == nil && subtitle.SlNo == 0 {
			subtitle.SlNo = slNo
			continue
		}

		parts := strings.Split(line, "-->")
		if len(parts) == 2 {
			subtitle.StartTime = strings.TrimSpace(parts[0])
			subtitle.EndTime = strings.TrimSpace(parts[1])
			continue
		}

		subtitle.Content[movie.DefaultLanguage] = line
		subtitles = append(subtitles, *subtitle)

		// Reset subtitle for next iteration
		subtitle.SlNo = 0
		subtitle.StartTime = ""
		subtitle.EndTime = ""
		subtitle.Content = make(map[string]string)
		for key := range movie.Languages {
			subtitle.Content[key] = ""
		}
	}

	// delete all subtitles for the movie
	_, err := db.Exec("DELETE FROM subtitles WHERE movie_id = ?", movie.ID)
	if err != nil {
		return fmt.Errorf("failed to delete existing subtitles: %w", err)
	}

	// Start a transaction for bulk insert
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare the bulk insert statement
	stmt, err := tx.Prepare(`
		INSERT INTO subtitles (movie_id, sl_no, start_time, end_time, content, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Execute bulk insert
	for _, subtitle := range subtitles {
		contentJson, err := json.Marshal(subtitle.Content)
		if err != nil {
			return fmt.Errorf("failed to marshal content: %w", err)
		}

		_, err = stmt.Exec(
			subtitle.MovieID,
			subtitle.SlNo,
			subtitle.StartTime,
			subtitle.EndTime,
			contentJson,
		)
		if err != nil {
			return fmt.Errorf("failed to execute statement: %w", err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s Subtitle) TranslateSubtitles(movieId int, sourceLanguage string, targetLanguage string) error {
	db := database.GetDB()
	if db == nil {
		return errors.New("database connection is nil")
	}

	movie := NewMovie()
	movie, err := movie.GetMovieByID(movieId)
	if err != nil {
		return fmt.Errorf("failed to get movie: %w", err)
	}

	translationService, err := NewTranslationService()
	if err != nil {
		return fmt.Errorf("failed to create translation service: %w", err)
	}
	defer translationService.Close()

	ctx := context.Background()

	// Get all subtitles for the movie
	rows, err := db.Query(`
		SELECT id, movie_id, content
		FROM subtitles 
		WHERE movie_id = ? 
		ORDER BY sl_no ASC
	`, movieId)
	if err != nil {
		return fmt.Errorf("failed to get subtitles: %w", err)
	}
	defer rows.Close()

	var subtitles []Subtitle
	for rows.Next() {
		var subtitle Subtitle
		var contentJson []byte
		err := rows.Scan(&subtitle.ID, &subtitle.MovieID, &contentJson)
		if err != nil {
			return fmt.Errorf("failed to scan subtitle: %w", err)
		}

		err = json.Unmarshal(contentJson, &subtitle.Content)
		if err != nil {
			return fmt.Errorf("failed to unmarshal content: %w", err)
		}
		subtitles = append(subtitles, subtitle)
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating subtitles: %w", err)
	}

	if len(subtitles) == 0 {
		return nil
	}

	// Collect unique texts for translation
	textsToTranslate := make(map[string]string)

	for _, subtitle := range subtitles {
		srcKey := subtitle.Content[sourceLanguage]
		if srcKey == "" {
			continue
		}

		if _, ok := textsToTranslate[srcKey]; !ok {
			textsToTranslate[srcKey] = ""
		}
	}

	// Process translations in parallel
	sourceLangFullText := movie.Languages[sourceLanguage]
	targetLangFullText := movie.Languages[targetLanguage]
	translations := translationService.processBatch(ctx, textsToTranslate, sourceLangFullText, targetLangFullText)

	// Update subtitles with translations
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	for _, subtitle := range subtitles {
		srcKey := subtitle.Content[sourceLanguage]
		if srcKey == "" {
			continue
		}
		translated := translations[srcKey]
		if translated == "" {
			continue
		}
		_, err = tx.Exec(`
			UPDATE subtitles 
			SET content = json_set(content, '$.' || ?, ?),
				updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, targetLanguage, translated, subtitle.ID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update subtitle: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s Subtitle) ExportSubtitle(movieId int, language string) (ExportResponse, error) {
	db := database.GetDB()
	if db == nil {
		return ExportResponse{}, errors.New("database connection is nil")
	}

	// Get movie details
	movie := NewMovie()
	movie, err := movie.GetMovieByID(movieId)
	if err != nil {
		return ExportResponse{}, fmt.Errorf("failed to get movie: %w", err)
	}

	// Get all subtitles for the movie
	rows, err := db.Query(`
		SELECT id, movie_id, sl_no, start_time, end_time, content, created_at, updated_at 
		FROM subtitles 
		WHERE movie_id = ? 
		ORDER BY sl_no ASC
	`, movieId)
	if err != nil {
		return ExportResponse{}, fmt.Errorf("failed to get subtitles: %w", err)
	}
	defer rows.Close()

	var subtitles []Subtitle
	for rows.Next() {
		var subtitle Subtitle
		var contentJson []byte
		err := rows.Scan(&subtitle.ID, &subtitle.MovieID, &subtitle.SlNo, &subtitle.StartTime, &subtitle.EndTime, &contentJson, &subtitle.CreatedAt, &subtitle.UpdatedAt)
		if err != nil {
			return ExportResponse{}, fmt.Errorf("failed to scan subtitle: %w", err)
		}

		err = json.Unmarshal(contentJson, &subtitle.Content)
		if err != nil {
			return ExportResponse{}, fmt.Errorf("failed to unmarshal content: %w", err)
		}
		subtitles = append(subtitles, subtitle)
	}

	if err = rows.Err(); err != nil {
		return ExportResponse{}, fmt.Errorf("error iterating subtitles: %w", err)
	}

	// Create subtitles directory if it doesn't exist
	subtitlesDir := "subtitles"
	if err := os.MkdirAll(subtitlesDir, 0755); err != nil {
		return ExportResponse{}, fmt.Errorf("failed to create subtitles directory: %w", err)
	}

	// Create movie directory
	movieDir := filepath.Join(subtitlesDir, movie.Title)
	if err := os.MkdirAll(movieDir, 0755); err != nil {
		return ExportResponse{}, fmt.Errorf("failed to create movie directory: %w", err)
	}

	// Create SRT file
	fileName := fmt.Sprintf("%s - %s.srt", movie.Title, movie.Languages[language])
	filePath := filepath.Join(movieDir, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return ExportResponse{}, fmt.Errorf("failed to create export file: %w", err)
	}
	defer file.Close()

	// Write subtitles in SRT format
	for _, subtitle := range subtitles {
		content := subtitle.Content[language]
		if content == "" {
			continue
		}

		// Write subtitle number
		fmt.Fprintf(file, "%d\n", subtitle.SlNo)

		// Write time
		fmt.Fprintf(file, "%s --> %s\n", subtitle.StartTime, subtitle.EndTime)

		// Write content
		fmt.Fprintf(file, "%s\n\n", content)
	}

	// Get absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return ExportResponse{}, fmt.Errorf("failed to get absolute path: %w", err)
	}

	return ExportResponse{
		FilePath: absPath,
	}, nil
}
