package backend

import (
	"errors"
	"infinity-subtitle/backend/database"
	"time"
)

type Language struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
}

func NewLanguage() *Language {
	return &Language{}
}

func (l Language) CreateLanguage(name string, code string) error {
	if name == "" || code == "" {
		return errors.New("name and code are required")
	}

	db := database.GetDB()
	_, err := db.Exec("INSERT INTO languages (name, code) VALUES (?, ?)", name, code)
	if err != nil {
		return err
	}
	return nil
}

func (l Language) UpdateLanguage(id int, name string) error {
	db := database.GetDB()
	_, err := db.Exec("UPDATE languages SET name = ? WHERE id = ?", name, id)
	if err != nil {
		return err
	}
	return nil
}

func (l Language) GetLanguageByID(id int) (*Language, error) {
	db := database.GetDB()
	row := db.QueryRow("SELECT * FROM languages WHERE id = ?", id)
	var language Language
	err := row.Scan(&language.ID, &language.Name, &language.Code, &language.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &language, nil
}

func (l Language) GetAllLanguages() ([]Language, error) {
	db := database.GetDB()
	rows, err := db.Query("SELECT * FROM languages")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	languages := []Language{}
	for rows.Next() {
		var language Language
		err := rows.Scan(&language.ID, &language.Name, &language.Code, &language.CreatedAt)
		if err != nil {
			return nil, err
		}
		languages = append(languages, language)
	}

	return languages, nil
}
