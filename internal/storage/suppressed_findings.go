package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/AnouarMohamed/StateSight/pkg/model"
)

func (r *Repository) CreateSuppressedFinding(ctx context.Context, params CreateSuppressedFindingParams) (model.SuppressedFinding, error) {
	const query = `
		INSERT INTO suppressed_findings (
			id, application_id, desired_snapshot_id, live_snapshot_id, ignore_rule_id,
			ignore_rule_name, ignore_rule_reason, title, category, severity,
			resource_ref, field_path, desired_value, live_value, difference_type
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, application_id, desired_snapshot_id, live_snapshot_id,
			COALESCE(ignore_rule_id::text, ''), ignore_rule_name, ignore_rule_reason,
			title, category, severity, resource_ref, field_path, desired_value,
			live_value, difference_type, suppressed_at
	`

	var record model.SuppressedFinding
	err := r.pool.QueryRow(
		ctx,
		query,
		uuid.NewString(),
		params.ApplicationID,
		params.DesiredSnapshotID,
		params.LiveSnapshotID,
		params.IgnoreRuleID,
		params.IgnoreRuleName,
		params.IgnoreRuleReason,
		params.Title,
		params.Category,
		params.Severity,
		params.ResourceRef,
		params.FieldPath,
		params.DesiredValue,
		params.LiveValue,
		params.DifferenceType,
	).Scan(
		&record.ID,
		&record.ApplicationID,
		&record.DesiredSnapshotID,
		&record.LiveSnapshotID,
		&record.IgnoreRuleID,
		&record.IgnoreRuleName,
		&record.IgnoreRuleReason,
		&record.Title,
		&record.Category,
		&record.Severity,
		&record.ResourceRef,
		&record.FieldPath,
		&record.DesiredValue,
		&record.LiveValue,
		&record.DifferenceType,
		&record.SuppressedAt,
	)
	if err != nil {
		return model.SuppressedFinding{}, fmt.Errorf("create suppressed finding: %w", err)
	}
	return record, nil
}

func (r *Repository) ListSuppressedFindingsByApplication(ctx context.Context, applicationID string) ([]model.SuppressedFinding, error) {
	const query = `
		SELECT id, application_id, desired_snapshot_id, live_snapshot_id,
			COALESCE(ignore_rule_id::text, ''), ignore_rule_name, ignore_rule_reason,
			title, category, severity, resource_ref, field_path, desired_value,
			live_value, difference_type, suppressed_at
		FROM suppressed_findings
		WHERE application_id = $1
		ORDER BY suppressed_at DESC, id DESC
	`

	rows, err := r.pool.Query(ctx, query, applicationID)
	if err != nil {
		return nil, fmt.Errorf("query suppressed findings by application: %w", err)
	}
	defer rows.Close()

	records, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.SuppressedFinding, error) {
		var record model.SuppressedFinding
		err := row.Scan(
			&record.ID,
			&record.ApplicationID,
			&record.DesiredSnapshotID,
			&record.LiveSnapshotID,
			&record.IgnoreRuleID,
			&record.IgnoreRuleName,
			&record.IgnoreRuleReason,
			&record.Title,
			&record.Category,
			&record.Severity,
			&record.ResourceRef,
			&record.FieldPath,
			&record.DesiredValue,
			&record.LiveValue,
			&record.DifferenceType,
			&record.SuppressedAt,
		)
		return record, err
	})
	if err != nil {
		return nil, fmt.Errorf("collect suppressed findings by application: %w", err)
	}
	return records, nil
}
