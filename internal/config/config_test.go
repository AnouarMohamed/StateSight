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

func TestLoadAPIRequiresOIDCConfigurationWhenAuthenticationEnabled(t *testing.T) {
	t.Setenv("AUTH_REQUIRED", "true")
	t.Setenv("OIDC_ISSUER_URL", "")
	t.Setenv("OIDC_AUDIENCE", "")

	if _, err := LoadAPI(); err == nil {
		t.Fatal("expected missing OIDC configuration to fail")
	}
}

func TestLoadAPIRejectsInsecureOIDCIssuerByDefault(t *testing.T) {
	t.Setenv("AUTH_REQUIRED", "true")
	t.Setenv("OIDC_ISSUER_URL", "http://identity.example.test")
	t.Setenv("OIDC_AUDIENCE", "statesight-api")

	if _, err := LoadAPI(); err == nil {
		t.Fatal("expected insecure OIDC issuer to fail without an explicit local-development override")
	}
}

func TestLoadAPIAcceptsConfiguredOIDCBoundary(t *testing.T) {
	t.Setenv("AUTH_REQUIRED", "true")
	t.Setenv("OIDC_ISSUER_URL", "https://identity.example.test")
	t.Setenv("OIDC_AUDIENCE", "statesight-api")

	cfg, err := LoadAPI()
	if err != nil {
		t.Fatalf("load api config: %v", err)
	}
	if !cfg.AuthRequired || cfg.OIDCIssuerURL != "https://identity.example.test" || cfg.OIDCAudience != "statesight-api" {
		t.Fatalf("unexpected OIDC configuration: %#v", cfg)
	}
}
