package diff

import (
	"context"
	"reflect"
	"testing"

	"github.com/AnouarMohamed/StateSight/internal/normalize"
	"github.com/AnouarMohamed/StateSight/pkg/model"
)

func TestSemanticEngineDetectsLabelAndServiceSelectorDrift(t *testing.T) {
	desired := normalizedSnapshot(serviceResource(
		map[string]any{
			"app.kubernetes.io/name": "ledger",
			"environment":            "production",
		},
		map[string]any{"delivery.example.com/team": "payments"},
		map[string]any{
			"app.kubernetes.io/name": "ledger",
			"track":                  "stable",
		},
	))
	live := normalizedSnapshot(serviceResource(
		map[string]any{
			"app.kubernetes.io/name": "ledger",
			"environment":            "staging",
			"observed":               "manual",
		},
		map[string]any{},
		map[string]any{
			"app.kubernetes.io/name": "ledger",
			"track":                  "canary",
		},
	))

	findings, err := (SemanticEngine{}).Compare(context.Background(), model.Application{}, desired, live)
	if err != nil {
		t.Fatalf("compare snapshots: %v", err)
	}

	expected := []Finding{
		{FieldPath: "metadata.annotations.delivery.example.com/team", DesiredValue: "payments", LiveValue: "<absent>", DifferenceType: "removed"},
		{FieldPath: "metadata.labels.environment", DesiredValue: "production", LiveValue: "staging", DifferenceType: "modified"},
		{FieldPath: "metadata.labels.observed", DesiredValue: "<absent>", LiveValue: "manual", DifferenceType: "added"},
		{FieldPath: "spec.selector.track", DesiredValue: "stable", LiveValue: "canary", DifferenceType: "modified"},
	}
	if len(findings) != len(expected) {
		t.Fatalf("expected %d findings, got %#v", len(expected), findings)
	}

	for index, want := range expected {
		got := findings[index]
		if got.FieldPath != want.FieldPath || got.DesiredValue != want.DesiredValue || got.LiveValue != want.LiveValue || got.DifferenceType != want.DifferenceType {
			t.Fatalf("finding %d = %#v, want field/value/type %#v", index, got, want)
		}
	}
	if findings[3].Category != "routing" || findings[3].Severity != "high" {
		t.Fatalf("expected selector drift to be high-severity routing evidence, got %#v", findings[3])
	}
}

func TestSemanticEngineOrdersFindingsDeterministically(t *testing.T) {
	desired := normalizedSnapshot(serviceResource(
		map[string]any{"z": "desired", "a": "desired"},
		map[string]any{"z": "desired", "a": "desired"},
		map[string]any{"z": "desired", "a": "desired"},
	))
	live := normalizedSnapshot(serviceResource(
		map[string]any{"z": "live", "a": "live"},
		map[string]any{"z": "live", "a": "live"},
		map[string]any{"z": "live", "a": "live"},
	))

	baseline, err := (SemanticEngine{}).Compare(context.Background(), model.Application{}, desired, live)
	if err != nil {
		t.Fatalf("compare baseline: %v", err)
	}
	wantPaths := []string{
		"metadata.annotations.a",
		"metadata.annotations.z",
		"metadata.labels.a",
		"metadata.labels.z",
		"spec.selector.a",
		"spec.selector.z",
	}
	gotPaths := make([]string, 0, len(baseline))
	for _, finding := range baseline {
		gotPaths = append(gotPaths, finding.FieldPath)
	}
	if !reflect.DeepEqual(gotPaths, wantPaths) {
		t.Fatalf("ordered paths = %#v, want %#v", gotPaths, wantPaths)
	}

	for attempt := 0; attempt < 25; attempt++ {
		got, compareErr := (SemanticEngine{}).Compare(context.Background(), model.Application{}, desired, live)
		if compareErr != nil {
			t.Fatalf("compare attempt %d: %v", attempt, compareErr)
		}
		if !reflect.DeepEqual(got, baseline) {
			t.Fatalf("attempt %d returned nondeterministic findings: %#v versus %#v", attempt, got, baseline)
		}
	}
}

