package ignorerules

import "context"

// Evaluator determines if a candidate should be suppressed.
type Evaluator interface {
	ShouldIgnore(ctx context.Context, applicationID, fieldPath string) (bool, string, error)
}

// NoopEvaluator keeps all findings visible until ignore rules are implemented.
type NoopEvaluator struct{}

func (NoopEvaluator) ShouldIgnore(_ context.Context, _ string, _ string) (bool, string, error) {
	return false, "", nil
}
