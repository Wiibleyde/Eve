package models

import "time"

// StatusResponse is returned by GET /api/v1/status.
type StatusResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// HealthResponse is returned by GET /api/v1/health.
//
// Uptime is the process uptime in seconds (fractional).
type HealthResponse struct {
	Health string  `json:"health"`
	Uptime float64 `json:"uptime"`
}

// InfoResponse is returned by GET /api/v1/info.
type InfoResponse struct {
	App     string `json:"app"`
	Version string `json:"version"`
	Author  string `json:"author"`
}

// PingResponse is returned by GET /api/v1/ping.
//
// Timestamp is a preformatted RFC3339 string rather than a time.Time so the
// wire format is pinned to RFC3339 (encoding/json emits RFC3339Nano for
// time.Time).
type PingResponse struct {
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}
