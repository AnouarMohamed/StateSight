package timelines

import (
	"context"
	"time"
)

// Event is a placeholder timeline event record.
type Event struct {
	At      time.Time `json:"at"`
	Type    string    `json:"type"`
	Summary string    `json:"summary"`
}

// Builder defines timeline construction boundaries.
type Builder interface {
	Build(ctx context.Context, incidentID string) ([]Event, error)
}

// PlaceholderBuilder returns a minimal TODO timeline.
type PlaceholderBuilder struct{}

func (PlaceholderBuilder) Build(_ context.Context, _ string) ([]Event, error) {
	return []Event{
		{
			At:      time.Now().UTC(),
			Type:    "todo",
			Summary: "Timeline construction will include git commits, cluster events, and reconciliations.",
		},
	}, nil
}
