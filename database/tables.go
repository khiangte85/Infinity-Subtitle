package database

func CreateTables() error {
	db := GetDB()

	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS movies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		default_language TEXT NOT NULL,
		languages JSON NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)

	if err != nil {
		return err
	}

	_, err = db.Exec(`
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
		return err
	}

	// create index on subtitles table
	_, err = db.Exec(`
	CREATE INDEX IF NOT EXISTS idx_movie_id ON subtitles (movie_id)
	CREATE INDEX IF NOT EXISTS idx_sl_no ON subtitles (sl_no)
	`)

	if err != nil {
		return err
	}

	return nil
}

func CheckTablesExists() (bool, error) {
	db := GetDB()

	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name='movies')").Scan(&exists)
	if err != nil {
		return false, err
	}

	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name='subtitles')").Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}


	