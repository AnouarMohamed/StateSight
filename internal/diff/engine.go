package diff

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/AnouarMohamed/StateSight/internal/normalize"
	"github.com/AnouarMohamed/StateSight/pkg/model"
)

// Finding is a field-level drift candidate before incident grouping.
type Finding struct {
	Title          string
	Category       string
	Severity       string
	Confidence     float64
	ResourceRef    string
	FieldPath      string
	DesiredValue   string
	LiveValue      string
	DifferenceType string
}

// Engine defines semantic diff boundaries.
type Engine interface {
	Compare(ctx context.Context, app model.Application, desired normalize.Snapshot, live normalize.Snapshot) ([]Finding, error)
}

// SemanticEngine compares normalized desired and live snapshots.
type SemanticEngine struct{}

func (e SemanticEngine) Compare(_ context.Context, _ model.Application, desired normalize.Snapshot, live normalize.Snapshot) ([]Finding, error) {
	findings := make([]Finding, 0)

	for key, d := range desired.Resources {
		l, exists := live.Resources[key]
		if !exists {
			findings = append(findings, Finding{
				Title:          "Resource missing from cluster",
				Category:       "existence",
				Severity:       "high",
				Confidence:     0.95,
				ResourceRef:    resourceRef(d),
				FieldPath:      "$resource",
				DesiredValue:   "present",
				LiveValue:      "missing",
				DifferenceType: "missing",
			})
			continue
		}
		findings = append(findings, compareResource(d, l)...)
	}

	for key, l := range live.Resources {
		if _, exists := desired.Resources[key]; exists {
			continue
		}
		findings = append(findings, Finding{
			Title:          "Unexpected live resource",
			Category:       "existence",
			Severity:       "medium",
			Confidence:     0.81,
			ResourceRef:    resourceRef(l),
			FieldPath:      "$resource",
			DesiredValue:   "missing",
			LiveValue:      "present",
			DifferenceType: "extra",
		})
	}

	sort.SliceStable(findings, func(i, j int) bool {
		return findings[i].ResourceRef < findings[j].ResourceRef
	})
	return findings, nil
}

func compareResource(desired, live normalize.Resource) []Finding {
	out := make([]Finding, 0)
	ref := resourceRef(desired)

	desiredReplicas, desiredHasReplicas := readNumber(desired.Spec, "replicas")
	liveReplicas, liveHasReplicas := readNumber(live.Spec, "replicas")
	if desiredHasReplicas && liveHasReplicas && desiredReplicas != liveReplicas {
		out = append(out, Finding{
			Title:          "Replica count drift",
			Category:       "workload",
			Severity:       "medium",
			Confidence:     0.89,
			ResourceRef:    ref,
			FieldPath:      "spec.replicas",
			DesiredValue:   fmt.Sprintf("%d", desiredReplicas),
			LiveValue:      fmt.Sprintf("%d", liveReplicas),
			DifferenceType: "modified",
		})
	}

	desiredImage, desiredHasImage := firstContainerImage(desired.Spec)
	liveImage, liveHasImage := firstContainerImage(live.Spec)
	if desiredHasImage && liveHasImage && desiredImage != liveImage {
		out = append(out, Finding{
			Title:          "Container image drift",
			Category:       "configuration",
			Severity:       "high",
			Confidence:     0.93,
			ResourceRef:    ref,
			FieldPath:      "spec.template.spec.containers[0].image",
			DesiredValue:   desiredImage,
			LiveValue:      liveImage,
			DifferenceType: "modified",
		})
	}

	keys := annotationKeys(desired.Annotations, live.Annotations)
	for _, key := range keys {
		dVal := fmt.Sprintf("%v", desired.Annotations[key])
		lVal := fmt.Sprintf("%v", live.Annotations[key])
		if dVal == lVal {
			continue
		}
		out = append(out, Finding{
			Title:          "Annotation drift",
			Category:       "metadata",
			Severity:       "low",
			Confidence:     0.73,
			ResourceRef:    ref,
			FieldPath:      "metadata.annotations." + key,
			DesiredValue:   dVal,
			LiveValue:      lVal,
			DifferenceType: "modified",
		})
	}

	return out
}

func readNumber(input map[string]any, key string) (int64, bool) {
	if input == nil {
		return 0, false
	}
	value, ok := input[key]
	if !ok {
		return 0, false
	}
	switch v := value.(type) {
	case int:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case float64:
		return int64(v), true
	default:
		return 0, false
	}
}

func firstContainerImage(spec map[string]any) (string, bool) {
	template, ok := spec["template"].(map[string]any)
	if !ok {
		return "", false
	}
	podSpecWrapper, ok := template["spec"].(map[string]any)
	if !ok {
		return "", false
	}
	containers, ok := podSpecWrapper["containers"].([]any)
	if !ok || len(containers) == 0 {
		return "", false
	}
	first, ok := containers[0].(map[string]any)
	if !ok {
		return "", false
	}
	image, ok := first["image"].(string)
	if !ok {
		return "", false
	}
	return strings.TrimSpace(image), true
}

func annotationKeys(desired, live map[string]any) []string {
	set := map[string]struct{}{}
	for key := range desired {
		set[key] = struct{}{}
	}
	for key := range live {
		set[key] = struct{}{}
	}

	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func resourceRef(resource normalize.Resource) string {
	apiVersion := strings.TrimSpace(resource.APIVersion)
	kind := strings.TrimSpace(resource.Kind)
	namespace := strings.TrimSpace(resource.Namespace)
	name := strings.TrimSpace(resource.Name)
	if apiVersion == "" {
		apiVersion = "v1"
	}
	return fmt.Sprintf("%s/%s:%s/%s", apiVersion, kind, namespace, name)
}
