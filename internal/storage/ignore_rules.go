package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/AnouarMohamed/StateSight/pkg/model"
)

func (r *Repository) ListActiveIgnoreRulesByWorkspace(ctx context.Context, workspaceID string) ([]model.IgnoreRule, error) {
	const query = `
		SELECT id, workspace_id, name, match_expression, reason, created_by, active, created_at
		FROM ignore_rules
		WHERE workspace_id = $1 AND active = TRUE
		ORDER BY created_at ASC, id ASC
	`

	rows, err := r.pool.Query(ctx, query, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list active ignore rules by workspace: %w", err)
	}
	defer rows.Close()

	rules, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.IgnoreRule, error) {
		var rule model.IgnoreRule
		err := row.Scan(
			&rule.ID,
			&rule.WorkspaceID,
			&rule.Name,
			&rule.MatchExpression,
			&rule.Reason,
			&rule.CreatedBy,
			&rule.Active,
			&rule.CreatedAt,
		)
		return rule, err
	})
	if err != nil {
		return nil, fmt.Errorf("scan active ignore rules by workspace: %w", err)
	}
	return rules, nil
}
