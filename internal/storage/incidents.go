package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/AnouarMohamed/StateSight/internal/timelines"
	"github.com/AnouarMohamed/StateSight/pkg/model"
)

func (r *Repository) CreateIncident(ctx context.Context, params CreateIncidentParams) (model.DriftIncident, error) {
	const query = `
		INSERT INTO drift_incidents (
			id, application_id, desired_snapshot_id, live_snapshot_id,
			title, category, severity, confidence, recommended_action, status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, application_id, desired_snapshot_id, live_snapshot_id, title, category, severity, confidence, recommended_action, status, created_at, updated_at
	`
	id := uuid.NewString()
	var incident model.DriftIncident
	err := r.pool.QueryRow(
		ctx,
		query,
		id,
		params.ApplicationID,
		params.DesiredSnapshotID,
		params.LiveSnapshotID,
		params.Title,
		params.Category,
		params.Severity,
		params.Confidence,
		params.RecommendedAction,
		params.Status,
	).Scan(
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
	if err != nil {
		return model.DriftIncident{}, fmt.Errorf("create incident: %w", err)
	}
	return incident, nil
}

func (r *Repository) CreateDriftField(ctx context.Context, params CreateDriftFieldParams) (model.DriftField, error) {
	const query = `
		INSERT INTO drift_fields (id, incident_id, resource_ref, field_path, desired_value, live_value, difference_type)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, incident_id, resource_ref, field_path, desired_value, live_value, difference_type, created_at
	`
	id := uuid.NewString()
	var field model.DriftField
	err := r.pool.QueryRow(
		ctx,
		query,
		id,
		params.IncidentID,
		params.ResourceRef,
		params.FieldPath,
		params.DesiredValue,
		params.LiveValue,
		params.DifferenceType,
	).Scan(
		&field.ID,
		&field.IncidentID,
		&field.ResourceRef,
		&field.FieldPath,
		&field.DesiredValue,
		&field.LiveValue,
		&field.DifferenceType,
		&field.CreatedAt,
	)
	if err != nil {
		return model.DriftField{}, fmt.Errorf("create drift field: %w", err)
	}
	return field, nil
}

func (r *Repository) CreateEvidenceRecord(ctx context.Context, params CreateEvidenceRecordParams) (model.EvidenceRecord, error) {
	const query = `
		INSERT INTO evidence_records (id, incident_id, source, detail, actor, confidence, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb)
		RETURNING id, incident_id, source, detail, actor, confidence, metadata::text, created_at
	`
	id := uuid.NewString()
	var record model.EvidenceRecord
	err := r.pool.QueryRow(
		ctx,
		query,
		id,
		params.IncidentID,
		params.Source,
		params.Detail,
		params.Actor,
		params.Confidence,
		params.Metadata,
	).Scan(
		&record.ID,
		&record.IncidentID,
		&record.Source,
		&record.Detail,
		&record.Actor,
		&record.Confidence,
		&record.Metadata,
		&record.CreatedAt,
	)
	if err != nil {
		return model.EvidenceRecord{}, fmt.Errorf("create evidence record: %w", err)
	}
	return record, nil
}

func (r *Repository) GetIncidentDetails(ctx context.Context, id string) (model.IncidentDetails, error) {
	const incidentQuery = `
		SELECT id, application_id, desired_snapshot_id, live_snapshot_id, title, category, severity, confidence, recommended_action, status, created_at, updated_at
		FROM drift_incidents
		WHERE id = $1
	`

	var details model.IncidentDetails
	err := r.pool.QueryRow(ctx, incidentQuery, id).Scan(
		&details.Incident.ID,
		&details.Incident.ApplicationID,
		&details.Incident.DesiredSnapshotID,
		&details.Incident.LiveSnapshotID,
		&details.Incident.Title,
		&details.Incident.Category,
		&details.Incident.Severity,
		&details.Incident.Confidence,
		&details.Incident.RecommendedAction,
		&details.Incident.Status,
		&details.Incident.CreatedAt,
		&details.Incident.UpdatedAt,
	)
	if err != nil {
		return model.IncidentDetails{}, mapNotFound(fmt.Errorf("get incident: %w", err))
	}

	fields, err := r.listDriftFieldsByIncident(ctx, id)
	if err != nil {
		return model.IncidentDetails{}, err
	}
	details.Fields = fields

	evidence, err := r.listEvidenceByIncident(ctx, id)
	if err != nil {
		return model.IncidentDetails{}, err
	}
	details.Evidence = evidence

	timeline, err := r.listTimelineByIncident(ctx, id)
	if err != nil {
		return model.IncidentDetails{}, err
	}
	details.Timeline = timeline

	return details, nil
}

func (r *Repository) GetIncidentTimeline(ctx context.Context, incidentID string) ([]model.TimelineEvent, error) {
	var exists bool
	if err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM drift_incidents WHERE id = $1)`, incidentID).Scan(&exists); err != nil {
		return nil, fmt.Errorf("check incident existence: %w", err)
	}
	if !exists {
		return nil, ErrNotFound
	}

	events, err := r.listTimelineByIncident(ctx, incidentID)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (r *Repository) listDriftFieldsByIncident(ctx context.Context, incidentID string) ([]model.DriftField, error) {
	const query = `
		SELECT id, incident_id, resource_ref, field_path, desired_value, live_value, difference_type, created_at
		FROM drift_fields
		WHERE incident_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, incidentID)
	if err != nil {
		return nil, fmt.Errorf("query drift fields: %w", err)
	}
	defer rows.Close()

	fields, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.DriftField, error) {
		var field model.DriftField
		err := row.Scan(
			&field.ID,
			&field.IncidentID,
			&field.ResourceRef,
			&field.FieldPath,
			&field.DesiredValue,
			&field.LiveValue,
			&field.DifferenceType,
			&field.CreatedAt,
		)
		return field, err
	})
	if err != nil {
		return nil, fmt.Errorf("collect drift fields: %w", err)
	}
	return fields, nil
}

