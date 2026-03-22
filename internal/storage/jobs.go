package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/AnouarMohamed/StateSight/pkg/model"
)

func (r *Repository) CreateJob(ctx context.Context, params CreateJobParams) (model.Job, error) {
	const query = `
		INSERT INTO jobs (id, job_type, status, application_id, payload)
		VALUES ($1, $2, $3, $4, $5::jsonb)
		RETURNING id, job_type, status, application_id, payload::text, error_message, created_at, started_at, completed_at
	`
	id := uuid.NewString()
	var job model.Job
	err := r.pool.QueryRow(
		ctx,
		query,
		id,
		params.JobType,
		model.JobStatusQueued,
		params.ApplicationID,
		params.PayloadJSON,
	).Scan(
		&job.ID,
		&job.JobType,
		&job.Status,
		&job.ApplicationID,
		&job.Payload,
		&job.ErrorMessage,
		&job.CreatedAt,
		&job.StartedAt,
		&job.CompletedAt,
	)
	if err != nil {
		return model.Job{}, fmt.Errorf("create job: %w", err)
	}
	return job, nil
}

func (r *Repository) MarkJobProcessing(ctx context.Context, id string) error {
	const query = `
		UPDATE jobs
		SET status = $2, started_at = $3, error_message = NULL
		WHERE id = $1
	`
	if _, err := r.pool.Exec(ctx, query, id, model.JobStatusProcessing, time.Now().UTC()); err != nil {
		return fmt.Errorf("mark job processing: %w", err)
	}
	return nil
}

func (r *Repository) MarkJobCompleted(ctx context.Context, id string) error {
	const query = `
		UPDATE jobs
		SET status = $2, completed_at = $3
		WHERE id = $1
	`
	if _, err := r.pool.Exec(ctx, query, id, model.JobStatusCompleted, time.Now().UTC()); err != nil {
		return fmt.Errorf("mark job completed: %w", err)
	}
	return nil
}

func (r *Repository) MarkJobFailed(ctx context.Context, id, message string) error {
	const query = `
		UPDATE jobs
		SET status = $2, error_message = $3, completed_at = $4
		WHERE id = $1
	`
	if _, err := r.pool.Exec(ctx, query, id, model.JobStatusFailed, message, time.Now().UTC()); err != nil {
		return fmt.Errorf("mark job failed: %w", err)
	}
	return nil
}
