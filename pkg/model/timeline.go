package model

import "time"

// TimelineEvent is a chronological event associated with incident analysis context.
type TimelineEvent struct {
	At      time.Time `json:"at"`
	Type    string    `json:"type"`
	Summary string    `json:"summary"`
}
