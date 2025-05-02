package backend

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"infinity-subtitle/backend/database"
	"infinity-subtitle/backend/logger"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type MovieQueue struct {
	ID              int               `json:"id"`
	MovieID         sql.NullInt64     `json:"movie_id"`
	Name            string            `json:"name"`
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
	Content         string   `json:"content"`
	SourceLanguage  string   `json:"source_language"`
	TargetLanguages []string `json:"target_languages"`
}

const (
	MovieQueueStatusPending = iota
	MovieQueueStatusMovieCreated
	MovieQueueStatusSubtitleCreated
	MovieQueueStatusTranslating
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

	query = "SELECT id, movie_id, name, source_language, target_languages, status, created_at, processed_at FROM movies_queue"
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
			name, content, source_language, target_languages, status, created_at
		) VALUES (?, ?, ?, ?, 0, CURRENT_TIMESTAMP)
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
		_, err = stmt.Exec(r.Name, r.Content, r.SourceLanguage, targetLanguagesJSON)
		if err != nil {
			return fmt.Errorf("failed to add movie to queue: %w", err)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to add movies to queue: %w", err)
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

func CreateMovieFromQueue(ctx context.Context) {
	logger, err := logger.GetLogger()
	if err != nil {
		logger.Error("failed to get logger: %w", err)
	}

	select {
	case <-ctx.Done():
		logger.Info("Context cancelled, exiting CreateMovieFromQueue")
		return
	default:
		db := database.GetDB()

		lang := NewLanguage()
		langs, err := lang.GetAllLanguages()
		if err != nil {
			logger.Error("failed to get languages: %w", err)
		}

		langMap := make(map[string]string)
		for _, lang := range langs {
			langMap[lang.Code] = lang.Name
		}

		m := NewMovie()

		rows, err := db.QueryContext(ctx, `
		SELECT id, name, content, source_language, target_languages, status
		FROM movies_queue 
		WHERE movie_id IS NULL
		AND status = ?
	`, MovieQueueStatusPending)
		if err != nil {
			logger.Error("failed to get movies from queue: %w", err)
		}
		defer rows.Close()

		var movies []MovieQueue

		for rows.Next() {
			var mq MovieQueue
			var targetLanguagesJSON []byte
			err := rows.Scan(&mq.ID, &mq.Name, &mq.Content, &mq.SourceLanguage, &targetLanguagesJSON, &mq.Status)
			if err != nil {
				logger.Error("failed to scan movie from queue: %w", err)
			}
			err = json.Unmarshal(targetLanguagesJSON, &mq.TargetLanguages)
			if err != nil {
				logger.Error("failed to unmarshal target languages: %w", err)
				continue
			}
			movies = append(movies, mq)
		}

		if err = rows.Err(); err != nil {
			logger.Error("failed to get movies from queue: %w", err)
		}

		for _, mq := range movies {

			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				logger.Error("failed to begin transaction: %w", err)
				continue
			}

			mq.TargetLanguages[mq.SourceLanguage] = langMap[mq.SourceLanguage]

			m, err := m.CreateMovie(mq.Name, mq.SourceLanguage, mq.TargetLanguages)
			if err != nil {
				logger.Error("failed to create movie from queue id: %d: %w", mq.ID, err)
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					logger.Error("failed to rollback transaction: %w", rollbackErr)
				}
				continue
			}
			logger.Info("movie created from queue id: %d", mq.ID)

			_, err = tx.ExecContext(ctx,
				"UPDATE movies_queue SET movie_id = ?, status = ? WHERE id = ?",
				m.ID, MovieQueueStatusMovieCreated, mq.ID)
			if err != nil {
				logger.Error("failed to update status of movie queue id: %d: %w", mq.ID, err)
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					logger.Error("failed to rollback transaction: %w", rollbackErr)
				}
				continue
			}

			if err = tx.Commit(); err != nil {
				logger.Error("failed to commit transaction: %w", err)
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					logger.Error("failed to rollback transaction: %w", rollbackErr)
				}
				continue
			}

			logger.Info("movie queue id status created: %d", mq.ID)
			mq.Status = MovieQueueStatusMovieCreated
			runtime.EventsEmit(ctx, "movie-created", mq)
		}
	}
}

