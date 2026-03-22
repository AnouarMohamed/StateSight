package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/AnouarMohamed/StateSight/pkg/model"
)

func (r *Repository) ListApplications(ctx context.Context) ([]model.Application, error) {
	const query = `
		SELECT id, workspace_id, cluster_id, source_definition_id, name, namespace, status, created_at, updated_at
		FROM applications
		ORDER BY name ASC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list applications: %w", err)
	}
	defer rows.Close()

	apps, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.Application, error) {
		var app model.Application
		err := row.Scan(
			&app.ID,
			&app.WorkspaceID,
			&app.ClusterID,
			&app.SourceDefinitionID,
			&app.Name,
			&app.Namespace,
			&app.Status,
			&app.CreatedAt,
			&app.UpdatedAt,
		)
		return app, err
	})
	if err != nil {
		return nil, fmt.Errorf("scan applications: %w", err)
	}
	return apps, nil
}

func (r *Repository) ListApplicationsByWorkspace(ctx context.Context, workspaceID string) ([]model.Application, error) {
	const query = `
		SELECT id, workspace_id, cluster_id, source_definition_id, name, namespace, status, created_at, updated_at
		FROM applications
		WHERE workspace_id = $1
		ORDER BY name ASC
	`
	rows, err := r.pool.Query(ctx, query, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list applications by workspace: %w", err)
	}
	defer rows.Close()

	apps, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.Application, error) {
		var app model.Application
		err := row.Scan(
			&app.ID,
			&app.WorkspaceID,
			&app.ClusterID,
			&app.SourceDefinitionID,
			&app.Name,
			&app.Namespace,
			&app.Status,
			&app.CreatedAt,
			&app.UpdatedAt,
		)
		return app, err
	})
	if err != nil {
		return nil, fmt.Errorf("scan applications by workspace: %w", err)
	}
	return apps, nil
}

func (r *Repository) CreateApplication(ctx context.Context, params CreateApplicationParams) (model.Application, error) {
	const query = `
		INSERT INTO applications (id, workspace_id, cluster_id, source_definition_id, name, namespace, status)
		VALUES ($1, $2, $3, $4, $5, $6, 'active')
		RETURNING id, workspace_id, cluster_id, source_definition_id, name, namespace, status, created_at, updated_at
	`
	id := uuid.NewString()
	var app model.Application
	err := r.pool.QueryRow(
		ctx,
		query,
		id,
		params.WorkspaceID,
		params.ClusterID,
		params.SourceDefinitionID,
		params.Name,
		params.Namespace,
	).Scan(
		&app.ID,
		&app.WorkspaceID,
		&app.ClusterID,
		&app.SourceDefinitionID,
		&app.Name,
		&app.Namespace,
		&app.Status,
		&app.CreatedAt,
		&app.UpdatedAt,
	)
	if err != nil {
		return model.Application{}, fmt.Errorf("create application: %w", err)
	}
	return app, nil
}

func (r *Repository) GetApplicationByID(ctx context.Context, id string) (model.Application, error) {
	const query = `
		SELECT id, workspace_id, cluster_id, source_definition_id, name, namespace, status, created_at, updated_at
		FROM applications
		WHERE id = $1
	`
	var app model.Application
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&app.ID,
		&app.WorkspaceID,
		&app.ClusterID,
		&app.SourceDefinitionID,
		&app.Name,
		&app.Namespace,
		&app.Status,
		&app.CreatedAt,
		&app.UpdatedAt,
	)
	if err != nil {
		return model.Application{}, mapNotFound(fmt.Errorf("get application by id: %w", err))
	}
	return app, nil
}
