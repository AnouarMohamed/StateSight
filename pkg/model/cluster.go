package model

import "time"

// Cluster identifies a Kubernetes cluster tracked by the platform.
type Cluster struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	Name        string    `json:"name"`
	Provider    string    `json:"provider"`
	CreatedAt   time.Time `json:"created_at"`
}
