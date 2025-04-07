package backend

type Pagination struct {
	SortBy       string `json:"sortBy"`
	Descending   bool   `json:"descending"`
	Page         int    `json:"page"`
	RowsPerPage  int    `json:"rowsPerPage"`
	RowsNumber   int    `json:"rowsNumber"`
}

func NewPagination() *Pagination {
	return &Pagination{
		SortBy:       "created_at",
		Descending:   true,
		Page:         1,
		RowsPerPage:  10,
		RowsNumber:   0,
	}
}