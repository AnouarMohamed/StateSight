package ignorerules

import (
	"strings"

	"github.com/AnouarMohamed/StateSight/pkg/model"
)

// Evaluator identifies the first persisted rule that suppresses a drift field.
type Evaluator interface {
	FindMatch(rules []model.IgnoreRule, resourceRef, fieldPath string) (model.IgnoreRule, bool)
}

// ExactFieldPathEvaluator matches canonical drift fields and optional exact resources.
type ExactFieldPathEvaluator struct{}

func (ExactFieldPathEvaluator) FindMatch(rules []model.IgnoreRule, resourceRef, fieldPath string) (model.IgnoreRule, bool) {
	resourceRef = strings.TrimSpace(resourceRef)
	fieldPath = strings.TrimSpace(fieldPath)
	if fieldPath == "" {
		return model.IgnoreRule{}, false
	}

	for _, rule := range rules {
		if !rule.Active {
			continue
		}
		ruleResourceRef := strings.TrimSpace(rule.ResourceRef)
		if ruleResourceRef != "" && ruleResourceRef != resourceRef {
			continue
		}
		if strings.TrimSpace(rule.MatchExpression) == fieldPath {
			return rule, true
		}
	}

	return model.IgnoreRule{}, false
}
