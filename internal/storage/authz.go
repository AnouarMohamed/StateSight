package storage

import (
	"context"
	"fmt"
)

func (r *Repository) GetUserIDByIdentity(ctx context.Context, issuer, subject string) (string, error) {
	const query = `
		SELECT user_id::text
		FROM user_identities
		WHERE issuer = $1 AND subject = $2
	`
	var userID string
	if err := r.pool.QueryRow(ctx, query, issuer, subject).Scan(&userID); err != nil {
		return "", mapNotFound(fmt.Errorf("get user by identity: %w", err))
	}
	return userID, nil
}

func (r *Repository) GetWorkspaceRole(ctx context.Context, userID, workspaceID string) (string, error) {
	const query = `
		SELECT role
		FROM workspace_memberships
		WHERE workspace_id = $1 AND user_id = $2
	`
	var role string
	if err := r.pool.QueryRow(ctx, query, workspaceID, userID).Scan(&role); err != nil {
		return "", mapNotFound(fmt.Errorf("get workspace role: %w", err))
	}
	return role, nil
}
