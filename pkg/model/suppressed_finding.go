package model

import "time"

// SuppressedFinding is a drift candidate withheld from incidents by an ignore rule.
type SuppressedFinding struct {
	ID                string    `json:"id"`
	ApplicationID     string    `json:"application_id"`
	DesiredSnapshotID string    `json:"desired_snapshot_id"`
	LiveSnapshotID    string    `json:"live_snapshot_id"`
	IgnoreRuleID      string    `json:"ignore_rule_id"`
	IgnoreRuleName    string    `json:"ignore_rule_name"`
	IgnoreRuleReason  string    `json:"ignore_rule_reason"`
	Title             string    `json:"title"`
	Category          string    `json:"category"`
	Severity          string    `json:"severity"`
	ResourceRef       string    `json:"resource_ref"`
	FieldPath         string    `json:"field_path"`
	DesiredValue      string    `json:"desired_value"`
	LiveValue         string    `json:"live_value"`
	DifferenceType    string    `json:"difference_type"`
	SuppressedAt      time.Time `json:"suppressed_at"`
}
