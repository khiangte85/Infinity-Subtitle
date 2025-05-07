package backend

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"infinity-subtitle/backend/database"
	"infinity-subtitle/backend/logger"
	"os"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type MovieQueue struct {
	ID              int               `json:"id"`
	MovieID         sql.NullInt64     `json:"movie_id"`
	Name            string            `json:"name"`
	Type            string            `json:"type"`
	FileType        string            `json:"file_type"`
	Content         string            `json:"content"`
	SourceLanguage  string            `json:"source_language"`
	TargetLanguages map[string]string `json:"target_languages"`
	Status          int               `json:"status"`
	CreatedAt       time.Time         `json:"created_at"`
	ProcessedAt     *time.Time        `json:"processed_at"`
}

type MovieQueueResponse struct {
	Movies     []MovieQueue `json:"movies"`
	Pagination Pagination   `json:"pagination"`
}

type AddToQueueRequest struct {
	Name            string   `json:"name"`
	Type            string   `json:"type"`
	FileType        string   `json:"file_type"`
	Content         string   `json:"content"`
	SourceLanguage  string   `json:"source_language"`
	TargetLanguages []string `json:"target_languages"`
}

const (
	MovieQueueStatusPending = iota
	MovieQueueStatusAudioTranscribed
	MovieQueueStatusMovieCreated
	MovieQueueStatusSubtitleCreated
	MovieQueueStatusSubtitleTranslated
	MovieQueueStatusFailed
)

func NewMovieQueue() *MovieQueue {
	return &MovieQueue{}
}

