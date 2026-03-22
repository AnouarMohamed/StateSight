package model

import "time"

// Workspace is an organizational boundary for clusters, sources, and applications.
type Workspace struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