func TestSemanticEngineDetectsNamedContainerConfigurationDrift(t *testing.T) {
	desired := normalizedSnapshot(workloadResource([]any{
		map[string]any{
			"name":  "api",
			"image": "registry.example.com/api:v1",
			"env": []any{
				map[string]any{"name": "MODE", "value": "production"},
				map[string]any{"name": "TOKEN", "valueFrom": map[string]any{"secretKeyRef": map[string]any{"name": "api-token", "key": "token"}}},
			},
			"resources": map[string]any{
				"requests": map[string]any{"cpu": "250m"},
				"limits":   map[string]any{"memory": "256Mi"},
			},
		},
		map[string]any{"name": "sidecar", "image": "registry.example.com/sidecar:v1"},
	}))
	live := normalizedSnapshot(workloadResource([]any{
		map[string]any{
			"name":  "api",
			"image": "registry.example.com/api:v1",
			"env": []any{
				map[string]any{"name": "MODE", "value": "staging"},
				map[string]any{"name": "TOKEN", "valueFrom": map[string]any{"configMapKeyRef": map[string]any{"name": "api-config", "key": "token"}}},
			},
			"resources": map[string]any{
				"requests": map[string]any{"cpu": "500m"},
			},
		},
		map[string]any{"name": "debug", "image": "registry.example.com/debug:v1"},
	}))

	findings, err := (SemanticEngine{}).Compare(context.Background(), model.Application{}, desired, live)
	if err != nil {
		t.Fatalf("compare snapshots: %v", err)
	}

	expected := []Finding{
		{
			FieldPath:      "spec.template.spec.containers[name=api].env[name=MODE]",
			DesiredValue:   `{"value":"production"}`,
			LiveValue:      `{"value":"staging"}`,
			DifferenceType: "modified",
		},
		{
			FieldPath:      "spec.template.spec.containers[name=api].env[name=TOKEN]",
			DesiredValue:   `{"valueFrom":{"secretKeyRef":{"key":"token","name":"api-token"}}}`,
			LiveValue:      `{"valueFrom":{"configMapKeyRef":{"key":"token","name":"api-config"}}}`,
			DifferenceType: "modified",
		},
		{
			FieldPath:      "spec.template.spec.containers[name=api].resources.limits.memory",
			DesiredValue:   "256Mi",
			LiveValue:      "<absent>",
			DifferenceType: "removed",
		},
		{
			FieldPath:      "spec.template.spec.containers[name=api].resources.requests.cpu",
			DesiredValue:   "250m",
			LiveValue:      "500m",
			DifferenceType: "modified",
		},
		{
			FieldPath:      "spec.template.spec.containers[name=debug]",
			DesiredValue:   "<absent>",
			LiveValue:      "present",
			DifferenceType: "added",
		},
		{
			FieldPath:      "spec.template.spec.containers[name=sidecar]",
			DesiredValue:   "present",
			LiveValue:      "<absent>",
			DifferenceType: "removed",
		},
	}
	if len(findings) != len(expected) {
		t.Fatalf("expected %d findings, got %#v", len(expected), findings)
	}
	for index, want := range expected {
		got := findings[index]
		if got.FieldPath != want.FieldPath || got.DesiredValue != want.DesiredValue || got.LiveValue != want.LiveValue || got.DifferenceType != want.DifferenceType {
			t.Fatalf("finding %d = %#v, want field/value/type %#v", index, got, want)
		}
	}
}

func TestSemanticEngineTreatsEquivalentResourceQuantitiesAsEqual(t *testing.T) {
	desired := normalizedSnapshot(workloadResource([]any{map[string]any{
		"name":  "api",
		"image": "registry.example.com/api:v1",
		"resources": map[string]any{
			"requests": map[string]any{"cpu": "1"},
			"limits":   map[string]any{"memory": "1Gi"},
		},
	}}))
	live := normalizedSnapshot(workloadResource([]any{map[string]any{
		"name":  "api",
		"image": "registry.example.com/api:v1",
		"resources": map[string]any{
			"requests": map[string]any{"cpu": "1000m"},
			"limits":   map[string]any{"memory": "1024Mi"},
		},
	}}))

	findings, err := (SemanticEngine{}).Compare(context.Background(), model.Application{}, desired, live)
	if err != nil {
		t.Fatalf("compare snapshots: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected equivalent Kubernetes quantities not to drift, got %#v", findings)
	}
}

func normalizedSnapshot(resources ...map[string]any) normalize.Snapshot {
	return (normalize.PassThroughNormalizer{}).Normalize(resources)
}

func serviceResource(labels, annotations, selector map[string]any) map[string]any {
	return map[string]any{
		"apiVersion": "v1",
		"kind":       "Service",
		"metadata": map[string]any{
			"name":        "ledger-api",
			"namespace":   "payments",
			"labels":      labels,
			"annotations": annotations,
		},
		"spec": map[string]any{
			"selector": selector,
		},
	}
}

func workloadResource(containers []any) map[string]any {
	return map[string]any{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]any{
			"name":      "ledger-api",
			"namespace": "payments",
		},
		"spec": map[string]any{
			"template": map[string]any{
				"spec": map[string]any{"containers": containers},
			},
		},
	}
}
