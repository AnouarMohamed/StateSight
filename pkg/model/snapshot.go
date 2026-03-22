package model

import "time"

// DesiredSnapshot represents the desired configuration from source control.
type DesiredSnapshot struct {
	ID            string    `json:"id"`
	ApplicationID string    `json:"application_id"`
	Revision      string    `json:"revision"`
	SummaryJSON   string    `json:"summary_json"`
	CapturedAt    time.Time `json:"captured_at"`
}

// LiveSnapshot represents observed state from a cluster.
type LiveSnapshot struct {
	ID            string    `json:"id"`
	ApplicationID string    `json:"application_id"`
	SummaryJSON   string    `json:"summary_json"`
	ObservedAt    time.Time `json:"observed_at"`
}
