package storage

import (
	"context"
	"fmt"

	"github.com/AnouarMohamed/StateSight/pkg/model"
)

func (r *Repository) GetOverview(ctx context.Context) (model.Overview, error) {
	const query = `
		SELECT
			(SELECT COUNT(*) FROM workspaces) AS workspace_count,
			(SELECT COUNT(*) FROM applications) AS application_count,
			(SELECT COUNT(*) FROM drift_incidents) AS incident_count,
			(SELECT COUNT(*) FROM jobs WHERE status IN ('queued', 'processing')) AS open_jobs_count
	`
	var overview model.Overview
	err := r.pool.QueryRow(ctx, query).Scan(
		&overview.WorkspaceCount,
		&overview.ApplicationCount,
		&overview.IncidentCount,
		&overview.OpenJobsCount,
	)
	if err != nil {
		return model.Overview{}, fmt.Errorf("get overview: %w", err)
	}
	return overview, nil
}

func (r *Repository) GetOverviewByWorkspace(ctx context.Context, workspaceID string) (model.Overview, error) {
	const query = `
		SELECT
			1 AS workspace_count,
			(SELECT COUNT(*) FROM applications WHERE workspace_id = $1) AS application_count,
			(
				SELECT COUNT(*)
				FROM drift_incidents di
				JOIN applications a ON a.id = di.application_id
				WHERE a.workspace_id = $1
			) AS incident_count,
			(
				SELECT COUNT(*)
				FROM jobs j
				JOIN applications a ON a.id = j.application_id
				WHERE a.workspace_id = $1 AND j.status IN ('queued', 'processing')
			) AS open_jobs_count
	`
	var overview model.Overview
	err := r.pool.QueryRow(ctx, query, workspaceID).Scan(
		&overview.WorkspaceCount,
		&overview.ApplicationCount,
		&overview.IncidentCount,
		&overview.OpenJobsCount,
	)
	if err != nil {
		return model.Overview{}, fmt.Errorf("get workspace overview: %w", err)
	}
	return overview, nil
}
