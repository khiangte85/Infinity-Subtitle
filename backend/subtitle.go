package backend

import "time"

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
