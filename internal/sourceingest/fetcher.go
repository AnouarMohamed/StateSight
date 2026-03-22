package sourceingest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/AnouarMohamed/StateSight/pkg/model"
)

// DesiredState is the desired representation fetched from Git source definitions.
type DesiredState struct {
	Revision  string
	Resources []map[string]any
	Summary   map[string]any
}

// Fetcher defines desired-state ingestion boundaries.
type Fetcher interface {
	FetchDesired(ctx context.Context, app model.Application, source model.SourceDefinition) (DesiredState, error)
}

// GitFetcher fetches desired manifests from source definition repositories.
type GitFetcher struct {
	GitBinary string
	CacheDir  string
}

func NewGitFetcher(gitBinary, cacheDir string) GitFetcher {
	return GitFetcher{
		GitBinary: gitBinary,
		CacheDir:  cacheDir,
	}
}

func (f GitFetcher) FetchDesired(ctx context.Context, app model.Application, source model.SourceDefinition) (DesiredState, error) {
	if strings.TrimSpace(source.RepoURL) == "" {
		return DesiredState{}, fmt.Errorf("source definition %s has empty repo_url", source.ID)
	}

	checkoutDir := filepath.Join(f.CacheDir, fmt.Sprintf("%s-%d", app.ID, time.Now().UTC().UnixNano()))
	if err := os.MkdirAll(checkoutDir, 0o755); err != nil {
		return DesiredState{}, fmt.Errorf("create checkout dir: %w", err)
	}
	defer os.RemoveAll(checkoutDir)

	branch := source.DefaultBranch
	if strings.TrimSpace(branch) == "" {
		branch = "main"
	}

	cloneArgs := []string{"clone", "--depth", "1", "--branch", branch, source.RepoURL, checkoutDir}
	if output, err := runCommand(ctx, f.GitBinary, cloneArgs...); err != nil {
		return DesiredState{}, fmt.Errorf("git clone failed: %w: %s", err, output)
	}

	revOutput, err := runCommand(ctx, f.GitBinary, "-C", checkoutDir, "rev-parse", "HEAD")
	if err != nil {
		return DesiredState{}, fmt.Errorf("resolve git revision: %w: %s", err, revOutput)
	}
	revision := strings.TrimSpace(revOutput)

	root := checkoutDir
	if strings.TrimSpace(source.Path) != "" && strings.TrimSpace(source.Path) != "." {
		root = filepath.Join(checkoutDir, source.Path)
	}

	resources, err := loadManifestResources(root)
	if err != nil {
		return DesiredState{}, err
	}

	return DesiredState{
		Revision:  revision,
		Resources: resources,
		Summary: map[string]any{
			"source":         "git",
			"application":    app.Name,
			"repository":     source.RepoURL,
			"path":           source.Path,
			"resource_count": len(resources),
		},
	}, nil
}

func runCommand(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return out.String() + stderr.String(), err
	}
	return out.String(), nil
}

func loadManifestResources(root string) ([]map[string]any, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("stat source path %s: %w", root, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("source path %s is not a directory", root)
	}

	var resources []map[string]any
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".yaml" && ext != ".yml" && ext != ".json" {
			return nil
		}

		fileResources, err := parseManifestFile(path)
		if err != nil {
			return fmt.Errorf("parse manifest %s: %w", path, err)
		}
		resources = append(resources, fileResources...)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk source manifests: %w", err)
	}

	return resources, nil
}

func parseManifestFile(path string) ([]map[string]any, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".json" {
		var doc map[string]any
		if err := json.Unmarshal(raw, &doc); err != nil {
			return nil, err
		}
		if len(doc) == 0 {
			return nil, nil
		}
		return []map[string]any{doc}, nil
	}

	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	var docs []map[string]any
	for {
		var node any
		if err := decoder.Decode(&node); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if node == nil {
			continue
		}

		jsonBytes, err := json.Marshal(node)
		if err != nil {
			return nil, err
		}
		var out map[string]any
		if err := json.Unmarshal(jsonBytes, &out); err != nil {
			return nil, err
		}
		if len(out) == 0 {
			continue
		}
		docs = append(docs, out)
	}
	return docs, nil
}
