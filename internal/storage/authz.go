package storage

import (
	"context"
	"fmt"
)

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
