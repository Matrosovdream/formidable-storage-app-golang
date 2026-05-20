package model

import "time"

type SiteResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SiteStatsDTO struct {
	Fields       int64            `json:"fields"`
	Emails       int64            `json:"emails"`
	EntryHistory int64            `json:"entry_history"`
	Queue        map[string]int64 `json:"queue"`
}

type SiteWithStatsResponse struct {
	SiteResponse
	Token string       `json:"token,omitempty"`
	Stats SiteStatsDTO `json:"stats"`
}

type CreateSiteRequest struct {
	Name string `json:"name" validate:"required,max=255"`
	URL  string `json:"url"  validate:"required,url,max=255"`
}
