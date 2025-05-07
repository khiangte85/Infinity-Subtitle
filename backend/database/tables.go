package database

import (
	"database/sql"
	"fmt"
	"infinity-subtitle/backend/logger"
	"log"
)

func createLanguagesTable(db *sql.DB) error {

	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS languages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		code TEXT NOT NULL UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)

	if err != nil {
		return fmt.Errorf("error creating languages table: %w", err)
	}

	_, err = db.Exec(`
	INSERT INTO languages (name, code) VALUES 
	('English', 'en'),
	('Chinese', 'zh'),
	('Japanese', 'ja'),
	('Korean', 'ko'),
	('Thai', 'th'),
	('Indonesian', 'id'),
	('Malay', 'ms'),
	('Vietnamese', 'vi'),
	('Hindi', 'hi')
	`)

	if err != nil {
		log.Println("[x] Error inserting languages:", err)
	}

	return nil
}

func createMoviesTable(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS movies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		default_language	 TEXT NOT NULL,
		languages JSON NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (default_language) REFERENCES languages(code)
	)`)

	if err != nil {
		return fmt.Errorf("error creating movies table: %w", err)
	}
	return nil
}

func createSubtitlesTable(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS subtitles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		movie_id INTEGER NOT NULL,
		sl_no INTEGER NOT NULL,
		start_time TEXT NOT NULL,
		end_time TEXT NOT NULL,
		content JSON NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (movie_id) REFERENCES movies(id)
	)`)

	if err != nil {
		return fmt.Errorf("error creating subtitles table: %w", err)
	}

	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_movie_id ON subtitles(movie_id)")
	if err != nil {
		return fmt.Errorf("error creating movie_id index: %w", err)
	}

	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_sl_no ON subtitles(sl_no)")
	if err != nil {
		return fmt.Errorf("error creating sl_no index: %w", err)
	}

	return nil
}

func createMoviesQueueTable(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS movies_queue (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		movie_id INTEGER,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		file_type TEXT,
		content LONGTEXT NOT NULL,
		source_language TEXT NOT NULL,
		target_languages JSON NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		processed_at DATETIME DEFAULT NULL,
		status SMALLINT NOT NULL DEFAULT 0
	)`)

	if err != nil {
		return fmt.Errorf("error creating movies_queues table: %w", err)
	}

	return nil
}

func CheckTablesExists() error {
	db := GetDB()
	logger, err := logger.GetLogger()
	if err != nil {
		logger.Error("Error getting logger:", err)
		return err
	}

	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name='languages')").Scan(&exists)
	if err != nil {
		logger.Error("Error checking languages table:", err)
		return err
	}

	if !exists {
		err = createLanguagesTable(db.DB)
		if err != nil {
			logger.Error("Error creating languages table:", err)
			return err
		}
	}

	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name='movies')").Scan(&exists)
	if err != nil {
		logger.Error("Error checking movies table:", err)
		return err
	}

	if !exists {
		err = createMoviesTable(db.DB)
		if err != nil {
			logger.Error("Error creating movies table:", err)
			return err
		}
	}

	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name='subtitles')").Scan(&exists)
	if err != nil {
		logger.Error("Error checking subtitles table:", err)
		return err
	}

	if !exists {
		err = createSubtitlesTable(db.DB)
		if err != nil {
			logger.Error("Error creating subtitles table:", err)
			return err
		}
	}

	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name='movies_queue')").Scan(&exists)
	if err != nil {
		logger.Error("Error checking movies_queue table:", err)
		return err
	}

	if !exists {
		err = createMoviesQueueTable(db.DB)
		if err != nil {
			logger.Error("Error creating movies_queue table:", err)
			return err
		}
	}

	return nil
}
