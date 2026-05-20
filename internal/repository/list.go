package repository

// ListParams is the common pagination/filter input for all list-style repo methods.
type ListParams struct {
	Filters map[string]any
	SortBy  string
	SortDir string
	Page    int
	PerPage int
}

// ListResult wraps paginated results.
type ListResult[T any] struct {
	Data     []T
	Total    int64
	Page     int
	PerPage  int
	LastPage int
}

func (p *ListParams) Normalize(defaultPerPage, maxPerPage int) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage <= 0 {
		p.PerPage = defaultPerPage
	}
	if p.PerPage > maxPerPage {
		p.PerPage = maxPerPage
	}
	if p.SortDir != "asc" && p.SortDir != "desc" {
		p.SortDir = "desc"
	}
}

func (p *ListParams) Offset() int { return (p.Page - 1) * p.PerPage }
