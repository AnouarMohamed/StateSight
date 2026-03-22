package model

import "time"

// SourceDefinition captures where desired state is sourced from.
type SourceDefinition struct {
	ID            string    `json:"id"`
	WorkspaceID   string    `json:"workspace_id"`
	Name          string    `json:"name"`
	RepoURL       string    `json:"repo_url"`
	DefaultBranch string    `json:"default_branch"`
	Path          string    `json:"path"`
	CreatedAt     time.Time `json:"created_at"`
}
