package backend

import (
	"encoding/json"
	"errors"
	"fmt"
	"infinity-subtitle/backend/database"
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

func NewSubtitle() *Subtitle {
	return &Subtitle{}
}

func (s *Subtitle) GetSubtitlesByMovieID(movieID int) ([]Subtitle, error) {
	db := database.GetDB()
	var subtitles []Subtitle

	rows, err := db.Query("SELECT * FROM subtitles WHERE movie_id = ?", movieID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var subtitle Subtitle
		err := rows.Scan(&subtitle.ID, &subtitle.MovieID, &subtitle.SlNo, &subtitle.StartTime, &subtitle.EndTime, &subtitle.Content, &subtitle.CreatedAt, &subtitle.UpdatedAt)
		if err != nil {
			return nil, err
		}
		subtitles = append(subtitles, subtitle)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return subtitles, nil
}

func (s *Subtitle) UpdateSubtitle(subtitle Subtitle) error {
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
