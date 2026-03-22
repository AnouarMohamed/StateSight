package k8scollect

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/AnouarMohamed/StateSight/pkg/model"
)

// LiveState captures a real cluster snapshot.
type LiveState struct {
	Resources []map[string]any
	Summary   map[string]any
}

// Collector defines live-state collection boundaries.
type Collector interface {
	CollectLiveState(ctx context.Context, cluster model.Cluster, app model.Application) (LiveState, error)
}

type CollectorOptions struct {
	KubectlBinary string
}

type Adapter interface {
	Collect(ctx context.Context, cluster model.Cluster, app model.Application) ([]map[string]any, error)
	Name() string
}

type collector struct {
	primary  Adapter
	fallback Adapter
}

func NewCollector(options CollectorOptions) Collector {
	binary := strings.TrimSpace(options.KubectlBinary)
	if binary == "" {
		binary = "kubectl"
	}

	return collector{
		primary:  KubectlAdapter{KubectlBinary: binary},
		fallback: SyntheticAdapter{},
	}
}

func (c collector) CollectLiveState(ctx context.Context, cluster model.Cluster, app model.Application) (LiveState, error) {
	resources, err := c.primary.Collect(ctx, cluster, app)
	if err == nil {
		return LiveState{
			Resources: resources,
			Summary: map[string]any{
				"source":         c.primary.Name(),
				"resource_count": len(resources),
				"namespace":      app.Namespace,
				"cluster":        cluster.Name,
			},
		}, nil
	}

	if c.fallback != nil {
		fallbackResources, fallbackErr := c.fallback.Collect(ctx, cluster, app)
		if fallbackErr == nil {
			return LiveState{
				Resources: fallbackResources,
				Summary: map[string]any{
					"source":          c.fallback.Name(),
					"resource_count":  len(fallbackResources),
					"namespace":       app.Namespace,
					"cluster":         cluster.Name,
					"fallback_reason": err.Error(),
				},
			}, nil
		}
	}

	return LiveState{}, fmt.Errorf("collect live state via %s: %w", c.primary.Name(), err)
}

// KubectlAdapter fetches resources using kubectl and kube context configuration.
type KubectlAdapter struct {
	KubectlBinary string
}

func (a KubectlAdapter) Name() string {
	return "kubectl"
}

func (a KubectlAdapter) Collect(ctx context.Context, cluster model.Cluster, app model.Application) ([]map[string]any, error) {
	args := []string{}
	if strings.TrimSpace(cluster.KubeconfigPath) != "" {
		args = append(args, "--kubeconfig", cluster.KubeconfigPath)
	}
	if strings.TrimSpace(cluster.KubeContext) != "" {
		args = append(args, "--context", cluster.KubeContext)
	}

	args = append(args,
		"get",
		"deployments,statefulsets,daemonsets,services,ingresses,configmaps",
		"-n", app.Namespace,
		"-o", "json",
	)

	cmd := exec.CommandContext(ctx, a.KubectlBinary, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("kubectl get resources failed: %w (%s)", err, strings.TrimSpace(stderr.String()))
	}

	var payload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		return nil, fmt.Errorf("decode kubectl output: %w", err)
	}
	return payload.Items, nil
}

// SyntheticAdapter is a fallback for local environments without cluster access.
type SyntheticAdapter struct{}

func (SyntheticAdapter) Name() string {
	return "synthetic"
}

func (SyntheticAdapter) Collect(_ context.Context, _ model.Cluster, app model.Application) ([]map[string]any, error) {
	return []map[string]any{
		{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]any{
				"name":      app.Name,
				"namespace": app.Namespace,
			},
			"spec": map[string]any{
				"replicas": float64(2),
				"template": map[string]any{
					"spec": map[string]any{
						"containers": []any{
							map[string]any{
								"name":  app.Name,
								"image": "registry.example.com/" + app.Name + ":live",
							},
						},
					},
				},
			},
		},
	}, nil
}
