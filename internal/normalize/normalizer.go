package normalize

// Snapshot is the normalization boundary for desired and live data.
type Snapshot struct {
	Body map[string]any
}

// Normalizer converts collected state into a canonical comparable form.
type Normalizer interface {
	Normalize(input map[string]any) Snapshot
}

// PassThroughNormalizer keeps baseline simple while preserving extension points.
type PassThroughNormalizer struct{}

func (n PassThroughNormalizer) Normalize(input map[string]any) Snapshot {
	return Snapshot{Body: input}
}
