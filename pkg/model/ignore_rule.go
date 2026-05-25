package model

import "time"

// IgnoreRule is a user-defined suppression rule for noisy drift.
type IgnoreRule struct {
	ID              string    `json:"id"`
	WorkspaceID     string    `json:"workspace_id"`
	ApplicationID   string    `json:"application_id"`
	ResourceRef     string    `json:"resource_ref"`
	Name            string    `json:"name"`
	MatchExpression string    `json:"match_expression"`
	Reason          string    `json:"reason"`
	CreatedBy       string    `json:"created_by"`
	Active          bool      `json:"active"`
	CreatedAt       time.Time `json:"created_at"`
}
