package database

import (
	"context"
	"fmt"
	"log"
)

func CreateTables() {
	db := GetDB()

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		log.Fatal("[x] Error beginning transaction:", err)
	}

	defer tx.Rollback()

	_, err = tx.Exec(`
	CREATE TABLE IF NOT EXISTS languages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		code TEXT NOT NULL UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)

	if err != nil {
		log.Fatal("[x] Error creating languages table:", err)
	}

	_, err = tx.Exec(`
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
		log.Fatal("[x] Error creating movies table:", err)
	}

	_, err = tx.Exec(`
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
		log.Fatal("[x] Error creating subtitles table:", err)
	}

	// Create indexes separately
	_, err = tx.Exec("CREATE INDEX IF NOT EXISTS idx_movie_id ON subtitles(movie_id)")
	if err != nil {
		log.Fatal("[x] Error creating movie_id index:", err)
	}

	_, err = tx.Exec("CREATE INDEX IF NOT EXISTS idx_sl_no ON subtitles(sl_no)")
	if err != nil {
		log.Fatal("[x] Error creating sl_no index:", err)
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal("[x] Error committing transaction:", err)
	}
}

func InsertLanguages() {
	db := GetDB()

	_, err := db.Exec(`
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
}

func CheckTablesExists() bool {
	db := GetDB()

	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name='languages')").Scan(&exists)
	if err != nil || !exists {
		fmt.Println("[x] Error checking languages table:", err)
		return false
	}

	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name='movies')").Scan(&exists)
	if err != nil || !exists {
		fmt.Println("[x] Error checking movies table:", err)
		return false
	}

	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name='subtitles')").Scan(&exists)
	if err != nil || !exists {
		fmt.Println("[x] Error checking subtitles table:", err)
		return false
	}

	if exists {
		fmt.Println("[+] Tables exist")
		return exists
	}

	return false
}
