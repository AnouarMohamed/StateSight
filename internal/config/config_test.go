package config

import "testing"

func TestLoadWorkerReadsCollectorConfiguration(t *testing.T) {
	t.Setenv("GIT_BIN", "/tools/git")
	t.Setenv("GIT_CACHE_DIR", "/var/cache/statesight")
	t.Setenv("KUBECTL_BIN", "/tools/kubectl")
	t.Setenv("ALLOW_SYNTHETIC_LIVE_STATE", "true")

	cfg, err := LoadWorker()
	if err != nil {
		t.Fatalf("load worker config: %v", err)
	}

	if cfg.GitBinary != "/tools/git" {
		t.Fatalf("expected configured git binary, got %q", cfg.GitBinary)
	}
	if cfg.GitCacheDir != "/var/cache/statesight" {
		t.Fatalf("expected configured git cache directory, got %q", cfg.GitCacheDir)
	}
	if cfg.KubectlBinary != "/tools/kubectl" {
		t.Fatalf("expected configured kubectl binary, got %q", cfg.KubectlBinary)
	}
	if !cfg.AllowSyntheticLiveState {
		t.Fatal("expected synthetic live state to be enabled by explicit configuration")
	}
}

func TestLoadWorkerDisablesSyntheticLiveStateByDefault(t *testing.T) {
	t.Setenv("ALLOW_SYNTHETIC_LIVE_STATE", "")

	cfg, err := LoadWorker()
	if err != nil {
		t.Fatalf("load worker config: %v", err)
	}

	if cfg.AllowSyntheticLiveState {
		t.Fatal("expected synthetic live state to be disabled by default")
	}
}
