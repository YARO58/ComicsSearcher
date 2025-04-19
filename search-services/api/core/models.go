package core

import (
	"net/http"
)

type UpdateStatus string

const (
	StatusUpdateUnknown UpdateStatus = "unknown"
	StatusUpdateIdle    UpdateStatus = "idle"
	StatusUpdateRunning UpdateStatus = "running"
)

type UpdateStats struct {
	WordsTotal    int
	WordsUnique   int
	ComicsFetched int
	ComicsTotal   int
}

type Comics struct {
	ID  int64
	URL string
}

type Middleware func(http.Handler) http.Handler

type RateLimitConfig struct {
	SearchRate        float64
	SearchConcurrency int
}

type Claims struct {
	Subject string `json:"sub"`
}