func (mq *MovieQueue) ListQueue(name string, pagination Pagination) (MovieQueueResponse, error) {
	db := database.GetDB()
	var response MovieQueueResponse
	var movies []MovieQueue

	query := "SELECT COUNT(*) FROM movies_queue"
	args := []any{}
	if name != "" {
		query += " WHERE name LIKE ?"
		args = append(args, "%"+name+"%")
	}

	row := db.QueryRow(query, args...)
	var total int
	err := row.Scan(&total)
	if err != nil {
		return response, fmt.Errorf("failed to get total count: %w", err)
	}
	pagination.RowsNumber = total

	offset := (pagination.Page - 1) * pagination.RowsPerPage

	query = "SELECT id, movie_id, name, type, file_type, source_language, target_languages, status," +
		"created_at, processed_at FROM movies_queue"
	if name != "" {
		query += " WHERE name LIKE ?"
		args = append(args, "%"+name+"%")
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
	query += " LIMIT ? OFFSET ?"
	args = append(args, pagination.RowsPerPage, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return response, fmt.Errorf("failed to get movies: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var movie MovieQueue
		var processedAt sql.NullTime
		var targetLanguagesJSON []byte
		err := rows.Scan(
			&movie.ID,
			&movie.MovieID,
			&movie.Name,
			&movie.Type,
			&movie.FileType,
			&movie.SourceLanguage,
			&targetLanguagesJSON,
			&movie.Status,
			&movie.CreatedAt,
			&processedAt,
		)
		if err != nil {
			return response, fmt.Errorf("failed to scan movie: %w", err)
		}
		if processedAt.Valid {
			movie.ProcessedAt = &processedAt.Time
		}
		err = json.Unmarshal(targetLanguagesJSON, &movie.TargetLanguages)
		if err != nil {
			return response, fmt.Errorf("failed to unmarshal target languages: %w", err)
		}

		movies = append(movies, movie)
	}

	if err = rows.Err(); err != nil {
		return response, fmt.Errorf("failed to get movies: %w", err)
	}

	response = MovieQueueResponse{
		Movies:     movies,
		Pagination: pagination,
	}

	return response, nil
}

func (mq *MovieQueue) AddToQueue(req []AddToQueueRequest) error {
	db := database.GetDB()

	languages := NewLanguage()
	langs, err := languages.GetAllLanguages()
	if err != nil {
		return fmt.Errorf("failed to get languages: %w", err)
	}

	langMap := make(map[string]string)
	for _, lang := range langs {
		langMap[lang.Code] = lang.Name
	}

	stmt, err := db.Prepare(`
		INSERT INTO movies_queue (
			name, type, file_type, content, source_language, target_languages, status, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, r := range req {
		targetLanguages := make(map[string]string)
		var targetLanguagesJSON []byte

		for _, lang := range r.TargetLanguages {
			targetLanguages[lang] = langMap[lang]
		}

		targetLanguagesJSON, err = json.Marshal(targetLanguages)
		if err != nil {
			return fmt.Errorf("failed to marshal target languages: %w", err)
		}

		// Always set initial status to pending
		_, err = stmt.Exec(r.Name, r.Type, r.FileType, r.Content, r.SourceLanguage, targetLanguagesJSON, MovieQueueStatusPending)
		if err != nil {
			return fmt.Errorf("failed to add movie to queue: %w", err)
		}
	}

	return nil
}

func (mq *MovieQueue) DeleteFromQueue(id int) error {
	db := database.GetDB()

	_, err := db.Exec("DELETE FROM movies_queue WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete movie from queue: %w", err)
	}

	return nil
}

func CreateMovieFromQueue(ctx context.Context) error {
	logger, err := logger.GetLogger()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	select {
	case <-ctx.Done():
		logger.Info("Context cancelled, exiting CreateMovieFromQueue function")
		return nil
	default:
		db := database.GetDB()

		lang := NewLanguage()
		langs, err := lang.GetAllLanguages()
		if err != nil {
			return fmt.Errorf("failed to get languages: %w", err)
		}

		langMap := make(map[string]string)
		for _, lang := range langs {
			langMap[lang.Code] = lang.Name
		}

		m := NewMovie()

		rows, err := db.QueryContext(ctx, `
		SELECT id, name, type, file_type, content, source_language, target_languages, status
		FROM movies_queue 
		WHERE movie_id IS NULL
		AND (
			(type = 'subtitle' AND status = ?) OR
			(type = 'audio' AND status = ?)
		)
	`, MovieQueueStatusPending, MovieQueueStatusAudioTranscribed)
		if err != nil {
			return fmt.Errorf("failed to get movies from queue: %w", err)
		}
		defer rows.Close()

		var movies []MovieQueue

		for rows.Next() {
			var mq MovieQueue
			var targetLanguagesJSON []byte
			err := rows.Scan(&mq.ID, &mq.Name, &mq.Type, &mq.FileType, &mq.Content, &mq.SourceLanguage, &targetLanguagesJSON, &mq.Status)
			if err != nil {
				return fmt.Errorf("failed to scan movie from queue: %w", err)
			}
			err = json.Unmarshal(targetLanguagesJSON, &mq.TargetLanguages)
			if err != nil {
				return fmt.Errorf("failed to unmarshal target languages: %w", err)
			}
			movies = append(movies, mq)
		}

		if err = rows.Err(); err != nil {
			return fmt.Errorf("failed to get movies from queue: %w", err)
		}

		for _, mq := range movies {
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				return fmt.Errorf("failed to begin transaction: %w", err)
			}

			mq.TargetLanguages[mq.SourceLanguage] = langMap[mq.SourceLanguage]

			m, err := m.CreateMovie(mq.Name, mq.SourceLanguage, mq.TargetLanguages)
			if err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					logger.Error("failed to rollback transaction: %w", rollbackErr)
				}
				return fmt.Errorf("failed to create movie from queue id: %d: %w", mq.ID, err)
			}
			logger.Info("movie created from queue id: %d", mq.ID)

			_, err = tx.ExecContext(ctx,
				"UPDATE movies_queue SET movie_id = ?, status = ? WHERE id = ?",
				m.ID, MovieQueueStatusMovieCreated, mq.ID)
			if err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					logger.Error("failed to rollback transaction: %w", rollbackErr)
				}
				return fmt.Errorf("failed to update status of movie queue id: %d: %w", mq.ID, err)
			}

			if err = tx.Commit(); err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					logger.Error("failed to rollback transaction: %w", rollbackErr)
				}
				return fmt.Errorf("failed to commit transaction: %w", err)
			}

			logger.Info("movie queue id status created: %d", mq.ID)
			mq.Status = MovieQueueStatusMovieCreated
			runtime.EventsEmit(ctx, "movie-created", mq.ID, MovieQueueStatusMovieCreated)
		}
		return nil
	}
}

func CreateSubtitleFromQueue(ctx context.Context) error {
	logger, err := logger.GetLogger()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	select {
	case <-ctx.Done():
		logger.Info("Context cancelled, exiting CreateSubtitleFromQueue function")
		return nil
	default:
		db := database.GetDB()

		s := NewSubtitle()

		rows, err := db.QueryContext(ctx, `
		SELECT mq.id as mid, mq.movie_id, mq.name, mq.content, mq.source_language, mq.target_languages, mq.status, 
		  mq.created_at as mq_created_at, mq.processed_at as mq_processed_at,
		  m.id, m.title, m.default_language, m.languages, m.created_at, m.updated_at
		FROM movies_queue mq
		LEFT JOIN movies m ON mq.movie_id = m.id
		WHERE mq.movie_id IS NOT NULL
		AND mq.movie_id != 0
		AND mq.status = ?
	`, MovieQueueStatusMovieCreated)
		if err != nil {
			return fmt.Errorf("failed to get movies from queue: %w", err)
		}
		defer rows.Close()

		type MovieWithContent struct {
			Movie   Movie
			MQ      MovieQueue
			Content string
		}

		var moviesWithContent []MovieWithContent

		for rows.Next() {
			var mwc MovieWithContent
			var processedAt sql.NullTime
			var targetLanguagesJSON []byte
			var jsonLanguages []byte
			err := rows.Scan(&mwc.MQ.ID, &mwc.MQ.MovieID, &mwc.MQ.Name, &mwc.MQ.Content, &mwc.MQ.SourceLanguage,
				&targetLanguagesJSON, &mwc.MQ.Status, &mwc.MQ.CreatedAt, &processedAt,
				&mwc.Movie.ID, &mwc.Movie.Title, &mwc.Movie.DefaultLanguage, &jsonLanguages,
				&mwc.Movie.CreatedAt, &mwc.Movie.UpdatedAt)
			if err != nil {
				return fmt.Errorf("failed to scan movie from queue: %w", err)
			}
			err = json.Unmarshal(jsonLanguages, &mwc.Movie.Languages)
			if err != nil {
				return fmt.Errorf("failed to unmarshal languages: %w", err)
			}
			err = json.Unmarshal(targetLanguagesJSON, &mwc.MQ.TargetLanguages)
			if err != nil {
				return fmt.Errorf("failed to unmarshal target languages: %w", err)
			}
			if processedAt.Valid {
				mwc.MQ.ProcessedAt = &processedAt.Time
			}
			moviesWithContent = append(moviesWithContent, mwc)
		}

		if err = rows.Err(); err != nil {
			return fmt.Errorf("failed to get movies from queue: %w", err)
		}

		for _, mwc := range moviesWithContent {
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				return fmt.Errorf("failed to begin transaction: %w", err)
			}

			err = s.ImportFromSRTFile(mwc.Movie, mwc.MQ.Content)
			if err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					logger.Error("failed to rollback transaction: %w", rollbackErr)
				}
				return fmt.Errorf("failed to create subtitle from queue id: %d: %w", mwc.MQ.ID, err)
			}

			logger.Info("subtitle created from queue %d", mwc.MQ.ID)
			_, err = tx.ExecContext(ctx, "UPDATE movies_queue SET status = ? WHERE movie_id = ?",
				MovieQueueStatusSubtitleCreated, mwc.Movie.ID)
			if err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					logger.Error("failed to rollback transaction: %w", rollbackErr)
				}
				return fmt.Errorf("failed to update status of movie queue id: %d: %w", mwc.MQ.ID, err)
			}

			if err = tx.Commit(); err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					logger.Error("failed to rollback transaction: %w", rollbackErr)
				}
				return fmt.Errorf("failed to commit transaction: %w", err)
			}

			logger.Info("subtitle queue id status created: %d", mwc.MQ.ID)
			runtime.EventsEmit(ctx, "subtitle-created", mwc.MQ.ID, MovieQueueStatusSubtitleCreated)
		}
		return nil
	}
}

func TranslateSubtitleFromQueue(ctx context.Context) error {
	logger, err := logger.GetLogger()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	select {
	case <-ctx.Done():
		logger.Info("Context cancelled, exiting TranslateSubtitleFromQueue function")
		return nil
	default:
		db := database.GetDB()

		rows, err := db.QueryContext(ctx, `
		SELECT movie_id, source_language, target_languages
		FROM movies_queue
		WHERE
		 status = ?
		AND movie_id IS NOT NULL
		AND movie_id != 0
		`, MovieQueueStatusSubtitleCreated)
		if err != nil {
			return fmt.Errorf("failed to get movies from queue: %w", err)
		}
		defer rows.Close()

		var movies []Movie
		for rows.Next() {
			var movie Movie
			var targetLanguagesJSON []byte
			err := rows.Scan(&movie.ID, &movie.DefaultLanguage, &targetLanguagesJSON)
			if err != nil {
				return fmt.Errorf("failed to scan movie id: %w", err)
			}
			err = json.Unmarshal(targetLanguagesJSON, &movie.Languages)
			if err != nil {
				return fmt.Errorf("failed to unmarshal target languages: %w", err)
			}
			movies = append(movies, movie)
		}

		if err = rows.Err(); err != nil {
			return fmt.Errorf("failed to get movies from queue: %w", err)
		}

		s := NewSubtitle()

		for _, movie := range movies {
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				return fmt.Errorf("failed to begin transaction: %w", err)
			}

			for code := range movie.Languages {
				if code == movie.DefaultLanguage {
					continue
				}

				err = s.TranslateSubtitles(movie.ID, movie.DefaultLanguage, code)
				if err != nil {
					if rollbackErr := tx.Rollback(); rollbackErr != nil {
						logger.Error("failed to rollback transaction: %w", rollbackErr)
					}
					return fmt.Errorf("failed to translate subtitles: %w", err)
				}
			}

			_, err = tx.ExecContext(ctx, "UPDATE movies_queue SET status = ? WHERE movie_id = ?",
				MovieQueueStatusSubtitleTranslated, movie.ID)
			if err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					logger.Error("failed to rollback transaction: %w", rollbackErr)
				}
				return fmt.Errorf("failed to update status of movie queue id: %d: %w", movie.ID, err)
			}

			if err = tx.Commit(); err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					logger.Error("failed to rollback transaction: %w", rollbackErr)
				}
				return fmt.Errorf("failed to commit transaction: %w", err)
			}

			logger.Info("subtitle queue id status translated: %d", movie.ID)
			runtime.EventsEmit(ctx, "subtitle-translated", movie.ID, MovieQueueStatusSubtitleTranslated)
		}
		return nil
	}
}

func TranscribeAudioFromQueue(ctx context.Context) error {
	logger, err := logger.GetLogger()
	if err != nil {
		return fmt.Errorf("failed to get logger: %w", err)
	}

	select {
	case <-ctx.Done():
		logger.Info("Context cancelled, exiting TranscribeAudioFromQueue function")
		return nil
	default:
		db := database.GetDB()

		// Use a transaction for the entire query to prevent locks
		tx, err := db.BeginTx(ctx, &sql.TxOptions{
			Isolation: sql.LevelSerializable,
		})
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer func() {
			if err != nil {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					logger.Error("failed to rollback transaction: %w", rollbackErr)
				}
			}
		}()

		// First, try to get and lock a single row
		rows, err := tx.QueryContext(ctx, `
		SELECT id, name, file_type, content, source_language, target_languages, status
		FROM movies_queue 
		WHERE status = ? AND type = 'audio'
		LIMIT 1
		`, MovieQueueStatusPending)
		if err != nil {
			return fmt.Errorf("failed to get audio files from queue: %w", err)
		}
		defer rows.Close()

		var mq MovieQueue
		var targetLanguagesJSON []byte
		var found bool

		// Only process one file at a time
		if rows.Next() {
			found = true
			err := rows.Scan(&mq.ID, &mq.Name, &mq.FileType, &mq.Content, &mq.SourceLanguage, &targetLanguagesJSON, &mq.Status)
			if err != nil {
				return fmt.Errorf("failed to scan audio file from queue: %w", err)
			}

			err = json.Unmarshal(targetLanguagesJSON, &mq.TargetLanguages)
			if err != nil {
				return fmt.Errorf("failed to unmarshal target languages: %w", err)
			}
		}

		if err = rows.Err(); err != nil {
			return fmt.Errorf("error iterating audio files: %w", err)
		}

		if !found {
			// No files to process, commit the transaction and return
			if err = tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit transaction: %w", err)
			}
			return nil
		}

		// Decode base64 content
		audioData, err := base64.StdEncoding.DecodeString(mq.Content)
		if err != nil {
			return fmt.Errorf("failed to decode base64 audio content: %w", err)
		}

		// Create a temporary file for the audio
		tmpFile, err := os.CreateTemp("", "audio-*."+mq.FileType)
		if err != nil {
			return fmt.Errorf("failed to create temporary file: %w", err)
		}
		defer os.Remove(tmpFile.Name())

		// Write audio data to temporary file
		if _, err := tmpFile.Write(audioData); err != nil {
			return fmt.Errorf("failed to write audio data to temporary file: %w", err)
		}
		tmpFile.Close()

		// Call transcription service
		srtContent, err := transcribeAudio(tmpFile.Name(), mq.SourceLanguage)
		if err != nil {
			return fmt.Errorf("failed to transcribe audio: %w", err)
		}

		// Update queue status and content with transcribed text
		_, err = tx.ExecContext(ctx,
			"UPDATE movies_queue SET content = ?, status = ? WHERE id = ?",
			srtContent, MovieQueueStatusAudioTranscribed, mq.ID)
		if err != nil {
			return fmt.Errorf("failed to update audio transcription status: %w", err)
		}

		// Commit the transaction
		if err = tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		logger.Info("audio transcription completed for queue id: %d", mq.ID)
		runtime.EventsEmit(ctx, "audio-transcribed", mq.ID, MovieQueueStatusAudioTranscribed)
		return nil
	}
}
