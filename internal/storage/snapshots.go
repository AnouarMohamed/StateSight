package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/AnouarMohamed/StateSight/pkg/model"
)

func (r *Repository) CreateDesiredSnapshot(ctx context.Context, params CreateDesiredSnapshotParams) (model.DesiredSnapshot, error) {
	return r.CreateDesiredSnapshotWithQuerier(ctx, r.pool, params)
}

func (r *Repository) CreateDesiredSnapshotWithQuerier(ctx context.Context, q Querier, params CreateDesiredSnapshotParams) (model.DesiredSnapshot, error) {
	const query = `
		INSERT INTO desired_snapshots (id, application_id, revision, summary_json)
		VALUES ($1, $2, $3, $4::jsonb)
		RETURNING id, application_id, revision, summary_json::text, captured_at
	`
	id := uuid.NewString()
	var snapshot model.DesiredSnapshot
	err := q.QueryRow(ctx, query, id, params.ApplicationID, params.Revision, params.SummaryJSON).Scan(
		&snapshot.ID,
		&snapshot.ApplicationID,
		&snapshot.Revision,
		&snapshot.SummaryJSON,
		&snapshot.CapturedAt,
	)
	if err != nil {
		return model.DesiredSnapshot{}, fmt.Errorf("create desired snapshot: %w", err)
	}
	return snapshot, nil
}

func (r *Repository) CreateLiveSnapshot(ctx context.Context, params CreateLiveSnapshotParams) (model.LiveSnapshot, error) {
	return r.CreateLiveSnapshotWithQuerier(ctx, r.pool, params)
}

func (r *Repository) CreateLiveSnapshotWithQuerier(ctx context.Context, q Querier, params CreateLiveSnapshotParams) (model.LiveSnapshot, error) {
	const query = `
		INSERT INTO live_snapshots (id, application_id, summary_json)
		VALUES ($1, $2, $3::jsonb)
		RETURNING id, application_id, summary_json::text, observed_at
	`
	id := uuid.NewString()
	var snapshot model.LiveSnapshot
	err := q.QueryRow(ctx, query, id, params.ApplicationID, params.SummaryJSON).Scan(
		&snapshot.ID,
		&snapshot.ApplicationID,
		&snapshot.SummaryJSON,
		&snapshot.ObservedAt,
	)
	if err != nil {
		return model.LiveSnapshot{}, fmt.Errorf("create live snapshot: %w", err)
	}
	return snapshot, nil
}
