package models

import "time"

type StatusResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

type HealthResponse struct {
	Health string  `json:"health"`
	Uptime float64 `json:"uptime"`
}

type InfoResponse struct {
	App     string `json:"app"`
	Version string `json:"version"`
	Author  string `json:"author"`
}

type PingResponse struct {
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}
