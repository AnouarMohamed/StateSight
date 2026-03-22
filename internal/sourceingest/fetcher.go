package sourceingest

import (
	"context"
	"fmt"
	"time"

	"github.com/AnouarMohamed/StateSight/pkg/model"
)

// DesiredState is the fetched desired representation before normalization.
type DesiredState struct {
	Revision string
	Summary  map[string]any
}

// Fetcher defines desired-state ingestion boundaries (Git providers later).
type Fetcher interface {
	FetchDesired(ctx context.Context, app model.Application) (DesiredState, error)
}

// MockFetcher is a deterministic placeholder for early development.
type MockFetcher struct{}

func (f MockFetcher) FetchDesired(_ context.Context, app model.Application) (DesiredState, error) {
	revision := fmt.Sprintf("mock-%s-%d", app.ID[:8], time.Now().UTC().Unix())
	return DesiredState{
		Revision: revision,
		Summary: map[string]any{
			"source":      "git",
			"repo":        "placeholder",
			"application": app.Name,
			"namespace":   app.Namespace,
			"resources":   12,
		},
	}, nil
}