func (r *Repository) listEvidenceByIncident(ctx context.Context, incidentID string) ([]model.EvidenceRecord, error) {
	const query = `
		SELECT id, incident_id, source, detail, actor, confidence, metadata::text, created_at
		FROM evidence_records
		WHERE incident_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, incidentID)
	if err != nil {
		return nil, fmt.Errorf("query evidence records: %w", err)
	}
	defer rows.Close()

	records, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.EvidenceRecord, error) {
		var record model.EvidenceRecord
		err := row.Scan(
			&record.ID,
			&record.IncidentID,
			&record.Source,
			&record.Detail,
			&record.Actor,
			&record.Confidence,
			&record.Metadata,
			&record.CreatedAt,
		)
		return record, err
	})
	if err != nil {
		return nil, fmt.Errorf("collect evidence records: %w", err)
	}
	return records, nil
}

func (r *Repository) listTimelineByIncident(ctx context.Context, incidentID string) ([]model.TimelineEvent, error) {
	const query = `
		WITH incident_base AS (
			SELECT id, application_id, desired_snapshot_id, live_snapshot_id, title, created_at
			FROM drift_incidents
			WHERE id = $1
		),
		incident_events AS (
			SELECT ib.created_at AS at, 'incident_opened'::text AS type, ib.title::text AS summary
			FROM incident_base ib
		),
		desired_events AS (
			SELECT ds.captured_at AS at, 'desired_snapshot'::text AS type,
			       ('Desired snapshot captured (revision ' || ds.revision || ')')::text AS summary
			FROM incident_base ib
			JOIN desired_snapshots ds ON ds.id = ib.desired_snapshot_id
		),
		live_events AS (
			SELECT ls.observed_at AS at, 'live_snapshot'::text AS type,
			       'Live snapshot collected from cluster'::text AS summary
			FROM incident_base ib
			JOIN live_snapshots ls ON ls.id = ib.live_snapshot_id
		),
		job_events AS (
			SELECT j.created_at AS at, 'analysis_job'::text AS type,
			       ('Analyze job ' || j.status)::text AS summary
			FROM incident_base ib
			JOIN jobs j ON j.application_id = ib.application_id
			WHERE j.job_type = 'analyze_application'
		),
		github_events_timeline AS (
			SELECT ge.received_at AS at, 'github_event'::text AS type,
			       ('GitHub event ' || ge.event_type || COALESCE(' / ' || ge.action, ''))::text AS summary
			FROM incident_base ib
			JOIN github_events ge ON ge.received_at BETWEEN ib.created_at - INTERVAL '48 hours' AND ib.created_at + INTERVAL '48 hours'
		)
		SELECT at, type, summary FROM incident_events
		UNION ALL
		SELECT at, type, summary FROM desired_events
		UNION ALL
		SELECT at, type, summary FROM live_events
		UNION ALL
		SELECT at, type, summary FROM job_events
		UNION ALL
		SELECT at, type, summary FROM github_events_timeline
		ORDER BY at ASC
		LIMIT 100
	`

	rows, err := r.pool.Query(ctx, query, incidentID)
	if err != nil {
		return nil, fmt.Errorf("query incident timeline: %w", err)
	}
	defer rows.Close()

	events, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.TimelineEvent, error) {
		var event model.TimelineEvent
		err := row.Scan(&event.At, &event.Type, &event.Summary)
		return event, err
	})
	if err != nil {
		return nil, fmt.Errorf("collect incident timeline: %w", err)
	}

	return timelines.DefaultBuilder{}.Build(events), nil
}
