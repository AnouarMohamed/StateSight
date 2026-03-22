package model

// IncidentDetails combines incident metadata with field-level and evidence details.
type IncidentDetails struct {
	Incident DriftIncident    `json:"incident"`
	Fields   []DriftField     `json:"fields"`
	Evidence []EvidenceRecord `json:"evidence"`
	Timeline []TimelineEvent  `json:"timeline"`
}
