package entity

import "time"

const (
	SystemStatusOperational = "operational"
	SystemStatusDegraded    = "degraded"
	SystemStatusDown        = "down"
	SystemStatusDisabled    = "disabled"
)

type EngineStatus struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Status   string `json:"status"`
}

type DependencyStatus struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	LatencyMS int64  `json:"latencyMs"`
	Detail    string `json:"detail,omitempty"`
}

// SystemStatusResponse is TITIK INTEGRASI API #4.
type SystemStatusResponse struct {
	Status        string             `json:"status"`
	Version       string             `json:"version"`
	Environment   string             `json:"environment"`
	UptimeSeconds int64              `json:"uptimeSeconds"`
	Engine        EngineStatus       `json:"engine"`
	Dependencies  []DependencyStatus `json:"dependencies"`
	CheckedAt     time.Time          `json:"checkedAt"`
}
