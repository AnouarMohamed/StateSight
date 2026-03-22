package model

import "time"

// GitHubEvent stores minimal webhook metadata for traceability.
type GitHubEvent struct {
	ID         string    `json:"id"`
	EventType  string    `json:"event_type"`
	DeliveryID string    `json:"delivery_id"`
	Action     string    `json:"action"`
	Repository string    `json:"repository"`
	Payload    string    `json:"payload"`
	ReceivedAt time.Time `json:"received_at"`
}
