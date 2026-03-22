package diff

import (
	"context"
	"fmt"
	"hash/fnv"

	"github.com/AnouarMohamed/StateSight/internal/normalize"
	"github.com/AnouarMohamed/StateSight/pkg/model"
)

// Finding is a field-level drift candidate before incident grouping.
type Finding struct {
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

// Engine defines semantic diff boundaries.
type Engine interface {
	Compare(ctx context.Context, app model.Application, desired normalize.Snapshot, live normalize.Snapshot) ([]Finding, error)
}

// SeededEngine emits deterministic mock findings and keeps shape close to real diff output.
type SeededEngine struct{}

func (e SeededEngine) Compare(_ context.Context, app model.Application, _ normalize.Snapshot, _ normalize.Snapshot) ([]Finding, error) {
	h := fnv.New32a()
	if _, err := h.Write([]byte(app.ID)); err != nil {
		return nil, fmt.Errorf("hash app id: %w", err)
	}
	count := int(h.Sum32()%3) + 1 // 1..3

	candidates := []Finding{
		{
			Title:          "Replica count drift",
			Category:       "workload",
			Severity:       "medium",
			Confidence:     0.82,
			ResourceRef:    fmt.Sprintf("apps/v1/Deployment:%s/%s", app.Namespace, app.Name),
			FieldPath:      "spec.replicas",
			DesiredValue:   "3",
			LiveValue:      "2",
			DifferenceType: "modified",
		},
		{
			Title:          "Container image drift",
			Category:       "configuration",
			Severity:       "high",
			Confidence:     0.91,
			ResourceRef:    fmt.Sprintf("apps/v1/Deployment:%s/%s", app.Namespace, app.Name),
			FieldPath:      "spec.template.spec.containers[0].image",
			DesiredValue:   "registry.example.com/ledger-api:v1.24.0",
			LiveValue:      "registry.example.com/ledger-api:v1.24.0-hotfix",
			DifferenceType: "modified",
		},
		{
			Title:          "Ingress annotation drift",
			Category:       "networking",
			Severity:       "low",
			Confidence:     0.68,
			ResourceRef:    fmt.Sprintf("networking.k8s.io/v1/Ingress:%s/%s", app.Namespace, app.Name),
			FieldPath:      "metadata.annotations.nginx.ingress.kubernetes.io/proxy-body-size",
			DesiredValue:   "8m",
			LiveValue:      "32m",
			DifferenceType: "modified",
		},
	}

	return candidates[:count], nil
}
