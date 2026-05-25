package k8scollect

import (
	"context"
	"os"
	"path/filepath"
	"strings"
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

func TestKubectlAdapterRequestsManagedFieldsForEvidence(t *testing.T) {
	argsPath := filepath.Join(t.TempDir(), "kubectl-args")
	t.Setenv("KUBECTL_ARGS_PATH", argsPath)

	kubectlBinary := filepath.Join(t.TempDir(), "kubectl")
	const script = "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$KUBECTL_ARGS_PATH\"\nprintf '%s\\n' '{\"items\":[]}'\n"
	if err := os.WriteFile(kubectlBinary, []byte(script), 0o700); err != nil {
		t.Fatalf("write kubectl fixture: %v", err)
	}

	_, err := (KubectlAdapter{KubectlBinary: kubectlBinary}).Collect(
		context.Background(),
		model.Cluster{},
		model.Application{Namespace: "payments"},
	)
	if err != nil {
		t.Fatalf("collect live state: %v", err)
	}

	args, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatalf("read kubectl arguments: %v", err)
	}
	if !strings.Contains(string(args), "--show-managed-fields=true\n") {
		t.Fatalf("expected managed fields to be requested, got arguments:\n%s", args)
	}
}
