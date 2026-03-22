package storage

import (
	"context"
	"fmt"

	"github.com/AnouarMohamed/StateSight/pkg/model"
)

func (r *Repository) GetClusterByID(ctx context.Context, id string) (model.Cluster, error) {
	const query = `
		SELECT id, workspace_id, name, provider, kube_context, kubeconfig_path, created_at
		FROM clusters
		WHERE id = $1
	`
	var cluster model.Cluster
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&cluster.ID,
		&cluster.WorkspaceID,
		&cluster.Name,
		&cluster.Provider,
		&cluster.KubeContext,
		&cluster.KubeconfigPath,
		&cluster.CreatedAt,
	)
	if err != nil {
		return model.Cluster{}, mapNotFound(fmt.Errorf("get cluster by id: %w", err))
	}
	return cluster, nil
}
