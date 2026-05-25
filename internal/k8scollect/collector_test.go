package k8scollect

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/AnouarMohamed/StateSight/pkg/model"
)

func TestCollectorFailsWithoutSyntheticFallback(t *testing.T) {
	collector := NewCollector(CollectorOptions{
		KubectlBinary: filepath.Join(t.TempDir(), "missing-kubectl"),
	})

	_, err := collector.CollectLiveState(context.Background(), model.Cluster{}, model.Application{
		Name:      "ledger-api",
		Namespace: "payments",
	})
	if err == nil {
		t.Fatal("expected collection failure when kubectl cannot run")
	}
}

func TestCollectorUsesSyntheticFallbackOnlyWhenEnabled(t *testing.T) {
	collector := NewCollector(CollectorOptions{
		KubectlBinary:          filepath.Join(t.TempDir(), "missing-kubectl"),
		AllowSyntheticFallback: true,
	})

	state, err := collector.CollectLiveState(context.Background(), model.Cluster{Name: "demo"}, model.Application{
		Name:      "ledger-api",
		Namespace: "payments",
	})
	if err != nil {
		t.Fatalf("collect synthetic live state: %v", err)
	}

	if got := state.Summary["source"]; got != "synthetic" {
		t.Fatalf("expected synthetic source, got %v", got)
	}
	if len(state.Resources) != 1 {
		t.Fatalf("expected one synthetic resource, got %d", len(state.Resources))
	}
}
