package jobs

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/AnouarMohamed/StateSight/internal/sourceingest"
	"github.com/AnouarMohamed/StateSight/pkg/model"
)

func TestNewProcessorAppliesAdapterOptions(t *testing.T) {
	kubectlBinary := filepath.Join(t.TempDir(), "kubectl")
	const kubectlScript = "#!/bin/sh\nprintf '%s\\n' '{\"items\":[]}'\n"
	if err := os.WriteFile(kubectlBinary, []byte(kubectlScript), 0o700); err != nil {
		t.Fatalf("write kubectl test binary: %v", err)
	}

	processor := NewProcessor(nil, slog.New(slog.NewTextHandler(io.Discard, nil)), ProcessorOptions{
		GitBinary:     "/tools/git",
		GitCacheDir:   "/var/cache/statesight",
		KubectlBinary: kubectlBinary,
	})

	fetcher, ok := processor.fetcher.(sourceingest.GitFetcher)
	if !ok {
		t.Fatalf("expected GitFetcher, got %T", processor.fetcher)
	}
	if fetcher.GitBinary != "/tools/git" || fetcher.CacheDir != "/var/cache/statesight" {
		t.Fatalf("expected configured git fetcher, got %#v", fetcher)
	}

	state, err := processor.collector.CollectLiveState(context.Background(), model.Cluster{}, model.Application{
		Namespace: "payments",
	})
	if err != nil {
		t.Fatalf("collect state through configured kubectl binary: %v", err)
	}
	if got := state.Summary["source"]; got != "kubectl" {
		t.Fatalf("expected kubectl source, got %v", got)
	}
}
