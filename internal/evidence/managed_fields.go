package evidence

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/AnouarMohamed/StateSight/internal/normalize"
)

type managedFieldsEntry struct {
	Manager    string
	Operation  string
	APIVersion string
	Timestamp  string
	Fields     map[string]any
	Index      int
}

func readManagedFields(raw map[string]any) []managedFieldsEntry {
	metadata, ok := raw["metadata"].(map[string]any)
	if !ok {
		return nil
	}
	rawEntries, ok := metadata["managedFields"].([]any)
	if !ok {
		return nil
	}

	entries := make([]managedFieldsEntry, 0, len(rawEntries))
	for _, rawEntry := range rawEntries {
		entry, ok := rawEntry.(map[string]any)
		if !ok {
			continue
		}
		fields, ok := entry["fieldsV1"].(map[string]any)
		if !ok {
			continue
		}
		entries = append(entries, managedFieldsEntry{
			Manager:    stringValue(entry["manager"]),
			Operation:  stringValue(entry["operation"]),
			APIVersion: stringValue(entry["apiVersion"]),
			Timestamp:  stringValue(entry["time"]),
			Fields:     fields,
		})
	}
	return entries
}

// managedFieldsPath maps only field paths the semantic diff engine emits today.
func managedFieldsPath(resource normalize.Resource, fieldPath string) ([]string, bool) {
	switch {
	case fieldPath == "spec.template.spec.containers[0].image":
		name, ok := firstContainerName(resource)
		if !ok {
			return nil, false
		}
		containerSegment, ok := managedListItemSegment(name)
		if !ok {
			return nil, false
		}
		return []string{"f:spec", "f:template", "f:spec", "f:containers", containerSegment, "f:image"}, true
	case strings.HasPrefix(fieldPath, "spec.template.spec.containers[name="):
		return namedContainerManagedFieldsPath(fieldPath)
	case strings.HasPrefix(fieldPath, "metadata.annotations."):
		annotation := strings.TrimPrefix(fieldPath, "metadata.annotations.")
		if annotation == "" {
			return nil, false
		}
		return []string{"f:metadata", "f:annotations", "f:" + annotation}, true
	case strings.HasPrefix(fieldPath, "metadata.labels."):
		label := strings.TrimPrefix(fieldPath, "metadata.labels.")
		if label == "" {
			return nil, false
		}
		return []string{"f:metadata", "f:labels", "f:" + label}, true
	case strings.HasPrefix(fieldPath, "spec.selector."):
		selector := strings.TrimPrefix(fieldPath, "spec.selector.")
		if selector == "" {
			return nil, false
		}
		return []string{"f:spec", "f:selector", "f:" + selector}, true
	case strings.HasPrefix(fieldPath, "spec.") && !strings.ContainsAny(fieldPath, "[]"):
		segments := strings.Split(fieldPath, ".")
		path := make([]string, 0, len(segments))
		for _, segment := range segments {
			if segment == "" {
				return nil, false
			}
			path = append(path, "f:"+segment)
		}
		return path, true
	default:
		return nil, false
	}
}

func namedContainerManagedFieldsPath(fieldPath string) ([]string, bool) {
	const containerPrefix = "spec.template.spec.containers[name="

	containerName, suffix, ok := consumeNamedPath(fieldPath, containerPrefix)
	// Presence and aggregate environment findings do not identify one owned leaf.
	if !ok || suffix == "" || strings.HasPrefix(suffix, ".env[name=") {
		return nil, false
	}
	containerSegment, ok := managedListItemSegment(containerName)
	if !ok {
		return nil, false
	}
	path := []string{"f:spec", "f:template", "f:spec", "f:containers", containerSegment}

	switch {
	case strings.HasPrefix(suffix, ".resources.requests."):
		resource := strings.TrimPrefix(suffix, ".resources.requests.")
		if resource == "" {
			return nil, false
		}
		return append(path, "f:resources", "f:requests", "f:"+resource), true
	case strings.HasPrefix(suffix, ".resources.limits."):
		resource := strings.TrimPrefix(suffix, ".resources.limits.")
		if resource == "" {
			return nil, false
		}
		return append(path, "f:resources", "f:limits", "f:"+resource), true
	default:
		return nil, false
	}
}

func consumeNamedPath(path, prefix string) (string, string, bool) {
	if !strings.HasPrefix(path, prefix) {
		return "", "", false
	}
	remainder := strings.TrimPrefix(path, prefix)
	close := strings.IndexByte(remainder, ']')
	if close <= 0 {
		return "", "", false
	}
	return remainder[:close], remainder[close+1:], true
}

func managedListItemSegment(name string) (string, bool) {
	key, err := json.Marshal(map[string]string{"name": name})
	if err != nil {
		return "", false
	}
	return "k:" + string(key), true
}

func firstContainerName(resource normalize.Resource) (string, bool) {
	template, ok := resource.Spec["template"].(map[string]any)
	if !ok {
		return "", false
	}
	spec, ok := template["spec"].(map[string]any)
	if !ok {
		return "", false
	}
	containers, ok := spec["containers"].([]any)
	if !ok || len(containers) == 0 {
		return "", false
	}
	container, ok := containers[0].(map[string]any)
	if !ok {
		return "", false
	}
	name := stringValue(container["name"])
	return name, name != ""
}

func ownsManagedPath(fields map[string]any, path []string) bool {
	current := fields
	for index, segment := range path {
		value, ok := current[segment]
		if !ok {
			return false
		}
		if index == len(path)-1 {
			return true
		}
		current, ok = value.(map[string]any)
		if !ok {
			return false
		}
	}
	return false
}

func (entry managedFieldsEntry) newerThan(other managedFieldsEntry) bool {
	entryTime, entryErr := time.Parse(time.RFC3339, entry.Timestamp)
	otherTime, otherErr := time.Parse(time.RFC3339, other.Timestamp)
	entryOK := entryErr == nil
	otherOK := otherErr == nil
	if entryOK && otherOK {
		return entryTime.After(otherTime)
	}
	if entryOK != otherOK {
		return entryOK
	}
	return entry.Index > other.Index
}

func stringValue(value any) string {
	typed, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(typed)
}
