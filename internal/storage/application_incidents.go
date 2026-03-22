package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/AnouarMohamed/StateSight/pkg/model"
)

func (r *Repository) ListIncidentsByApplication(ctx context.Context, applicationID string) ([]model.DriftIncident, error) {
	const query = `
		SELECT id, application_id, desired_snapshot_id, live_snapshot_id, title, category, severity, confidence, recommended_action, status, created_at, updated_at
		FROM drift_incidents
		WHERE application_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, applicationID)
	if err != nil {
		return nil, fmt.Errorf("query incidents by application: %w", err)
	}
	defer rows.Close()

	incidents, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.DriftIncident, error) {
		var incident model.DriftIncident
		err := row.Scan(
			&incident.ID,
			&incident.ApplicationID,
			&incident.DesiredSnapshotID,
			&incident.LiveSnapshotID,
			&incident.Title,
			&incident.Category,
			&incident.Severity,
			&incident.Confidence,
			&incident.RecommendedAction,
			&incident.Status,
			&incident.CreatedAt,
			&incident.UpdatedAt,
		)
		return incident, err
	})
	if err != nil {
		return nil, fmt.Errorf("collect incidents by application: %w", err)
	}
	return incidents, nil
}
