package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/AnouarMohamed/StateSight/pkg/model"
)

func (r *Repository) ListActiveIgnoreRulesForAnalysis(ctx context.Context, workspaceID, applicationID string) ([]model.IgnoreRule, error) {
	const query = `
		SELECT id, workspace_id, COALESCE(application_id::text, ''), COALESCE(resource_ref, ''),
			name, match_expression, reason, created_by, active, created_at
		FROM ignore_rules
		WHERE workspace_id = $1
			AND (application_id IS NULL OR application_id = $2)
			AND active = TRUE
		ORDER BY (application_id IS NOT NULL) DESC, (resource_ref IS NOT NULL) DESC, created_at ASC, id ASC
	`

	rules, err := r.collectIgnoreRules(ctx, query, workspaceID, applicationID)
	if err != nil {
		return nil, fmt.Errorf("list active ignore rules for analysis: %w", err)
	}
	return rules, nil
}

func (r *Repository) ListIgnoreRulesForApplication(ctx context.Context, workspaceID, applicationID string) ([]model.IgnoreRule, error) {
	const query = `
		SELECT id, workspace_id, COALESCE(application_id::text, ''), COALESCE(resource_ref, ''),
			name, match_expression, reason, created_by, active, created_at
		FROM ignore_rules
		WHERE workspace_id = $1
			AND (application_id IS NULL OR application_id = $2)
		ORDER BY (application_id IS NOT NULL) DESC, (resource_ref IS NOT NULL) DESC, active DESC, created_at ASC, id ASC
	`

	rules, err := r.collectIgnoreRules(ctx, query, workspaceID, applicationID)
	if err != nil {
		return nil, fmt.Errorf("list ignore rules for application: %w", err)
	}
	return rules, nil
}

func (r *Repository) CreateIgnoreRule(ctx context.Context, params CreateIgnoreRuleParams) (model.IgnoreRule, error) {
	const query = `
		INSERT INTO ignore_rules (
			id, workspace_id, application_id, resource_ref, name, match_expression, reason, created_by
		) VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6, $7, $8)
		RETURNING id, workspace_id, application_id::text, COALESCE(resource_ref, ''),
			name, match_expression, reason, created_by, active, created_at
	`

	rule, err := scanIgnoreRule(r.pool.QueryRow(
		ctx,
		query,
		uuid.NewString(),
		params.WorkspaceID,
		params.ApplicationID,
		params.ResourceRef,
		params.Name,
		params.MatchExpression,
		params.Reason,
		params.CreatedBy,
	))
	if err != nil {
		return model.IgnoreRule{}, mapConflict(fmt.Errorf("create ignore rule: %w", err))
	}
	return rule, nil
}

func (r *Repository) SetIgnoreRuleActiveForApplication(ctx context.Context, id, applicationID string, active bool) (model.IgnoreRule, error) {
	const query = `
		UPDATE ignore_rules
		SET active = $3
		WHERE id = $1 AND application_id = $2
		RETURNING id, workspace_id, application_id::text, COALESCE(resource_ref, ''),
			name, match_expression, reason, created_by, active, created_at
	`

	rule, err := scanIgnoreRule(r.pool.QueryRow(ctx, query, id, applicationID, active))
	if err != nil {
		return model.IgnoreRule{}, mapNotFound(fmt.Errorf("set application ignore rule active: %w", err))
	}
	return rule, nil
}

func (r *Repository) collectIgnoreRules(ctx context.Context, query string, args ...any) ([]model.IgnoreRule, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.IgnoreRule, error) {
		return scanIgnoreRule(row)
	})
}

func scanIgnoreRule(row interface{ Scan(...any) error }) (model.IgnoreRule, error) {
	var rule model.IgnoreRule
	err := row.Scan(
		&rule.ID,
		&rule.WorkspaceID,
		&rule.ApplicationID,
		&rule.ResourceRef,
		&rule.Name,
		&rule.MatchExpression,
		&rule.Reason,
		&rule.CreatedBy,
		&rule.Active,
		&rule.CreatedAt,
	)
	return rule, err
}
