package storage

import (
	"context"
	"fmt"
)

// SeedResult returns key IDs for local demos.
type SeedResult struct {
	WorkspaceID      string
	ClusterID        string
	SourceID         string
	ApplicationOneID string
	ApplicationTwoID string
	IncidentID       string
}

// SeedBaselineData inserts deterministic sample rows for local development.
func (r *Repository) SeedBaselineData(ctx context.Context) (SeedResult, error) {
	result := SeedResult{
		WorkspaceID:      "11111111-1111-1111-1111-111111111111",
		ClusterID:        "22222222-2222-2222-2222-222222222222",
		SourceID:         "33333333-3333-3333-3333-333333333333",
		ApplicationOneID: "44444444-4444-4444-4444-444444444444",
		ApplicationTwoID: "55555555-5555-5555-5555-555555555555",
		IncidentID:       "66666666-6666-6666-6666-666666666666",
	}

	statements := []string{
		fmt.Sprintf(`INSERT INTO workspaces (id, name) VALUES ('%s', 'StateSight Demo Workspace') ON CONFLICT (id) DO NOTHING`, result.WorkspaceID),
		`INSERT INTO users (id, email, display_name) VALUES ('cccccccc-cccc-cccc-cccc-cccccccccccc', 'demo.admin@statesight.local', 'StateSight Demo Admin') ON CONFLICT (id) DO NOTHING`,
		fmt.Sprintf(`INSERT INTO workspace_memberships (workspace_id, user_id, role) VALUES ('%s', 'cccccccc-cccc-cccc-cccc-cccccccccccc', 'admin') ON CONFLICT (workspace_id, user_id) DO NOTHING`, result.WorkspaceID),
		fmt.Sprintf(`INSERT INTO clusters (id, workspace_id, name, provider, kube_context, kubeconfig_path) VALUES ('%s', '%s', 'prod-eu-cluster', 'eks', '', '') ON CONFLICT (id) DO NOTHING`, result.ClusterID, result.WorkspaceID),
		fmt.Sprintf(`INSERT INTO source_definitions (id, workspace_id, name, repo_url, default_branch, path) VALUES ('%s', '%s', 'platform-config', 'https://github.com/example/platform-config', 'main', 'clusters/prod') ON CONFLICT (id) DO NOTHING`, result.SourceID, result.WorkspaceID),
		fmt.Sprintf(`INSERT INTO applications (id, workspace_id, cluster_id, source_definition_id, name, namespace, status) VALUES ('%s', '%s', '%s', '%s', 'ledger-api', 'payments', 'active') ON CONFLICT (id) DO NOTHING`, result.ApplicationOneID, result.WorkspaceID, result.ClusterID, result.SourceID),
		fmt.Sprintf(`INSERT INTO applications (id, workspace_id, cluster_id, source_definition_id, name, namespace, status) VALUES ('%s', '%s', '%s', '%s', 'risk-engine', 'risk', 'active') ON CONFLICT (id) DO NOTHING`, result.ApplicationTwoID, result.WorkspaceID, result.ClusterID, result.SourceID),
		fmt.Sprintf(`INSERT INTO desired_snapshots (id, application_id, revision, summary_json) VALUES ('77777777-7777-7777-7777-777777777777', '%s', 'f2a11cd', '{"resources":12,"source":"git"}') ON CONFLICT (id) DO NOTHING`, result.ApplicationOneID),
		fmt.Sprintf(`INSERT INTO live_snapshots (id, application_id, summary_json) VALUES ('88888888-8888-8888-8888-888888888888', '%s', '{"resources":12,"source":"cluster"}') ON CONFLICT (id) DO NOTHING`, result.ApplicationOneID),
		fmt.Sprintf(`INSERT INTO drift_incidents (id, application_id, desired_snapshot_id, live_snapshot_id, title, category, severity, confidence, recommended_action, status) VALUES ('%s', '%s', '77777777-7777-7777-7777-777777777777', '88888888-8888-8888-8888-888888888888', 'Replica drift detected', 'workload', 'medium', 0.86, 'investigate', 'open') ON CONFLICT (id) DO NOTHING`, result.IncidentID, result.ApplicationOneID),
		fmt.Sprintf(`INSERT INTO drift_fields (id, incident_id, resource_ref, field_path, desired_value, live_value, difference_type) VALUES ('99999999-9999-9999-9999-999999999999', '%s', 'apps/v1/Deployment:payments/ledger-api', 'spec.replicas', '3', '2', 'modified') ON CONFLICT (id) DO NOTHING`, result.IncidentID),
		fmt.Sprintf(`INSERT INTO evidence_records (id, incident_id, source, detail, actor, confidence, metadata) VALUES ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '%s', 'managedFields', 'last-applied-manager=kubectl-rollout', 'system:serviceaccount:payments:deployer', 0.78, '{"fieldManager":"kubectl-rollout"}') ON CONFLICT (id) DO NOTHING`, result.IncidentID),
		fmt.Sprintf(`INSERT INTO evidence_records (id, incident_id, source, detail, actor, confidence, metadata) VALUES ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '%s', 'event', 'manual scaling observed after deployment', 'alice@example.com', 0.71, '{"reason":"hotfix"}') ON CONFLICT (id) DO NOTHING`, result.IncidentID),
	}

	for _, stmt := range statements {
		if _, err := r.pool.Exec(ctx, stmt); err != nil {
			return SeedResult{}, fmt.Errorf("seed statement failed: %w", err)
		}
	}

	return result, nil
}
