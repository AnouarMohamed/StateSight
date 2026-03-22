package storage

import (
	"context"
	"fmt"

	"github.com/AnouarMohamed/StateSight/pkg/model"
)

func (r *Repository) GetSourceDefinitionByID(ctx context.Context, id string) (model.SourceDefinition, error) {
	const query = `
		SELECT id, workspace_id, name, repo_url, default_branch, path, created_at
		FROM source_definitions
		WHERE id = $1
	`
	var source model.SourceDefinition
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&source.ID,
		&source.WorkspaceID,
		&source.Name,
		&source.RepoURL,
		&source.DefaultBranch,
		&source.Path,
		&source.CreatedAt,
	)
	if err != nil {
		return model.SourceDefinition{}, mapNotFound(fmt.Errorf("get source definition by id: %w", err))
	}
	return source, nil
}