func CreateSubtitleFromQueue(ctx context.Context) {
	logger, err := logger.GetLogger()
	if err != nil {
		logger.Error("failed to get logger: %w", err)
	}

	select {
	case <-ctx.Done():
		logger.Info("Context cancelled, exiting CreateSubtitleFromQueue")
		return
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
			logger.Error("failed to get movies from queue: %w", err)
			return
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
				logger.Error("failed to scan movie from queue: %w", err)
				continue
			}
			err = json.Unmarshal(jsonLanguages, &mwc.Movie.Languages)
			if err != nil {
				logger.Error("failed to unmarshal languages: %w", err)
				continue
			}
			err = json.Unmarshal(targetLanguagesJSON, &mwc.MQ.TargetLanguages)
			if err != nil {
				logger.Error("failed to unmarshal target languages: %w", err)
				continue
			}
			if processedAt.Valid {
				mwc.MQ.ProcessedAt = &processedAt.Time
			}
			moviesWithContent = append(moviesWithContent, mwc)
		}

		if err = rows.Err(); err != nil {
			logger.Error("failed to get movies from queue: %w", err)
			return
		}

		for _, mwc := range moviesWithContent {
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				logger.Error("failed to begin transaction: %w", err)
				continue
			}

			err = s.ImportFromSRTFile(mwc.Movie, mwc.MQ.Content)
			if err != nil {
				logger.Error("failed to create subtitle from queue id: %d: %w", mwc.MQ.ID, err)
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					logger.Error("failed to rollback transaction: %w", rollbackErr)
				}
				continue
			}

			logger.Info("subtitle created from queue %d", mwc.MQ.ID)
			_, err = tx.ExecContext(ctx, "UPDATE movies_queue SET status = ? WHERE movie_id = ?",
				MovieQueueStatusSubtitleCreated, mwc.Movie.ID)
			if err != nil {
				logger.Error("failed to update status of movie queue id: %d: %w", mwc.MQ.ID, err)
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					logger.Error("failed to rollback transaction: %w", rollbackErr)
				}
				continue
			}

			if err = tx.Commit(); err != nil {
				logger.Error("failed to commit transaction: %w", err)
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					logger.Error("failed to rollback transaction: %w", rollbackErr)
				}
				continue
			}

			logger.Info("subtitle queue id status created: %d", mwc.MQ.ID)
			mwc.MQ.Status = MovieQueueStatusSubtitleCreated
			runtime.EventsEmit(ctx, "subtitle-created", mwc.MQ)
		}
	}

}

func TranslateSubtitleFromQueue(ctx context.Context) {
	logger, err := logger.GetLogger()
	if err != nil {
		logger.Error("failed to get logger: %w", err)
	}

	select {
	case <-ctx.Done():
		logger.Info("Context cancelled, exiting TranslateSubtitleFromQueue")
		return
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
			logger.Error("failed to get movies from queue: %w", err)
			return
		}
		defer rows.Close()

		var movies []Movie
		for rows.Next() {
			var movie Movie
			var targetLanguagesJSON []byte
			err := rows.Scan(&movie.ID, &movie.DefaultLanguage, &targetLanguagesJSON)
			if err != nil {
				logger.Error("failed to scan movie id: %w", err)
				continue
			}
			err = json.Unmarshal(targetLanguagesJSON, &movie.Languages)
			if err != nil {
				logger.Error("failed to unmarshal target languages: %w", err)
				continue
			}
			movies = append(movies, movie)
		}

		if err = rows.Err(); err != nil {
			logger.Error("failed to get movies from queue: %w", err)
			return
		}

		s := NewSubtitle()

		for _, movie := range movies {
			_, err = db.ExecContext(ctx, "UPDATE movies_queue SET status = ? WHERE movie_id = ?",
				MovieQueueStatusTranslating, movie.ID)
			if err != nil {
				logger.Error("failed to update status of movie queue id: %d: %w", movie.ID, err)
				continue
			}
			logger.Info("subtitle translating: %d", movie.ID)
			runtime.EventsEmit(ctx, "subtitle-translating", movie.ID, MovieQueueStatusTranslating)

			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				logger.Error("failed to begin transaction: %w", err)
				continue
			}

			for code := range movie.Languages {
				if code == movie.DefaultLanguage {
					continue
				}

				err = s.TranslateSubtitles(movie.ID, movie.DefaultLanguage, code)
				if err != nil {
					logger.Error("failed to translate subtitles: %w", err)
					continue
				}
			}

			_, err = tx.ExecContext(ctx, "UPDATE movies_queue SET status = ? WHERE movie_id = ?",
				MovieQueueStatusSubtitleTranslated, movie.ID)
			if err != nil {
				logger.Error("failed to update status of movie queue id: %d: %w", movie.ID, err)
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					logger.Error("failed to rollback transaction: %w", rollbackErr)
				}
			}

			if err = tx.Commit(); err != nil {
				logger.Error("failed to commit transaction: %w", err)
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					logger.Error("failed to rollback transaction: %w", rollbackErr)
				}
				continue
			}

			logger.Info("subtitle queue id status translated: %d", movie.ID)
			runtime.EventsEmit(ctx, "subtitle-translated", movie.ID, MovieQueueStatusSubtitleTranslated)
		}

	}
}
