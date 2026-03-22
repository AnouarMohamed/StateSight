package incidents

import (
	"context"

	"github.com/AnouarMohamed/StateSight/internal/diff"
)

// Candidate is a grouped incident candidate ready for persistence.
type Candidate struct {
	Title          string
	Category       string
	Severity       string
	Confidence     float64
	ResourceRef    string
	FieldPath      string
	DesiredValue   string
	LiveValue      string
	DifferenceType string
}

// Grouper groups field-level drift findings into incident-level units.
type Grouper interface {
	Group(ctx context.Context, findings []diff.Finding) ([]Candidate, error)
}

// SimpleGrouper keeps one finding per incident for baseline simplicity.
type SimpleGrouper struct{}

func (g SimpleGrouper) Group(_ context.Context, findings []diff.Finding) ([]Candidate, error) {
	out := make([]Candidate, 0, len(findings))
	for _, finding := range findings {
		out = append(out, Candidate{
			Title:          finding.Title,
			Category:       finding.Category,
			Severity:       finding.Severity,
			Confidence:     finding.Confidence,
			ResourceRef:    finding.ResourceRef,
			FieldPath:      finding.FieldPath,
			DesiredValue:   finding.DesiredValue,
			LiveValue:      finding.LiveValue,
			DifferenceType: finding.DifferenceType,
		})
	}
	return out, nil
}
