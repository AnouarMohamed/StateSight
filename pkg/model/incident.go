package model

import "time"

const (
	IncidentRecommendedIgnore      = "ignore"
	IncidentRecommendedMonitor     = "monitor"
	IncidentRecommendedInvestigate = "investigate"
	IncidentRecommendedReconcile   = "reconcile"
)

// DriftIncident groups one or more field-level drift observations.
type DriftIncident struct {
	ID                string    `json:"id"`
	ApplicationID     string    `json:"application_id"`
	DesiredSnapshotID string    `json:"desired_snapshot_id"`
	LiveSnapshotID    string    `json:"live_snapshot_id"`
	Title             string    `json:"title"`
	Category          string    `json:"category"`
	Severity          string    `json:"severity"`
	Confidence        float64   `json:"confidence"`
	RecommendedAction string    `json:"recommended_action"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// DriftField captures a specific field-level difference.
type DriftField struct {
	ID             string    `json:"id"`
	IncidentID     string    `json:"incident_id"`
	ResourceRef    string    `json:"resource_ref"`
	FieldPath      string    `json:"field_path"`
	DesiredValue   string    `json:"desired_value"`
	LiveValue      string    `json:"live_value"`
	DifferenceType string    `json:"difference_type"`
	CreatedAt      time.Time `json:"created_at"`
}
