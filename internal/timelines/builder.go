package timelines

import (
	"sort"

	"github.com/AnouarMohamed/StateSight/pkg/model"
)

// Builder assembles and normalizes timeline events.
type Builder interface {
	Build(events []model.TimelineEvent) []model.TimelineEvent
}

type DefaultBuilder struct{}

func (DefaultBuilder) Build(events []model.TimelineEvent) []model.TimelineEvent {
	sorted := make([]model.TimelineEvent, 0, len(events))
	sorted = append(sorted, events...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].At.Before(sorted[j].At)
	})

	out := make([]model.TimelineEvent, 0, len(sorted))
	var prev *model.TimelineEvent
	for i := range sorted {
		event := sorted[i]
		if prev != nil && prev.At.Equal(event.At) && prev.Type == event.Type && prev.Summary == event.Summary {
			continue
		}
		out = append(out, event)
		prev = &out[len(out)-1]
	}
	return out
}
