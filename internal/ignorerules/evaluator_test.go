package ignorerules

import (
	"testing"

	"github.com/AnouarMohamed/StateSight/pkg/model"
)

func TestExactFieldPathEvaluatorFindMatch(t *testing.T) {
	evaluator := ExactFieldPathEvaluator{}
	rules := []model.IgnoreRule{
		{ID: "inactive", MatchExpression: "spec.replicas", Active: false},
		{ID: "wildcard-looking", MatchExpression: "spec.*", Active: true},
		{ID: "replicas", MatchExpression: " spec.replicas ", Active: true},
		{ID: "later-replicas", MatchExpression: "spec.replicas", Active: true},
	}

	rule, matched := evaluator.FindMatch(rules, "spec.replicas")
	if !matched {
		t.Fatal("expected exact field path match")
	}
	if rule.ID != "replicas" {
		t.Fatalf("expected first active exact match, got %q", rule.ID)
	}
}

func TestExactFieldPathEvaluatorRejectsPartialMatch(t *testing.T) {
	evaluator := ExactFieldPathEvaluator{}
	rules := []model.IgnoreRule{{ID: "annotation", MatchExpression: "metadata.annotations", Active: true}}

	_, matched := evaluator.FindMatch(rules, "metadata.annotations.rollout")
	if matched {
		t.Fatal("expected partial expression not to suppress a field")
	}
}
