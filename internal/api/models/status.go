package models

import "time"

type StatusResponse struct {
	Status    string    `json:"status" example:"ok"`
	Timestamp time.Time `json:"timestamp" description:"Current server time, UTC."`
}

type HealthResponse struct {
	Health string  `json:"health" example:"good"`
	Uptime float64 `json:"uptime" description:"Process uptime in seconds." example:"3612.48"`
}

type InfoResponse struct {
	App     string `json:"app" example:"Eve - API"`
	Version string `json:"version" description:"Build version, stamped at build time via ldflags. Reports dev when unset." example:"dev"`
	Author  string `json:"author" example:"Wiibleyde"`
}

type PingResponse struct {
	Message   string `json:"message" example:"pong"`
	Timestamp string `json:"timestamp" format:"date-time" description:"Current server time, UTC, RFC 3339."`
}
