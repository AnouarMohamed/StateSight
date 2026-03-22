package model

import "time"

// Application is a tracked workload that can be analyzed for drift.
type Application struct {
	ID                 string    `json:"id"`
	WorkspaceID        string    `json:"workspace_id"`
	ClusterID          string    `json:"cluster_id"`
	SourceDefinitionID string    `json:"source_definition_id"`
	Name               string    `json:"name"`
	Namespace          string    `json:"namespace"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
