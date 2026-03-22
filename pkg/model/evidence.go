package model

import "time"

// EvidenceRecord describes attribution clues for an incident.
type EvidenceRecord struct {
	ID         string    `json:"id"`
	IncidentID string    `json:"incident_id"`
	Source     string    `json:"source"`
	Detail     string    `json:"detail"`
	Actor      string    `json:"actor"`
	Confidence float64   `json:"confidence"`
	Metadata   string    `json:"metadata"`
	CreatedAt  time.Time `json:"created_at"`
}
