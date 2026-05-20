package model

import "fmt"

// Paginated mirrors Laravel's default paginator JSON.
type Paginated[T any] struct {
	CurrentPage  int     `json:"current_page"`
	Data         []T     `json:"data"`
	FirstPageURL string  `json:"first_page_url"`
	From         int     `json:"from"`
	LastPage     int     `json:"last_page"`
	LastPageURL  string  `json:"last_page_url"`
	NextPageURL  *string `json:"next_page_url"`
	Path         string  `json:"path"`
	PerPage      int     `json:"per_page"`
	PrevPageURL  *string `json:"prev_page_url"`
	To           int     `json:"to"`
	Total        int64   `json:"total"`
}

// BuildPaginated assembles the Laravel-shaped paginator envelope.
//   path     - canonical URL of the listing endpoint (no query string)
//   data     - the page slice
//   total    - total count across all pages
//   page     - current page number (1-based)
//   perPage  - rows per page
//   lastPage - total pages
func BuildPaginated[T any](path string, data []T, total int64, page, perPage, lastPage int) Paginated[T] {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 1
	}
	if lastPage < 1 {
		lastPage = 1
	}

	from := (page-1)*perPage + 1
	to := from + len(data) - 1
	if len(data) == 0 {
		from, to = 0, 0
	}

	first := fmt.Sprintf("%s?page=1", path)
	last := fmt.Sprintf("%s?page=%d", path, lastPage)

	var prev, next *string
	if page > 1 {
		s := fmt.Sprintf("%s?page=%d", path, page-1)
		prev = &s
	}
	if page < lastPage {
		s := fmt.Sprintf("%s?page=%d", path, page+1)
		next = &s
	}

	return Paginated[T]{
		CurrentPage:  page,
		Data:         data,
		FirstPageURL: first,
		From:         from,
		LastPage:     lastPage,
		LastPageURL:  last,
		NextPageURL:  next,
		Path:         path,
		PerPage:      perPage,
		PrevPageURL:  prev,
		To:           to,
		Total:        total,
	}
}
