package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/AnouarMohamed/StateSight/pkg/model"
)

func (r *Repository) InsertGitHubEvent(ctx context.Context, params UpsertGitHubEventParams) (model.GitHubEvent, error) {
	const query = `
		INSERT INTO github_events (id, event_type, delivery_id, action, repository, payload)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb)
		RETURNING id, event_type, delivery_id, action, repository, payload::text, received_at
	`
	id := uuid.NewString()
	var event model.GitHubEvent
	err := r.pool.QueryRow(
		ctx,
		query,
		id,
		params.EventType,
		params.DeliveryID,
		params.Action,
		params.Repository,
		params.Payload,
	).Scan(
		&event.ID,
		&event.EventType,
		&event.DeliveryID,
		&event.Action,
		&event.Repository,
		&event.Payload,
		&event.ReceivedAt,
	)
	if err != nil {
		return model.GitHubEvent{}, fmt.Errorf("insert github event: %w", err)
	}
	return event, nil
}
