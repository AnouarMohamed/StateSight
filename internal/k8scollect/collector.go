package k8scollect

import (
	"context"

	"github.com/AnouarMohamed/StateSight/pkg/model"
)

// LiveState is cluster-observed representation before normalization.
type LiveState struct {
	Summary map[string]any
}

// Collector defines live-state collection boundaries.
type Collector interface {
	CollectLiveState(ctx context.Context, app model.Application) (LiveState, error)
}

// MockCollector is a placeholder until client-go collection is added.
type MockCollector struct{}

func (c MockCollector) CollectLiveState(_ context.Context, app model.Application) (LiveState, error) {
	return LiveState{
		Summary: map[string]any{
			"source":      "cluster",
			"cluster":     "placeholder",
			"application": app.Name,
			"namespace":   app.Namespace,
			"resources":   12,
		},
	}, nil
}
