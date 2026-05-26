package model

// GenerateTimings reports the wall-clock cost of the two phases of a /generate call.
// All values are in milliseconds.
type GenerateTimings struct {
	GenerationMs float64 `json:"generation_ms"`
	InsertionMs  float64 `json:"insertion_ms"`
	TotalMs      float64 `json:"total_ms"`
}

// GenerateResponse is the unified envelope returned by every /api/sites/generate/* endpoint.
type GenerateResponse struct {
	Success    bool            `json:"success"`
	Kind       string          `json:"kind"`
	SiteID     int64           `json:"site_id"`
	Count      int             `json:"count"`
	Length     int             `json:"length,omitempty"`
	Concurrent bool            `json:"concurrent"`
	Workers    int             `json:"workers,omitempty"`
	Timings    GenerateTimings `json:"timings"`
}
