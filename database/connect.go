package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	*sql.DB
}

var defaultDb = &Database{}

func (db *Database) Connect() (error) {
	var err error
	db.DB, err = sql.Open("sqlite3", "./database.db")
	if err != nil {
		return err
	}

	return nil
}

func GetDB() *Database {
	return defaultDb
}

func (db *Database) Init() error {
	return db.Connect()
}

func (db *Database) Close() error {
	return db.DB.Close()
}

 