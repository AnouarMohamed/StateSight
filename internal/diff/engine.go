package diff

import (
	"context"
	"fmt"
	"reflect"
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
	ResourceKey    string
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
				ResourceKey:    d.Key,
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
			ResourceKey:    l.Key,
			ResourceRef:    resourceRef(l),
			FieldPath:      "$resource",
			DesiredValue:   "missing",
			LiveValue:      "present",
			DifferenceType: "extra",
		})
	}

	sort.Slice(findings, func(i, j int) bool {
		left := findings[i]
		right := findings[j]
		if left.ResourceRef != right.ResourceRef {
			return left.ResourceRef < right.ResourceRef
		}
		if left.FieldPath != right.FieldPath {
			return left.FieldPath < right.FieldPath
		}
		if left.DifferenceType != right.DifferenceType {
			return left.DifferenceType < right.DifferenceType
		}
		if left.DesiredValue != right.DesiredValue {
			return left.DesiredValue < right.DesiredValue
		}
		return left.LiveValue < right.LiveValue
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
			ResourceKey:    desired.Key,
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
			ResourceKey:    desired.Key,
			ResourceRef:    ref,
			FieldPath:      "spec.template.spec.containers[0].image",
			DesiredValue:   desiredImage,
			LiveValue:      liveImage,
			DifferenceType: "modified",
		})
	}

	out = append(out, compareFlatMap(desired, ref, desired.Annotations, live.Annotations, mapComparison{
		title:       "Annotation drift",
		category:    "metadata",
		severity:    "low",
		confidence:  0.73,
		fieldPrefix: "metadata.annotations.",
	})...)
	out = append(out, compareFlatMap(desired, ref, desired.Labels, live.Labels, mapComparison{
		title:       "Label drift",
		category:    "metadata",
		severity:    "low",
		confidence:  0.78,
		fieldPrefix: "metadata.labels.",
	})...)

	if strings.TrimSpace(desired.Kind) == "Service" {
		out = append(out, compareFlatMap(desired, ref, readMap(desired.Spec, "selector"), readMap(live.Spec, "selector"), mapComparison{
			title:       "Service selector drift",
			category:    "routing",
			severity:    "high",
			confidence:  0.92,
			fieldPrefix: "spec.selector.",
		})...)
	}

	return out
}

type mapComparison struct {
	title       string
	category    string
	severity    string
	confidence  float64
	fieldPrefix string
}

func compareFlatMap(resource normalize.Resource, ref string, desired, live map[string]any, comparison mapComparison) []Finding {
	out := make([]Finding, 0)
	for _, key := range mapKeys(desired, live) {
		desiredValue, desiredExists := desired[key]
		liveValue, liveExists := live[key]
		if desiredExists && liveExists && reflect.DeepEqual(desiredValue, liveValue) {
			continue
		}

		out = append(out, Finding{
			Title:          comparison.title,
			Category:       comparison.category,
			Severity:       comparison.severity,
			Confidence:     comparison.confidence,
			ResourceKey:    resource.Key,
			ResourceRef:    ref,
			FieldPath:      comparison.fieldPrefix + key,
			DesiredValue:   displayMapValue(desiredValue, desiredExists),
			LiveValue:      displayMapValue(liveValue, liveExists),
			DifferenceType: mapDifferenceType(desiredExists, liveExists),
		})
	}
	return out
}

func displayMapValue(value any, exists bool) string {
	if !exists {
		return "<absent>"
	}
	return fmt.Sprintf("%v", value)
}

func mapDifferenceType(desiredExists, liveExists bool) string {
	switch {
	case !desiredExists:
		return "added"
	case !liveExists:
		return "removed"
	default:
		return "modified"
	}
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

func mapKeys(desired, live map[string]any) []string {
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
