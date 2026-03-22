package scoring

import (
	"context"

	"github.com/AnouarMohamed/StateSight/internal/incidents"
	"github.com/AnouarMohamed/StateSight/pkg/model"
)

// Recommendation computes incident-level action guidance.
type Recommendation interface {
	Recommend(ctx context.Context, candidate incidents.Candidate) (action string, confidence float64, err error)
}

// RuleBasedRecommendation is a baseline deterministic scorer.
type RuleBasedRecommendation struct{}

func (r RuleBasedRecommendation) Recommend(_ context.Context, candidate incidents.Candidate) (string, float64, error) {
	switch candidate.Severity {
	case "high":
		return model.IncidentRecommendedReconcile, 0.92, nil
	case "medium":
		return model.IncidentRecommendedInvestigate, 0.81, nil
	default:
		return model.IncidentRecommendedMonitor, 0.69, nil
	}
}
