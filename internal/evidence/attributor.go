package evidence

import (
	"context"
	"fmt"

	"github.com/AnouarMohamed/StateSight/internal/incidents"
	"github.com/AnouarMohamed/StateSight/pkg/model"
)

// Attribution links an incident candidate to probable provenance signals.
type Attribution struct {
	Source     string
	Detail     string
	Actor      string
	Confidence float64
	Metadata   map[string]any
}

// Attributor defines evidence generation boundaries.
type Attributor interface {
	BuildAttributions(ctx context.Context, app model.Application, candidate incidents.Candidate) ([]Attribution, error)
}

// MockAttributor emits realistic placeholder evidence.
type MockAttributor struct{}

func (a MockAttributor) BuildAttributions(_ context.Context, app model.Application, candidate incidents.Candidate) ([]Attribution, error) {
	return []Attribution{
		{
			Source:     "managedFields",
			Detail:     fmt.Sprintf("field manager changed %s", candidate.FieldPath),
			Actor:      "system:serviceaccount:ops:drift-watcher",
			Confidence: 0.77,
			Metadata: map[string]any{
				"fieldManager": "kubectl-rollout",
				"application":  app.Name,
			},
		},
		{
			Source:     "audit",
			Detail:     "manual change observed near deployment window",
			Actor:      "platform.oncall@example.com",
			Confidence: 0.64,
			Metadata: map[string]any{
				"window": "post-deploy",
			},
		},
	}, nil
}
