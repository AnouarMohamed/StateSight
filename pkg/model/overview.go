package model

// Overview is the top-level dashboard summary model.
type Overview struct {
	WorkspaceCount   int64 `json:"workspace_count"`
	ApplicationCount int64 `json:"application_count"`
	IncidentCount    int64 `json:"incident_count"`
	OpenJobsCount    int64 `json:"open_jobs_count"`
}
