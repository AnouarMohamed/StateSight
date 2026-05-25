package sourceingest

import "testing"

func TestNewGitFetcherDefaultsEmptySettings(t *testing.T) {
	fetcher := NewGitFetcher("", "")

	if fetcher.GitBinary != "git" {
		t.Fatalf("expected default git binary, got %q", fetcher.GitBinary)
	}
	if fetcher.CacheDir != ".statesight/git-cache" {
		t.Fatalf("expected default cache directory, got %q", fetcher.CacheDir)
	}
}
