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
