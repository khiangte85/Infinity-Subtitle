package database

import (
	"database/sql"
	"fmt"
	"time"

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

	fmt.Println("[+] Connected to database")

	db.DB.SetMaxOpenConns(10)
	db.DB.SetMaxIdleConns(10)
	db.DB.SetConnMaxLifetime(time.Minute * 5)
	
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

 