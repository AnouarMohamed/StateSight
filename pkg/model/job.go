package model

import "time"

const (
	JobStatusQueued     = "queued"
	JobStatusProcessing = "processing"
	JobStatusCompleted  = "completed"
	JobStatusFailed     = "failed"
)

const (
	JobTypeAnalyzeApplication = "analyze_application"
	JobTypeIngestGitHubEvent  = "ingest_github_event"
)

// Job tracks asynchronous work requested by API calls or event ingestion.
type Job struct {
	ID            string     `json:"id"`
	JobType       string     `json:"job_type"`
	Status        string     `json:"status"`
	ApplicationID *string    `json:"application_id,omitempty"`
	Payload       string     `json:"payload"`
	ErrorMessage  *string    `json:"error_message,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
}
