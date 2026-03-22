package normalize

import (
	"fmt"
	"strings"
)

type Resource struct {
	Key         string         `json:"key"`
	APIVersion  string         `json:"api_version"`
	Kind        string         `json:"kind"`
	Namespace   string         `json:"namespace"`
	Name        string         `json:"name"`
	Spec        map[string]any `json:"spec"`
	Annotations map[string]any `json:"annotations"`
	Raw         map[string]any `json:"raw"`
}

type Snapshot struct {
	Resources map[string]Resource `json:"resources"`
}

// Normalizer converts raw manifests/cluster objects into comparable resources.
type Normalizer interface {
	Normalize(resources []map[string]any) Snapshot
}

// PassThroughNormalizer performs lightweight semantic extraction.
type PassThroughNormalizer struct{}

func (n PassThroughNormalizer) Normalize(resources []map[string]any) Snapshot {
	index := make(map[string]Resource, len(resources))
	for _, raw := range resources {
		resource := extractResource(raw)
		if resource.Key == "" {
			continue
		}
		index[resource.Key] = resource
	}
	return Snapshot{Resources: index}
}

func extractResource(raw map[string]any) Resource {
	apiVersion := readString(raw, "apiVersion")
	kind := readString(raw, "kind")
	metadata := readMap(raw, "metadata")
	namespace := readString(metadata, "namespace")
	name := readString(metadata, "name")
	spec := readMap(raw, "spec")
	annotations := readMap(readMap(raw, "metadata"), "annotations")

	key := resourceKey(kind, namespace, name)
	return Resource{
		Key:         key,
		APIVersion:  apiVersion,
		Kind:        kind,
		Namespace:   namespace,
		Name:        name,
		Spec:        spec,
		Annotations: annotations,
		Raw:         raw,
	}
}

func resourceKey(kind, namespace, name string) string {
	kind = strings.TrimSpace(kind)
	namespace = strings.TrimSpace(namespace)
	name = strings.TrimSpace(name)
	if kind == "" || name == "" {
		return ""
	}
	return fmt.Sprintf("%s|%s|%s", kind, namespace, name)
}

func readMap(input map[string]any, key string) map[string]any {
	if input == nil {
		return map[string]any{}
	}
	value, ok := input[key]
	if !ok {
		return map[string]any{}
	}
	typed, ok := value.(map[string]any)
	if !ok {
		return map[string]any{}
	}
	return typed
}

func readString(input map[string]any, key string) string {
	if input == nil {
		return ""
	}
	value, ok := input[key]
	if !ok {
		return ""
	}
	typed, ok := value.(string)
	if !ok {
		return ""
	}
	return typed
}
