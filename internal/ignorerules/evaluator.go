package ignorerules

import (
	"strings"

	"github.com/AnouarMohamed/StateSight/pkg/model"
)

// Evaluator identifies the first persisted rule that suppresses a drift field.
type Evaluator interface {
	FindMatch(rules []model.IgnoreRule, fieldPath string) (model.IgnoreRule, bool)
}

// ExactFieldPathEvaluator matches a rule expression to one canonical drift field path.
type ExactFieldPathEvaluator struct{}

func (ExactFieldPathEvaluator) FindMatch(rules []model.IgnoreRule, fieldPath string) (model.IgnoreRule, bool) {
	fieldPath = strings.TrimSpace(fieldPath)
	if fieldPath == "" {
		return model.IgnoreRule{}, false
	}

	for _, rule := range rules {
		if !rule.Active {
			continue
		}
		if strings.TrimSpace(rule.MatchExpression) == fieldPath {
			return rule, true
		}
	}

	return model.IgnoreRule{}, false
}
